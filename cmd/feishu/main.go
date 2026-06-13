// cmd/feishu/main.go
package main

import (
	"log"
	"net/http"
	"os"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/feishu"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/tools"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
)

func main() {
	// 1. 环境检查
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先导出 ZHIPU_API_KEY 环境变量")
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
	log.Println("==================================================")
	log.Printf("🚀 go-agent-claw 飞书服务端已启动\n")
	log.Printf("📁 工作区: %s\n", workDir)
	log.Printf("🤖 模型: %s\n", modelName)
	log.Printf("🌐 监听端口: %s\n", port)
	log.Println("==================================================")

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}