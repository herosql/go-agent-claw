# 飞书工作区任务调度

> 状态: 待扩展

---

## 1. 问题背景

当群里同时有多人艾特机器人时：
- 后台瞬间拉起多个 Main Loop
- 并发在同一 WorkDir 执行命令
- 可能导致文件锁冲突和状态混乱

---

## 2. 分片队列管理器

```go
type WorkspaceScheduler struct {
    mu       sync.RWMutex
    queues   map[string]chan *Task  // workDir -> task queue
    workers  map[string]context.CancelFunc  // workDir -> worker cancel
}

type Task struct {
    SessionID string
    UserID    string
    Prompt    string
    ReplyCh   chan *TaskResult
}
```

**设计原则**: 不同工作区并行，同工作区排队

---

## 3. 流控机制

### 队列容量限制

```go
const MaxQueuePerWorkDir = 3

func (ws *WorkspaceScheduler) Enqueue(task *Task) error {
    ws.mu.Lock()
    queue, ok := ws.queues[task.WorkDir]
    if !ok {
        queue = make(chan *Task, MaxQueuePerWorkDir)
        ws.queues[task.WorkDir] = queue
    }
    ws.mu.Unlock()

    select {
    case queue <- task:
        return nil  // 入队成功
    default:
        return ErrQueueFull  // 队列满，快速失败
    }
}
```

### 即时反馈

当触发满载熔断时，返回：

```json
{
  "msg": "⚠️ 系统正忙：智能助手目前正在处理该项目下的高优先级任务，排队队列已满。请稍候再试。"
}
```

---

## 4. Worker 协程

```go
func (ws *WorkspaceScheduler) startWorker(workDir string) {
    queue := ws.queues[workDir]

    for task := range queue {
        engine := GetOrCreateEngine(workDir)

        result := engine.Run(task.Context, task.Prompt)

        task.ReplyCh <- result
    }
}
```

---

## 5. 熔断保护

- 队列满时：**零成本**（未向 LLM 发起 Token 请求）
- HTTP 线程：**200 OK** 立即返回
- 用户感知：收到友好的"系统正忙"提示

---

## 6. 待实现细节

- [ ] WorkspaceScheduler 结构
- [ ] 分片队列的创建与销毁
- [ ] Worker 协程的生命周期管理
- [ ] 队列满时的熔断返回
- [ ] 飞书消息的即时反馈