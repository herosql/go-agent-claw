// internal/feishu/bot.go
package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/schema"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// ==========================================
// 1. Context 传递机制：解决并发 Reporter 的提取
// ==========================================

// reporterKey 定义 Context 中存放 Reporter 的专属键
type reporterKey struct{}

// ContextWithReporter 将专属的 Reporter 封入上下文
func ContextWithReporter(ctx context.Context, r engine.Reporter) context.Context {
	return context.WithValue(ctx, reporterKey{}, r)
}

// ReporterFromContext 供底层的 Middleware 提取专属的 Reporter 发送审批卡片
func ReporterFromContext(ctx context.Context) engine.Reporter {
	if r, ok := ctx.Value(reporterKey{}).(engine.Reporter); ok {
		return r
	}
	return nil
}

// ==========================================
// 2. 飞书 Bot 核心调度器
// ==========================================

// AgentEngineFactory 允许每次收到消息时，根据 Session 动态创建引擎
type AgentEngineFactory func(session *ctxpkg.Session) *engine.AgentEngine

type FeishuBot struct {
	client    *lark.Client
	appID     string
	appSecret string
	workDir   string             // 保存从入口传来的工作区路径
	factory   AgentEngineFactory // 替换掉原来的单一 engine 引用
}

func NewFeishuBotWithFactory(factory AgentEngineFactory, workDir string) *FeishuBot {
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")

	if appID == "" || appSecret == "" {
		slog.Error("请设置 FEISHU_APP_ID 和 FEISHU_APP_SECRET")
		os.Exit(1)
	}

	client := lark.NewClient(appID, appSecret)

	return &FeishuBot{
		client:    client,
		appID:     appID,
		appSecret: appSecret,
		factory:   factory,
		workDir:   workDir,
	}
}

// NewFeishuBot 简单的构造器，直接接受已组装好的 Engine 和 Session
func NewFeishuBot(eng *engine.AgentEngine, sess *ctxpkg.Session) *FeishuBot {
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")

	if appID == "" || appSecret == "" {
		slog.Error("请设置 FEISHU_APP_ID 和 FEISHU_APP_SECRET")
		os.Exit(1)
	}

	client := lark.NewClient(appID, appSecret)

	// 创建一个简单的工厂函数
	factory := func(session *ctxpkg.Session) *engine.AgentEngine {
		return eng
	}

	return &FeishuBot{
		client:    client,
		appID:     appID,
		appSecret: appSecret,
		workDir:   sess.WorkDir,
		factory:   factory,
	}
}

func (b *FeishuBot) GetEventDispatcher() *dispatcher.EventDispatcher {
	encryptKey := os.Getenv("FEISHU_ENCRYPT_KEY")
	verifyToken := os.Getenv("FEISHU_VERIFY_TOKEN")

	handler := dispatcher.NewEventDispatcher(verifyToken, encryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			contentStr := *event.Event.Message.Content
			contentStr = strings.TrimPrefix(contentStr, `{"text":"`)
			contentStr = strings.TrimSuffix(contentStr, `"}`)

			chatId := *event.Event.Message.ChatId
			slog.Info("[Feishu] 收到会话 " + chatId + " 消息: " + contentStr)

			// 拦截人工审批的特殊口令，并唤醒挂起的 Registry 协程
			if strings.HasPrefix(contentStr, "approve ") {
				taskID := strings.TrimPrefix(contentStr, "approve ")
				taskID = strings.TrimSpace(taskID)
				GlobalApprovalMgr.ResolveApproval(taskID, true, "人类管理员已批准操作")
				slog.Info("[Feishu] 会话 " + chatId + ": ✅ 已为您批准任务 " + taskID)
				return nil
			}
			if strings.HasPrefix(contentStr, "reject ") {
				taskID := strings.TrimPrefix(contentStr, "reject ")
				taskID = strings.TrimSpace(taskID)
				GlobalApprovalMgr.ResolveApproval(taskID, false, "人类管理员认为该操作存在极高风险，已无情拒绝")
				slog.Info("[Feishu] 会话 " + chatId + ": 🚫 已拒绝任务 " + taskID)
				return nil
			}

			// 如果是普通对话，新开一个 Goroutine 去启动 Agent，防止阻塞 Webhook
			go b.handleAgentRun(chatId, contentStr)

			return nil
		}).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			// 消息已读事件，静默忽略
			return nil
		})

	return handler
}

func (b *FeishuBot) handleAgentRun(chatId string, prompt string) {
	// 为当前并发请求实例化一个专属的 Reporter
	reporter := &FeishuReporter{
		client: b.client,
		chatId: chatId,
	}

	// 1. 获取物理隔离的 Session
	sess := ctxpkg.GlobalSessionMgr.GetOrCreate(chatId, b.workDir)
	sess.Append(schema.Message{Role: schema.RoleUser, Content: prompt})

	// 2. 通过工厂模式，为当前会话生成一个挂好了专属 CostTracker 的新引擎
	eng := b.factory(sess)

	// 3. 【驾驭核心】：将专属的 reporter 塞入 Context 并传给引擎！
	runCtx := ContextWithReporter(context.Background(), reporter)

	if err := eng.Run(runCtx, sess, reporter); err != nil {
		reporter.sendMsg(fmt.Sprintf("❌ Agent 运行崩溃: %v", err))
	}
}

// ==========================================
// 3. 飞书 Reporter 实现 ()
// ==========================================

type FeishuReporter struct {
	client *lark.Client
	chatId string
}

func (r *FeishuReporter) sendMsg(text string) {
	textContent := map[string]string{
		"text": text,
	}
	contentBytes, err := json.Marshal(textContent)
	if err != nil {
		return
	}
	contentStr := string(contentBytes)

	msgReq := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(r.chatId).
			MsgType(larkim.MsgTypeText).
			Content(contentStr).
			Build()).
		Build()

	_, err = r.client.Im.Message.Create(context.Background(), msgReq)
	if err != nil {
		// 消息发送失败不影响主流程，仅记录日志
		fmt.Printf("警告: 飞书消息发送失败: %v\n", err)
	}
}

func (r *FeishuReporter) OnThinking(ctx context.Context) {
	r.sendMsg("🤔 模型正在慢思考 (Thinking)...")
}

func (r *FeishuReporter) OnToolCall(ctx context.Context, toolName string, args string) {
	// 将 JSON 参数中的字符串 \n 转换为真实换行，使显示更易读
	args = strings.ReplaceAll(args, "\\n", "\n")
	r.sendMsg(fmt.Sprintf("🛠️ **正在执行工具**：`%s`\n参数：`%s`", toolName, args))
}

func (r *FeishuReporter) OnToolRunning(ctx context.Context, toolName string, elapsedSecs int) {
	r.sendMsg(fmt.Sprintf("⏳ **执行中** (%s) — 已运行 %ds...", toolName, elapsedSecs))
}

func (r *FeishuReporter) OnToolResult(ctx context.Context, toolName string, result string, isError bool) {
	display := result
	if len(display) > 200 {
		display = display[:200] + "... (已截断)"
	}
	if isError {
		r.sendMsg(fmt.Sprintf("⚠️ **执行报错** (%s)：\n%s", toolName, display))
	} else {
		r.sendMsg(fmt.Sprintf("✅ **执行成功** (%s)\n%s", toolName, display))
	}
}

func (r *FeishuReporter) OnMessage(ctx context.Context, content string) {
	r.sendMsg(content)
}

// 确保 FeishuReporter 实现了 Reporter 接口
var _ engine.Reporter = (*FeishuReporter)(nil)
