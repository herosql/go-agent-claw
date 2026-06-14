# 成本与状态追踪：在 Harness 层拦截并记录 Token 消耗与执行耗时

> 状态: 待扩展

---

## 背景与问题

如果一个 bash 命令执行了一个需要编译 5 分钟的巨型 Go 项目，这 5 分钟的物理世界耗时，目前的 CostTracker 捕获不到。需要编写一个能记录"工具在本地物理执行真正耗费了多少毫秒"的拦截器，且不修改 internal/tools/bash.go 源码。

---

## 核心方案：物理工具耗时拦截器

### 1. 结构层：设计工具执行上下文（ToolContext）

```go
type ToolContext struct {
    ToolName   string
    Arguments  map[string]interface{}
    Ctx        context.Context
}

type ToolResult struct {
    Output   string
    IsError  bool
    Duration time.Duration
}
```

### 2. 行为层：工具级装饰器中间件（Tool Runtime Middleware）

```go
// 符合 MiddlewareFunc 签名的核心拦截器
func NewToolDurationTracker() tools.MiddlewareFunc {
    return func(next tools.ToolExecutor) tools.ToolExecutor {
        return func(ctx context.Context, args map[string]interface{}) (string, error) {
            // 前置：打下起始时间戳
            start := time.Now()
            
            // 流转下发：调用真正的物理工具（同步挂起）
            output, err := next(ctx, args)
            
            // 后置：精准计算物理耗时
            duration := time.Since(start)
            
            // 提取工具名
            toolName := extractToolName(ctx)
            
            // 指标落盘与审计
            record := ToolAuditRecord{
                ToolName: toolName,
                Duration: duration,
                IsError:  err != nil,
                OutputLen: len(output),
            }
            
            // 异步投递监控系统 or 调用 Session.RecordToolDuration
            go func() {
                session := extractSession(ctx)
                session.RecordToolDuration(toolName, duration)
                emitToMonitoring(record)
            }()
            
            // 打印审计日志
            log.Printf("[Tracker][Tool-Audit] 🔧 工具 [%s] 执行完毕 | 物理耗时: %d ms | 状态: %s",
                toolName, duration.Milliseconds(), map[bool]string{true: "FAILED", false: "SUCCESS"}[err != nil])
            
            return output, err
        }
    }
}
```

### 3. 挂载层：在 cmd/claw/main.go 中套娃挂载

```go
registry := tools.NewRegistry()

// 注入可观测性工具拦截器（最外层套娃）
registry.Use(observability.NewToolDurationTracker())
registry.Use(observability.NewTokenTracker())

// 注册那些纯粹的、没有任何监控代码的物理工具
registry.Register(tools.NewBashTool())
registry.Register(tools.NewReadFileTool())
registry.Register(tools.NewWriteFileTool())
```

---

## 核心数据结构

```go
type ToolAuditRecord struct {
    ToolName   string        `json:"tool_name"`
    Duration   time.Duration `json:"duration_ms"`
    IsError    bool          `json:"is_error"`
    OutputLen  int           `json:"output_len"`
    Timestamp  time.Time     `json:"timestamp"`
}

type CostSummary struct {
    TotalPromptTokens     int64             `json:"total_prompt_tokens"`
    TotalCompletionTokens int64             `json:"total_completion_tokens"`
    TotalCost             float64           `json:"total_cost"`
    ToolDurations         map[string]time.Duration `json:"tool_durations"`  // 工具名 -> 累计耗时
}
```

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **多 Provider 的成本合并**：多个 Provider 同时使用时的成本汇总
2. **工具耗时的告警阈值**：长耗时工具的自动报警机制

### 待扩展2

> 未来规划方向

1. **火焰图集成**：基于工具耗时的性能分析可视化
2. **成本预算控制**：接近月度预算时自动降级