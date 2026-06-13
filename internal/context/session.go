// internal/context/session.go
package context

import (
	"sync"
	"time"

	"github.com/herosql/go-agent-claw/internal/schema"
)

// Session 代表了一次持续的人机交互过程。它负责维护该会话的完整历史。
type Session struct {
	ID        string
	WorkDir   string // 该会话绑定的物理工作区
	CreatedAt time.Time
	UpdatedAt time.Time

	// 存放此 Session 中所有的用户输入、大模型回复和工具调用结果
	history []schema.Message
	mu      sync.RWMutex // 读写锁，防止并发读写历史时发生 Data Race
}

func NewSession(id string, workDir string) *Session {
	return &Session{
		ID:        id,
		WorkDir:   workDir,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		history:   make([]schema.Message, 0),
	}
}

// Append 线程安全地向 Session 中追加消息
func (s *Session) Append(msgs ...schema.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, msgs...)
	s.UpdatedAt = time.Now()
}

// // GetWorkingMemory 是驾驭工程的核心！
// // 它不返回全量历史，而是从后往前截取最近的 N 条消息，形成 Agent 的"短期工作记忆"。
// func (s *Session) GetWorkingMemory(limit int) []schema.Message {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()

// 	total := len(s.history)
// 	if total <= limit || limit <= 0 {
// 		res := make([]schema.Message, total)
// 		copy(res, s.history)
// 		return res
// 	}

// 	// 截取最近的 limit 条消息
// 	res := make([]schema.Message, limit)
// 	copy(res, s.history[total-limit:])

// 	for len(res) > 0 {
// 		if res[0].Role == schema.RoleUser && res[0].ToolCallID != "" {
// 			res = res[1:]
// 		} else {
// 			break
// 		}
// 	}

//		return res
//	}
func (s *Session) GetWorkingMemory(limit int) []schema.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.history)
	if total <= limit || limit <= 0 {
		res := make([]schema.Message, total)
		copy(res, s.history)
		return res
	}

	res := make([]schema.Message, limit)
	copy(res, s.history[total-limit:])

	// 处理截断边缘的 ToolResult 孤儿问题
	for len(res) > 0 {
		if res[0].Role == schema.RoleUser && res[0].ToolCallID != "" {
			res = res[1:]
		} else {
			break
		}
	}

	return res
}

// ==========================================
// 全局 Session Manager: 用于多用户/多终端隔离
// ==========================================

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var GlobalSessionMgr = &SessionManager{
	sessions: make(map[string]*Session),
}

// GetOrCreate 获取或创建一个会话
func (sm *SessionManager) GetOrCreate(id string, workDir string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sess, exists := sm.sessions[id]; exists {
		return sess
	}
	sess := NewSession(id, workDir)
	sm.sessions[id] = sess
	return sess
}
