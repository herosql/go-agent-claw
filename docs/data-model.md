# 数据模型

> 无持久化 DB，所有状态均为内存或文件。所有"持久化"通过工作区文件系统实现。

---

## 核心实体

### Message（schema）
```go
type Message struct {
    Role       Role           // system | user | assistant
    Content    string
    ToolCalls  []ToolCall    // assistant 消息中模型决定调用的工具
    ToolCallID string        // user 消息中回复某次工具调用的结果
    Usage      *Usage        // 本次 API 调用的 Token 消耗
}

type Role string
const RoleSystem    Role = "system"
const RoleUser     Role = "user"
const RoleAssistant Role = "assistant"

type Usage struct {
    PromptTokens     int
    CompletionTokens int
}
```
**说明**：`ToolCallID` 用于将工具返回结果（user 角色）关联到对应的 ToolCall，形成完整的工具调用链。

---

### ToolCall（schema）
```go
type ToolCall struct {
    ID        string
    Name      string      // 工具名，如 "bash"、"read_file"
    Arguments json.RawMessage  // JSON 参数
}
```

---

### ToolResult（schema）
```go
type ToolResult struct {
    ToolCallID string
    Output    string   // 工具执行的控制台输出或报错堆栈
    IsError   bool
}
```

---

### ToolDefinition（schema）
```go
type ToolDefinition struct {
    Name        string
    Description string
    InputSchema interface{}  // JSON Schema
}
```

---

### Session（context）
```go
type Session struct {
    ID                   string
    WorkDir              string
    CreatedAt, UpdatedAt time.Time
    TotalPromptTokens    int
    TotalCompletionTokens int
    TotalCostCNY         float64
    history              []schema.Message  // 完整消息历史
    mu                  sync.RWMutex
}
```
**生命周期**：随 `GlobalSessionMgr.GetOrCreate` 创建，进程级常驻，支持断点续传。

---

### Span（observability）
```go
type Span struct {
    Name        string
    StartTime   time.Time
    EndTime     time.Time
    DurationMs  int64
    Attributes  map[string]interface{}
    Children    []*Span   // 树形父子关系
    mu          sync.Mutex
}
```
**生命周期**：一次任务一根 Root Span，每个 Turn、每次 API 调用、每次工具执行均创建子 Span。任务结束时序列化到 `.claw/traces/trace_<session>_<timestamp>.json`。

---

### Skill（context）
```go
type Skill struct {
    Name        string   // 从 YAML frontmatter 解析
    Description string   // 从 YAML frontmatter 解析
    Body        string   // Markdown 正文
}
```
**来源**：`.claw/skills/<dir>/SKILL.md`，由 `SkillLoader.LoadAll()` 扫描解析并拼入 System Prompt。

---

### ApprovalResult / ApprovalManager（feishu）
```go
type ApprovalResult struct {
    Allowed bool
    Reason  string
}

type ApprovalManager struct {
    mu           sync.RWMutex
    pendingTasks  map[string]chan ApprovalResult  // taskID → 解锁 channel
}
var GlobalApprovalMgr = &ApprovalManager{...}
```
**生命周期**：`WaitForApproval` 创建容量为 1 的 channel 并阻塞，`ResolveApproval` 由飞书 Webhook 在收到审批口令后调用写入结果。

---

### Registry（tools）
```go
type Registry interface {
    Register(tool BaseTool)
    GetAvailableTools() []schema.ToolDefinition
    Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
    Use(mw MiddlewareFunc)
}

type MiddlewareFunc func(ctx context.Context, call schema.ToolCall) (bool, string)
```
**Middleware 链顺序**：按注册顺序依次执行，任一返回 `allowed=false` 立即拒绝。

---

### CostTracker（observability）
```go
type CostTracker struct {
    nextProvider LLMProvider
    modelName    string
    session      *Session
}

var PricingModel = map[string]struct {
    InputPrice  float64  // 美元/1M Tokens
    OutputPrice float64
}{
    "glm-4.5-air": {InputPrice: 0.15, OutputPrice: 0.15},
}
```

---

### TestCase / TestResult（eval）
```go
type TestCase struct {
    ID             string  // 唯一标识
    Name           string
    SetupScript    string  // 运行前 bash 脚本（可选）
    TaskPrompt     string  // 发给 Agent 的任务描述
    ValidateScript string  // 运行后 bash 校验脚本，exit 0 = 通过
    MaxTurns       int
}

type TestResult struct {
    TestCaseID   string
    Passed       bool
    TotalCostCNY float64
    DurationMs   int64
    ErrorMsg     string
}
```

---

## 文件型持久化结构

| 文件路径 | 格式 | 触发写入 | 读取方 |
|---|---|---|---|
| `.claw/traces/trace_<sid>_<ts>.json` | JSON（Span 树） | `AgentEngine.Run` 结束时 | 人工复盘 |
| `.claw/skills/<name>/SKILL.md` | Markdown+YAML | 人工维护 | `SkillLoader.LoadAll()` |
| `<workDir>/AGENTS.md` | Markdown | 人工维护 | `PromptComposer.Build()` |
| `<workDir>/PLAN.md` | Markdown | Agent 在 PlanMode 下创建 | Agent 自身断点续传 |
| `<workDir>/TODO.md` | Markdown（Checkbox） | Agent 在 PlanMode 下创建/更新 | Agent 自身断点续传 |

---

## ER 关系（内存模型）

```
Session 1──N Message
Session 1──1 CostTracker (通过 session 字段)
Session 1──N Span (RootSpan → Children 树)

Message 1──N ToolCall
ToolCall 1──1 ToolResult (通过 ToolCallID 关联)

ApprovalManager 1──N pendingTasks (taskID → channel)

Registry 1──N BaseTool
Registry 1──N MiddlewareFunc

AgentEngine 1──1 Compactor
AgentEngine 1──1 ReminderInjector
AgentEngine 1──1 PromptComposer
AgentEngine 1──1 RecoveryManager
```