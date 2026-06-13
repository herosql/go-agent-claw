package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/herosql/go-agent-claw/internal/schema"
)

// P0-10: Registry.Execute 路由到不存在的工具返回 IsError=true
func TestRegistry_Execute_ToolNotFound(t *testing.T) {
	reg := NewRegistry()
	// 不注册任何工具

	result := reg.Execute(context.Background(), schema.ToolCall{
		ID:        "call_1",
		Name:      "non_existent_tool",
		Arguments: json.RawMessage(`{}`),
	})

	if !result.IsError {
		t.Errorf("expected IsError=true for non-existent tool")
	}
	if result.ToolCallID != "call_1" {
		t.Errorf("expected ToolCallID preserved, got %s", result.ToolCallID)
	}
}
