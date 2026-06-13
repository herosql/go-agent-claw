# 测试缺口清单

---

## P0 — 必须补充（CI 必须全绿）

| # | 场景描述 | 建议测试类型 | 目标模块 |
|---|---|---|---|
| P0-01 | `Compactor.Compact` 输入未超阈值，直接返回原数组 | 单元测试 | `internal/context/compactor.go` |
| P0-02 | `Compactor.Compact` 超阈值时，历史区消息被正确掩码，保护区消息被掐头去尾 | 单元测试 | `internal/context/compactor.go` |
| P0-03 | `Compactor.Compact` 不修改 ToolCalls 字段 | 单元测试 | `internal/context/compactor.go` |
| P0-04 | `fuzzyReplace` L1 精确匹配：唯一匹配时正确替换 | 单元测试 | `internal/tools/edit_file.go` |
| P0-05 | `fuzzyReplace` L1 精确匹配：匹配 0 处返回 error | 单元测试 | `internal/tools/edit_file.go` |
| P0-06 | `fuzzyReplace` L1 精确匹配：匹配 >1 处返回 error（需更多上下文）| 单元测试 | `internal/tools/edit_file.go` |
| P0-07 | `fuzzyReplace` L4 逐行去缩进：消除缩进差异后正确匹配并替换 | 单元测试 | `internal/tools/edit_file.go` |
| P0-08 | `Session.Append` 并发安全：多 goroutine 并发写入不 panic | 并发测试 | `internal/context/session.go` |
| P0-09 | `Session.GetWorkingMemory` 截断时正确跳过 ToolResult 孤儿 | 单元测试 | `internal/context/session.go` |
| P0-10 | `Registry.Execute` 路由到不存在的工具返回 IsError=true | 单元测试 | `internal/tools/registry.go` |

---

## P1 — 应该补充（重要但非 CI 阻塞）

| # | 场景描述 | 建议测试类型 | 目标模块 |
|---|---|---|---|
| P1-01 | `Registry.Use` Middleware 链按注册顺序依次执行，First 返回 false 时后续不执行 | 单元测试 | `internal/tools/registry.go` |
| P1-02 | `Registry.Execute` Middleware 拦截危险工具时返回 IsError=true + rejectReason | 单元测试 | `internal/tools/registry.go` |
| P1-03 | `RecoveryManager.AnalyzeAndInject` 针对 `bash "command not found"` 返回含 hint 的错误 | 单元测试 | `internal/context/recovery.go` |
| P1-04 | `RecoveryManager.AnalyzeAndInject` 针对 `bash "超时"` 返回含超时 hint 的错误 | 单元测试 | `internal/context/recovery.go` |
| P1-05 | `ReminderInjector` 连续 3 次相同失败参数时返回打断指令，第 3 次成功后计数器清零 | 单元测试 | `internal/engine/reminder.go` |
| P1-06 | `ApprovalManager.WaitForApproval` 创建 channel 并立即阻塞 | 单元测试（mock）| `internal/feishu/approval.go` |
| P1-07 | `ApprovalManager.ResolveApproval` 接收到 allowed=true 后 WaitForApproval 解除阻塞 | 单元测试（mock）| `internal/feishu/approval.go` |
| P1-08 | `CostTracker.Generate` 解析出 Usage 时正确计算并累加到 Session | 单元测试（mock）| `internal/observability/tracker.go` |
| P1-09 | `CostTracker.Generate` 返回 Usage=nil 时不计费不 panic | 单元测试（mock）| `internal/observability/tracker.go` |
| P1-10 | `GlobalSessionMgr.GetOrCreate` 相同 ID 返回同一 Session 实例，不同 ID 创建不同实例 | 单元测试 | `internal/context/session.go` |