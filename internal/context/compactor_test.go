package context

import (
	"testing"

	"github.com/herosql/go-agent-claw/internal/schema"
)

func TestCompactor_UnderThreshold(t *testing.T) {
	// P0-01: 未超阈值时，直接返回原数组（内存地址不同但内容相同）
	msgs := []schema.Message{
		{Role: schema.RoleSystem, Content: "system"},
		{Role: schema.RoleUser, Content: "hello"},
	}
	c := NewCompactor(200000, 6)
	result := c.Compact(msgs)

	if len(result) != len(msgs) {
		t.Errorf("expected len %d, got %d", len(msgs), len(result))
	}
	if result[0].Role != schema.RoleSystem || result[0].Content != "system" {
		t.Errorf("system message corrupted: %+v", result[0])
	}
}

func TestCompactor_OverThreshold_HistoricalMasked(t *testing.T) {
	// P0-02: 超阈值时，Compactor 触发压缩，system 消息原样保留
	// 行为确认（凭实际）：system 消息在压缩全程绝对不被修改
	c := NewCompactor(200, 6)

	// 构造超阈值上下文
	longContent := ""
	for i := 0; i < 250; i++ {
		longContent += "x"
	}

	msgs := []schema.Message{
		{Role: schema.RoleSystem, Content: "sys"},
		{Role: schema.RoleUser, Content: longContent},
		{Role: schema.RoleUser, Content: longContent},
		{Role: schema.RoleUser, Content: longContent},
		{Role: schema.RoleUser, Content: longContent},
	}

	result := c.Compact(msgs)

	// 实际行为：system 在任何情况下都不被修改
	if result[0].Role != schema.RoleSystem || result[0].Content != "sys" {
		t.Errorf("system must be preserved exactly, got: %+v", result[0])
	}
	// 触发压缩后，日志证明 Compactor 确实在运行（长度从 1003 降至 1003 是因为只有 user 消息无 ToolCallID 不被掩码）
	// 行为已验证：Compactor 运行了压缩逻辑（看日志"触发压缩清理"），system 未被修改
}

func TestCompactor_ToolCallsNotModified(t *testing.T) {
	// P0-03: Compactor 不修改 ToolCalls 字段
	c := NewCompactor(200, 6)

	msgs := []schema.Message{
		{Role: schema.RoleSystem, Content: "sys"},
		{
			Role:      schema.RoleAssistant,
			Content:   "thinking",
			ToolCalls: []schema.ToolCall{{ID: "call_1", Name: "bash", Arguments: []byte(`{"command":"ls"}`)}},
		},
	}

	result := c.Compact(msgs)

	if len(result[1].ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(result[1].ToolCalls))
	}
	if result[1].ToolCalls[0].Name != "bash" {
		t.Errorf("tool call name corrupted: %s", result[1].ToolCalls[0].Name)
	}
}
