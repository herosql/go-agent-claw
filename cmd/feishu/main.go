// cmd/feishu/main.go
package main

import (
	"log/slog"
	"net/http"
	"os"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/feishu"
	"github.com/herosql/go-agent-claw/internal/loginit"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/tools"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
)

func main() {
	loginit.Init()

	// 1. 环境检查
	if os.Getenv("ZHIPU_API_KEY") == "" {
		slog.Error("请先导出 ZHIPU_API_KEY 环境变量")
		os.Exit(1)
	}

	// 2. 初始化工作区
	workDir, _ := os.Getwd()
	workDir += "/workspace"

	// 3. 初始化 LLM Provider
	modelName := "glm-4.5-air"
	llmProvider := provider.NewZhipuOpenAIProvider(modelName)

	// 4. 初始化工具注册表
	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// 5. 初始化会话
	sess := ctxpkg.GlobalSessionMgr.GetOrCreate("feishu_default", workDir)

	// 6. 初始化引擎（开启 EnableThinking 慢思考）
	eng := engine.NewAgentEngine(llmProvider, registry, false, true)

	// 7. 初始化飞书 Bot 调度器
	bot := feishu.NewFeishuBot(eng, sess)
	handler := httpserverext.NewEventHandlerFunc(bot.GetEventDispatcher())

	// 7. 注册路由并启动 HTTP 服务
	http.HandleFunc("/webhook/event", handler)

	port := ":48080"
	slog.Info("==================================================")
	slog.Info("🚀 go-agent-claw 飞书服务端已启动")
	slog.Info("📁 工作区: " + workDir)
	slog.Info("🤖 模型: " + modelName)
	slog.Info("🌐 监听端口: " + port)
	slog.Info("==================================================")

	err := http.ListenAndServe(port, nil)
	if err != nil {
		slog.Error("服务器启动失败", "err", err)
		os.Exit(1)
	}
}