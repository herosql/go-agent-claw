// cmd/agentops/main.go
package main

import (
	"context"
	"log/slog"
	"os"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/feishu"
	"github.com/herosql/go-agent-claw/internal/loginit"
	"github.com/herosql/go-agent-claw/internal/observability"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/schema"
	"github.com/herosql/go-agent-claw/internal/tools"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func main() {
	loginit.Init()
	slog.Info("🚀 正在启动 go-agent-claw AgentOps 飞书服务端 (WebSocket)...")

	if os.Getenv("ZHIPU_API_KEY") == "" || os.Getenv("FEISHU_APP_ID") == "" || os.Getenv("FEISHU_APP_SECRET") == "" {
		slog.Error("❌ 请先导出 ZHIPU_API_KEY、FEISHU_APP_ID 和 FEISHU_APP_SECRET")
		os.Exit(1)
	}

	// 1. 设定监控的物理工作区
	workDir, _ := os.Getwd()
	workDir += "/workspace"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		slog.Error("无法创建工作区", "err", err)
		os.Exit(1)
	}

	// 2. 初始化底层大脑与注册表
	modelName := "glm-4.5-air"
	llmProvider := provider.NewZhipuOpenAIProvider(modelName)

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir)) // 必备的运维工具

	// 3. 【核心防御】：注入安全拦截 Middleware
	registry.Use(func(ctx context.Context, call schema.ToolCall) (bool, string) {
		argsStr := string(call.Arguments)

		// 检查是否命中危险命令黑名单
		if feishu.IsDangerousCommand(call.Name, argsStr) {
			taskID := call.ID
			slog.Info("[Middleware] 拦截到高危操作: " + call.Name + "，触发飞书审批挂起...")

			// 【驾驭魔术】：从 Context 中优雅地取出专属于发起该请求群聊的 Reporter！
			currentReporter, _ := feishu.ReporterFromContext(ctx).(*feishu.FeishuReporter)

			// 当前 Goroutine 死死挂起，向飞书发送卡片，等待人类决定
			allowed, reason := feishu.GlobalApprovalMgr.WaitForApproval(taskID, call.Name, argsStr, currentReporter)

			if !allowed {
				return false, reason // 拒绝，将理由作为 ToolResult 喂回给大模型
			}
			return true, "" // 同意，放行底层物理执行
		}

		// 普通读取命令，YOLO 放行
		return true, ""
	})
	slog.Info("🛡️ 安全防御 Middleware 已挂载。")

	// 4. 动态 Factory 组装器：保证高并发调用的物理独立性与账单准确追踪
	engineFactory := func(session *ctxpkg.Session) *engine.AgentEngine {
		// 让 Tracker 绑定当前特定用户的 Session 账本
		trackedProvider := observability.NewCostTracker(llmProvider, modelName, session)

		// 返回一个新组装的 Engine 实例
		return engine.NewAgentEngine(trackedProvider, registry, false, false)
	}

	// 5. 初始化飞书 Bot 调度中心
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")

	bot := feishu.NewFeishuBotWithFactory(engineFactory, workDir)

	// 6. 创建 WebSocket 客户端
	wsClient := larkws.NewClient(appID, appSecret,
		larkws.WithEventHandler(bot.GetEventDispatcher()),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
		larkws.WithAutoReconnect(true),
	)

	slog.Info("==================================================")
	slog.Info("🚀 go-agent-claw AgentOps 飞书服务端已启动")
	slog.Info("📁 工作区: " + workDir)
	slog.Info("🤖 模型: " + modelName)
	slog.Info("==================================================")

	// 7. 启动 WebSocket 长连接
	err := wsClient.Start(context.Background())
	if err != nil {
		slog.Error("服务器启动失败", "err", err)
		os.Exit(1)
	}
}