// internal/feishu/bot.go
package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/schema"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// FeishuBot 封装了飞书机器人的配置与核心业务流
type FeishuBot struct {
	client    *lark.Client
	appID     string
	appSecret string
	engine    *engine.AgentEngine // 持有核心引擎引用
	sess      *ctxpkg.Session     // 绑定的会话
	r         *FeishuReporter     // Reporter 实例
}

func NewFeishuBot(eng *engine.AgentEngine, sess *ctxpkg.Session) *FeishuBot {
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")

	if appID == "" || appSecret == "" {
		log.Fatal("请设置 FEISHU_APP_ID 和 FEISHU_APP_SECRET")
	}

	client := lark.NewClient(appID, appSecret)

	return &FeishuBot{
		client:    client,
		appID:     appID,
		appSecret: appSecret,
		engine:    eng,
		sess:      sess,
	}
}

// GetEventDispatcher 用于注册到 HTTP 服务器，处理来自飞书的 POST 事件
func (b *FeishuBot) GetEventDispatcher() *dispatcher.EventDispatcher {
	encryptKey := os.Getenv("FEISHU_ENCRYPT_KEY")
	verifyToken := os.Getenv("FEISHU_VERIFY_TOKEN")

	handler := dispatcher.NewEventDispatcher(verifyToken, encryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			contentStr := *event.Event.Message.Content
			contentStr = strings.TrimPrefix(contentStr, `{"text":"`)
			contentStr = strings.TrimSuffix(contentStr, `"}`)

			chatId := *event.Event.Message.ChatId
			log.Printf("[Feishu] 收到会话 %s 消息: %s\n", chatId, contentStr)

			// 【新增】：拦截人工审批的特殊口令
			if strings.HasPrefix(contentStr, "approve ") {
				taskID := strings.TrimPrefix(contentStr, "approve ")
				taskID = strings.TrimSpace(taskID)
				GlobalApprovalMgr.ResolveApproval(taskID, true, "人类管理员已批准操作")
				log.Printf("[Feishu] 会话 %s: ✅ 已为您批准任务 %s", chatId, taskID)
				return nil
			}
			if strings.HasPrefix(contentStr, "reject ") {
				taskID := strings.TrimPrefix(contentStr, "reject ")
				taskID = strings.TrimSpace(taskID)
				GlobalApprovalMgr.ResolveApproval(taskID, false, "人类管理员认为该操作存在极高风险，已无情拒绝")
				log.Printf("[Feishu] 会话 %s: 🚫 已拒绝任务 %s", chatId, taskID)
				return nil
			}

			// 如果不是审批命令，则是正常对话，启动一个新的 Agent 任务去处理
			go b.handleAgentRun(chatId, contentStr)

			return nil
		}).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			return nil
		})

	return handler
}

// Reporter 返回绑定的 FeishuReporter 实例
func (b *FeishuBot) Reporter() *FeishuReporter {
	return b.r
}

// handleAgentRun 是连接飞书与底层引擎的桥梁
func (b *FeishuBot) handleAgentRun(chatId string, prompt string) {
	reporter := &FeishuReporter{
		client: b.client,
		chatId: chatId,
	}
	b.r = reporter

	// 将 prompt 加入 Session
	b.sess.Append(schema.Message{Role: schema.RoleUser, Content: prompt})

	err := b.engine.Run(context.Background(), b.sess, reporter)
	if err != nil {
		reporter.sendMsg(fmt.Sprintf("❌ Agent 运行崩溃: %v", err))
	}
}

// ==========================================
// FeishuReporter: 将引擎的输出格式化后发给飞书
// ==========================================
type FeishuReporter struct {
	client *lark.Client
	chatId string
}

// sendMsg 封装了调用飞书 OpenAPI 发送卡片/文本的操作
func (r *FeishuReporter) sendMsg(text string) {
	textContent := map[string]string{
		"text": text,
	}
	contentBytes, _ := json.Marshal(textContent)
	contentStr := string(contentBytes)

	msgReq := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(r.chatId).
			MsgType(larkim.MsgTypeText).
			Content(contentStr).
			Build()).
		Build()
	_, _ = r.client.Im.Message.Create(context.Background(), msgReq)
}

func (r *FeishuReporter) OnThinking(ctx context.Context) {
	r.sendMsg("🤔 模型正在慢思考 (Thinking)...")
}

func (r *FeishuReporter) OnToolCall(ctx context.Context, toolName string, args string) {
	r.sendMsg(fmt.Sprintf("🛠️ **正在执行工具**：`%s`\n参数：`%s`", toolName, args))
}

func (r *FeishuReporter) OnToolResult(ctx context.Context, toolName string, result string, isError bool) {
	if isError {
		r.sendMsg(fmt.Sprintf("⚠️ **执行报错** (%s)：\n%s", toolName, result))
	} else {
		r.sendMsg(fmt.Sprintf("✅ **执行成功** (%s)", toolName))
	}
}

func (r *FeishuReporter) OnMessage(ctx context.Context, content string) {
	r.sendMsg(content)
}

// 编译时类型检查
var _ engine.Reporter = (*FeishuReporter)(nil)
