// internal/engine/loop.go
package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/herosql/go-agent-claw/internal/provider"
	"github.com/herosql/go-agent-claw/internal/schema"
	"github.com/herosql/go-agent-claw/internal/tools"
)

type AgentEngine struct {
	provider       provider.LLMProvider
	registry       tools.Registry
	WorkDir        string
	EnableThinking bool // 【新增】慢思考模式开关
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		WorkDir:        workDir,
		EnableThinking: enableThinking,
	}
}

// internal/engine/loop.go (续)

func (e *AgentEngine) Run(ctx context.Context, userPrompt string) error {
	slog.Info("[Engine] 引擎启动，锁定工作区", "workDir", e.WorkDir)
	slog.Info("[Engine] 慢思考模式", "enableThinking", e.EnableThinking)

	contextHistory := []schema.Message{
		{
			Role:    schema.RoleSystem,
			Content: "You are go-tiny-claw, an expert coding assistant. You have full access to tools in the workspace.",
		},
		{
			Role:    schema.RoleUser,
			Content: userPrompt,
		},
	}

	turnCount := 0

	for {
		turnCount++
		slog.Info("[Turn] 开始", "turn", turnCount)

		// 获取当前挂载的所有工具定义
		availableTools := e.registry.GetAvailableTools()

		// ====================================================================
		// Phase 1: 慢思考阶段 (Thinking) - 剥夺工具，强制规划
		// ====================================================================
		if e.EnableThinking {
			slog.Info("[Phase 1] 剥夺工具访问权，强制进入慢思考与规划阶段...")

			// 核心机制：传入的 availableTools 为 nil！
			// 大模型看不到任何 JSON Schema，被迫只能输出纯文本的思考过程。
			thinkResp, err := e.provider.Generate(ctx, contextHistory, nil)
			if err != nil {
				return fmt.Errorf("Thinking 阶段生成失败: %w", err)
			}

			// 如果模型输出了思考过程，我们将其作为 Assistant 消息追加到上下文中
			if thinkResp.Content != "" {
				fmt.Printf("🧠 [内部思考 Trace]: %s\n", thinkResp.Content)
				contextHistory = append(contextHistory, *thinkResp)
			}
		}

		// ====================================================================
		// Phase 2: 行动阶段 (Action) - 恢复工具，顺着规划执行
		// ====================================================================
		slog.Info("[Phase 2] 恢复工具挂载，等待模型采取行动...")

		// 此时的 contextHistory 中已经包含了上一阶段模型自己的 Thinking Trace。
		// 模型会顺着自己的逻辑，结合恢复的 availableTools 发起精准的工具调用。
		actionResp, err := e.provider.Generate(ctx, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("Action 阶段生成失败: %w", err)
		}

		contextHistory = append(contextHistory, *actionResp)

		if actionResp.Content != "" {
			fmt.Printf("🤖 [对外回复]: %s\n", actionResp.Content)
		}

		// ====================================================================
		// 退出与执行逻辑 (与上一讲保持一致)
		// ====================================================================
		if len(actionResp.ToolCalls) == 0 {
			slog.Info("[Engine] 模型未请求调用工具，任务宣告完成。")
			break
		}

		slog.Info("[Engine] 模型请求调用工具", "count", len(actionResp.ToolCalls))

		for _, toolCall := range actionResp.ToolCalls {
			slog.Info("[Engine] 执行工具", "name", toolCall.Name, "args", string(toolCall.Arguments))

			result := e.registry.Execute(ctx, toolCall)

			if result.IsError {
				slog.Error("[Engine] 工具执行报错", "output", result.Output)
			} else {
				slog.Info("[Engine] 工具执行成功", "bytes", len(result.Output))
			}

			// 将工具执行的观察结果追加到 Context，准备进入下一轮
			observationMsg := schema.Message{
				Role:       schema.RoleUser,
				Content:    result.Output,
				ToolCallID: toolCall.ID,
			}
			contextHistory = append(contextHistory, observationMsg)
		}
	}

	return nil
}
