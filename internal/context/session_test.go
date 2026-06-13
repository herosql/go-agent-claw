package context

import (
	"sync"
	"testing"

	"github.com/herosql/go-agent-claw/internal/schema"
)

// P0-08: Session.Append 并发安全，多 goroutine 并发写入不 panic
func TestSession_Append_Concurrent(t *testing.T) {
	s := NewSession("test-concurrent", "/tmp")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s.Append(schema.Message{Role: schema.RoleUser, Content: "msg"})
		}(i)
	}
	wg.Wait()

	// 不 panic 即通过；长度应为 100
	if len(s.history) != 100 {
		t.Errorf("expected 100 messages, got %d", len(s.history))
	}
}

// P0-09: GetWorkingMemory 截断时正确跳过 ToolResult 孤儿
func TestSession_GetWorkingMemory_SkipsOrphanToolResults(t *testing.T) {
	s := NewSession("test-truncation", "/tmp")

	// 构造场景：开头是 ToolResult 孤儿（RoleUser + ToolCallID）
	s.Append(schema.Message{Role: schema.RoleUser, Content: "result 1", ToolCallID: "call_1"})
	s.Append(schema.Message{Role: schema.RoleUser, Content: "result 2", ToolCallID: "call_2"})
	s.Append(schema.Message{Role: schema.RoleUser, Content: "normal msg"})

	// GetWorkingMemory(2) 取最后 2 条：[call_2 result, normal msg]
	// 然后跳过头部孤儿（第一条是 call_2 result，是 RoleUser+ToolCallID，所以跳过）
	// 实际行为：跳过头部孤儿后，只剩 1 条 normal msg
	result := s.GetWorkingMemory(2)

	// 行为确认（凭实际）：
	// 1. 不 panic（并发安全已由 P0-08 覆盖）
	// 2. 返回消息中无孤儿 ToolResult（第一条有 ToolCallID 的被跳过）
	if len(result) != 1 {
		t.Errorf("expected 1 message after skipping orphan, got %d", len(result))
	}
	if result[0].Content != "normal msg" {
		t.Errorf("expected 'normal msg', got '%s'", result[0].Content)
	}
}

// P1-10: GlobalSessionMgr 相同 ID 返回同一实例，不同 ID 创建不同实例
func TestGlobalSessionMgr_GetOrCreate(t *testing.T) {
	mgr := &SessionManager{sessions: make(map[string]*Session)}

	s1 := mgr.GetOrCreate("id1", "/tmp")
	s2 := mgr.GetOrCreate("id1", "/tmp")
	s3 := mgr.GetOrCreate("id2", "/tmp")

	if s1 != s2 {
		t.Errorf("same ID should return same instance")
	}
	if s1 == s3 {
		t.Errorf("different IDs should return different instances")
	}
}
