# 并行工具执行与流控

> 状态: 待扩展

---

## 1. 基于缓冲 Channel 的信号量隔离

### 核心方案：Semaphore Pattern

```go
// 初始化：最多允许 5 个工具并发
sem := make(chan struct{}, 5)

for _, toolCall := range toolCalls {
    go func(tc ToolCall) {
        // 策略 A: 外层阻塞（主循环阻塞）
        sem <- struct{}{}  // 阻塞等待令牌

        defer func() { <-sem }()  // 归还令牌

        executeTool(tc)
    }(toolCall)
}
```

---

## 2. 读写分流策略

| 工具类型 | 执行方式 | 示例 |
|----------|----------|------|
| 只读工具 | 并发执行 | ReadFile, Grep, Glob |
| 修改工具 | 串行执行 | Edit, Write, Bash |

---

## 3. RWMutex 锁机制（物理路径级别）

```go
type Registry struct {
    mu    sync.RWMutex
    tools map[string]*ToolDefinition
    // 按物理路径的写锁
    pathLocks map[string]*sync.Mutex
}
```

**策略**: 只读并发、涉写串行

---

## 4. Rate Limit 防护

如果一次性发起 50 个并发网络请求，可能被目标网站封杀。

**解决方案**: 使用带缓冲的 Channel 作为全局并发控制

```go
// 每个工作区独立的信号量
sem := make(chan struct{}, 5)  // 最多 5 个并发

// 熔断机制：队列满时快速失败
select {
case taskQueue <- task:
    // 入队成功
default:
    // 队列满，拒绝任务
    return ErrQueueFull
}
```

---

## 5. 待实现细节

- [ ] 基于 `semaphore` 的并发控制
- [ ] Registry 添加 `pathLocks` 映射
- [ ] 实现"只读并发、涉写串行"策略
- [ ] 队列满时的熔断返回