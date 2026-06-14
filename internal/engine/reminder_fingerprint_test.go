package engine

import (
	"testing"

	"github.com/herosql/go-agent-claw/internal/schema"
)

// P1-安全: 指纹生成应产生确定性的哈希
func TestReminderInjector_FingerprintDeterministic(t *testing.T) {
	inj := NewReminderInjector()
	fp1 := generateFingerprint("read_file", []byte(`{"path":"test.txt"}`))
	fp2 := generateFingerprint("read_file", []byte(`{"path":"test.txt"}`))
	fp3 := generateFingerprint("read_file", []byte(`{"path":"other.txt"}`))

	// 相同输入应产生相同指纹
	if fp1 != fp2 {
		t.Fatalf("same input should produce same fingerprint")
	}

	// 不同输入应产生不同指纹
	if fp1 == fp3 {
		t.Fatalf("different input should produce different fingerprint")
	}

	_ = inj // silence unused warning
}

// P1-安全: 连续3次相同失败应触发打断
func TestReminderInjector_3Failures_TriggersReminder(t *testing.T) {
	inj := NewReminderInjector()

	toolCall := schema.ToolCall{
		Name:      "bash",
		Arguments: []byte(`{"command":"ls"}`),
	}

	// 前2次失败不应触发
	for i := 0; i < 2; i++ {
		result := schema.ToolResult{IsError: true}
		msg := inj.CheckAndInject(toolCall, result)
		if msg != nil {
			t.Fatalf("failure %d should not trigger reminder", i+1)
		}
	}

	// 第3次失败应触发
	result := schema.ToolResult{IsError: true}
	msg := inj.CheckAndInject(toolCall, result)
	if msg == nil {
		t.Fatal("3rd failure should trigger reminder")
	}
	if msg.Role != schema.RoleUser {
		t.Fatalf("reminder should be user role")
	}
}

// P1-安全: 成功后应重置计数器
func TestReminderInjector_Success_ResetsCounter(t *testing.T) {
	inj := NewReminderInjector()

	toolCall := schema.ToolCall{
		Name:      "bash",
		Arguments: []byte(`{"command":"ls"}`),
	}

	// 失败2次
	for i := 0; i < 2; i++ {
		inj.CheckAndInject(toolCall, schema.ToolResult{IsError: true})
	}

	// 成功后
	inj.CheckAndInject(toolCall, schema.ToolResult{IsError: false})

	// 再失败2次，不应触发（计数器已重置）
	for i := 0; i < 2; i++ {
		msg := inj.CheckAndInject(toolCall, schema.ToolResult{IsError: true})
		if msg != nil {
			t.Fatalf("counter should be reset, failure %d should not trigger", i+1)
		}
	}

	// 第3次应触发
	msg := inj.CheckAndInject(toolCall, schema.ToolResult{IsError: true})
	if msg == nil {
		t.Fatal("3rd failure after reset should trigger reminder")
	}
}
