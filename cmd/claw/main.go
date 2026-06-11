// cmd/claw/main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/feishu"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/tools"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

func main() {
	// 1. 初始化引擎依赖
	workDir, _ := os.Getwd()

	// 默认使用智谱 GLM-4
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先导出 ZHIPU_API_KEY 环境变量")
	}
	llmProvider := provider.NewZhipuOpenAIProvider("glm-4.5-air")

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// 开启慢思考
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, true)

	// 2. 初始化飞书 Bot 调度器
	bot := feishu.NewFeishuBot(eng)

	//  http版
	// handler := httpserverext.NewEventHandlerFunc(bot.GetEventDispatcher())

	// // 3. 注册路由并启动 HTTP 服务
	// http.HandleFunc("/webhook/event", handler)

	// port := ":48080"
	// log.Printf("🚀 go-tiny-claw 飞书服务端已启动，正在监听 %s 端口\n", port)

	// err := http.ListenAndServe(port, nil)
	// if err != nil {
	// 	log.Fatalf("服务器启动失败: %v", err)
	// }

	// websocket版
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")
	if appID == "" || appSecret == "" {
		log.Fatal("请设置 FEISHU_APP_ID 和 FEISHU_APP_SECRET")
	}

	// 2. 创建WebSocket客户端（官方标准参数）
	wsClient := larkws.NewClient(appID, appSecret,
		larkws.WithEventHandler(bot.GetEventDispatcher()), // 绑定事件调度器
		larkws.WithLogLevel(larkcore.LogLevelInfo),        // 设置日志级别
		larkws.WithAutoReconnect(true),                    // 开启自动重连（官方推荐）
	)
	// 官方启动方法：Start()会阻塞当前协程，保持长连接在线
	err := wsClient.Start(context.Background())
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
