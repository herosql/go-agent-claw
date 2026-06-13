// cmd/claw/main.go
package main

import (
	"context"
	"log"
	"os"

	ctxpkg "github.com/herosql/go-agent-claw/internal/context"
	"github.com/herosql/go-agent-claw/internal/engine"
	"github.com/herosql/go-agent-claw/internal/feishu"
	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/schema"
	"github.com/herosql/go-agent-claw/internal/tools"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// cmd/claw/main.go

func main() {
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先导出 ZHIPU_API_KEY 环境变量")
	}

	workDir, _ := os.Getwd()
	workDir += "/workspace"

	llmProvider := provider.NewZhipuOpenAIProvider("glm-4.5-air")

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	eng := engine.NewAgentEngine(llmProvider, registry, false, false)

	// 假设一个bot绑定一个session
	sessionID := "test_command_intercept_001"
	sess := ctxpkg.GlobalSessionMgr.GetOrCreate(sessionID, workDir)
	sess.Append(schema.Message{Role: schema.RoleUser, Content: ""})

	bot := feishu.NewFeishuBot(eng, sess)
	// http版
	// handler := httpserverext.NewEventHandlerFunc(bot.GetEventDispatcher())

	// 【核心注入】注册安全拦截 Middleware
	registry.Use(func(ctx context.Context, call schema.ToolCall) (bool, string) {
		argsStr := string(call.Arguments)

		// 检查是否命中高危特征库
		if feishu.IsDangerousCommand(call.Name, argsStr) {
			taskID := call.ID // 使用大模型生成的唯一 ToolCallID 作为 TaskID

			// 挂起当前协程，发送消息给飞书，死死等待人类的审批！
			allowed, reason := feishu.GlobalApprovalMgr.WaitForApproval(taskID, call.Name, argsStr, bot.Reporter())

			if !allowed {
				return false, reason // 拒绝，将理由传回给大模型
			}
			return true, "" // 同意，放行底层工具
		}

		// 没命中黑名单，直接 YOLO 放行
		return true, ""
	})

	// 3. 注册路由并启动 HTTP 服务
	// http.HandleFunc("/webhook/event", handler)

	// port := ":48080"
	// log.Printf("🚀 go-tiny-claw 飞书服务端已启动，正在监听 %s 端口\n", port)

	// err := http.ListenAndServe(port, nil)
	// if err != nil {
	// 	log.Fatalf("服务器启动失败: %v", err)
	// }

	// websocket 版
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
