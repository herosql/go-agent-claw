// cmd/fei/main.go
package main

import (
	"context"
	"log/slog"
	"os"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/feishu"
	"github.com/herosql/go-agent-claw/internal/loginit"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/tools"
)

func main() {
	loginit.Init()

	// 1. 环境检查
	if os.Getenv("ZHIPU_API_KEY") == "" {
		slog.Error("请先导出 ZHIPU_API_KEY 环境变量")
		os.Exit(1)
	}

	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")
	if appID == "" || appSecret == "" {
		slog.Error("请设置 FEISHU_APP_ID 和 FEISHU_APP_SECRET")
		os.Exit(1)
	}

	// 2. 初始化工作区
	workDir, err := os.Getwd()
	if err != nil {
		slog.Error("获取工作目录失败", "err", err)
		os.Exit(1)
	}
	workDir += "/workspace"

	// 3. 初始化 LLM Provider
	modelName := "glm-4.5-air"
	realProvider := provider.NewZhipuOpenAIProvider(modelName)

	// 4. 初始化工具注册表
	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// 5. 初始化会话
	sess := ctxpkg.GlobalSessionMgr.GetOrCreate("feishu_default", workDir)

	// 6. 初始化引擎（关闭 Thinking 加速响应）
	eng := engine.NewAgentEngine(realProvider, registry, false, false)

	// 7. 初始化飞书 Bot 调度器
	bot := feishu.NewFeishuBot(eng, sess)

	// 8. 创建 WebSocket 客户端
	wsClient := larkws.NewClient(appID, appSecret,
		larkws.WithEventHandler(bot.GetEventDispatcher()),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
		larkws.WithAutoReconnect(true),
	)

	slog.Info("==================================================")
	slog.Info("🚀 go-agent-claw 飞书服务端已启动")
	slog.Info("📁 工作区: " + workDir)
	slog.Info("🤖 模型: " + modelName)
	slog.Info("==================================================")

	// 8. 启动 WebSocket 长连接
	err = wsClient.Start(context.Background())
	if err != nil {
		slog.Error("服务器启动失败", "err", err)
		os.Exit(1)
	}
}
