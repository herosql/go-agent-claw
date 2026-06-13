// internal/tools/registry.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
		log.Printf("[Warning] 工具 '%s' 已经被注册，将被覆盖。\n", name)
	}
	r.tools[name] = tool
	log.Printf("[Registry] 成功挂载工具: %s\n", name)
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
	// 1. 路由查找
	tool, exists := r.tools[call.Name]
	if !exists {
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
			log.Printf("[Registry] ⚠️ 工具 %s 被 Middleware 拦截: %s\n", call.Name, reason)
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
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     fmt.Sprintf("Error executing %s: %v", call.Name, err),
			IsError:    true,
		}
	}

	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     output,
		IsError:    false,
	}
}
