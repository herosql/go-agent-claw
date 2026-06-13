// internal/tools/registry.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/herosql/go-agent-claw/internal/observability"
	"github.com/herosql/go-agent-claw/internal/schema"
)

// MiddlewareFunc 定义了中间件的签名。
// 它接收当前的 ToolCall，并返回一个是否允许执行的布尔值 (allowed)，以及拦截时的原因 (rejectReason)。
type MiddlewareFunc func(ctx context.Context, call schema.ToolCall) (allowed bool, rejectReason string)

// BaseTool 是所有具体工具必须实现的通用接口
type BaseTool interface {
	// Name 返回工具的全局唯一名称
	Name() string
	// Definition 返回用于提交给大模型的工具元信息和参数 JSON Schema
	Definition() schema.ToolDefinition
	// Execute 接收大模型吐出的 JSON 参数，执行具体业务逻辑
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

// Registry 定义了工具的注册与分发接口
type Registry interface {
	Register(tool BaseTool)
	GetAvailableTools() []schema.ToolDefinition
	Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
	Use(mw MiddlewareFunc)
}

// registryImpl 是 Registry 接口的默认实现
type registryImpl struct {
	tools       map[string]BaseTool
	middlewares []MiddlewareFunc
}

func NewRegistry() Registry {
	return &registryImpl{
		tools:       make(map[string]BaseTool),
		middlewares: make([]MiddlewareFunc, 0),
	}
}

func (r *registryImpl) Register(tool BaseTool) {
	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		slog.Info("[Warning] 工具 '" + name + "' 已经被注册，将被覆盖。")
	}
	r.tools[name] = tool
	slog.Info("[Registry] 成功挂载工具: " + name)
}

func (r *registryImpl) Use(mw MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *registryImpl) GetAvailableTools() []schema.ToolDefinition {
	var defs []schema.ToolDefinition
	for _, tool := range r.tools {
		defs = append(defs, tool.Definition())
	}
	return defs
}

func (r *registryImpl) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	// 【埋点 5】：开启工具执行的 Span
	ctx, span := observability.StartSpan(ctx, "Tool.Execute")
	span.AddAttribute("tool_name", call.Name)
	span.AddAttribute("arguments", string(call.Arguments))

	defer span.EndSpan()

	// 1. 路由查找
	tool, exists := r.tools[call.Name]
	if !exists {
		span.AddAttribute("error", fmt.Sprintf("tool '%s' not found", call.Name))
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     fmt.Sprintf("Error: 系统中不存在名为 '%s' 的工具。", call.Name),
			IsError:    true,
		}
	}

	// 2. 【核心防御】在执行底层逻辑前，依次运行所有的 Middleware
	for _, mw := range r.middlewares {
		allowed, reason := mw(ctx, call)
		if !allowed {
			slog.Info("[Registry] ⚠️ 工具 " + call.Name + " 被 Middleware 拦截: " + reason)
			span.AddAttribute("intercepted", true)
			span.AddAttribute("reject_reason", reason)
			return schema.ToolResult{
				ToolCallID: call.ID,
				Output:     fmt.Sprintf("执行被系统拦截。原因: %s", reason),
				IsError:    true,
			}
		}
	}

	// 3. 执行工具逻辑
	output, err := tool.Execute(ctx, call.Arguments)
	if err != nil {
		span.AddAttribute("error", err.Error())
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     fmt.Sprintf("Error executing %s: %v", call.Name, err),
			IsError:    true,
		}
	}

	// 只截取输出的前 100 字符放入 Trace，防止 Trace 文件过度膨胀
	span.AddAttribute("output_preview", truncate(output, 100))

	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     output,
		IsError:    false,
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
