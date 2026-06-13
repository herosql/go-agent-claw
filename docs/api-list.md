# API 接口清单

## 外部接口（对外暴露）

### 飞书消息接收 Webhook
| 方法 | 路径 | 端口 | 来源 | 说明 |
|---|---|---|---|---|
| POST | `/webhook/event` | 48080 | `cmd/feishu` | 飞书开放平台 HTTP 回调，接收私聊消息事件 |

**请求体**（飞书平台格式）：
```json
{
  "event": {
    "message": {
      "chat_id": "oc_xxx",
      "content": "{\"text\":\"用户输入\"}"
    }
  }
}
```

**触发动作**：
- `approve <taskID>` → 解锁 ApprovalManager 协程，允许危险操作
- `reject <taskID>` → 拒绝危险操作
- 其他文本 → 新起 Goroutine 运行 Agent

---

### 飞书 WebSocket 长连接
| 启动端 | 来源 | 连接类型 | 说明 |
|---|---|---|---|
| `cmd/agentops` | AgentOps 完整版 | WebSocket | 带审批 Middleware，每个会话独立 Engine |
| `cmd/fei` | feishu variant | WebSocket | 简化版，单 Engine 共享 |

**连接参数**：
- AppID：`FEISHU_APP_ID`
- AppSecret：`FEISHU_APP_SECRET`
- 自动重连：`WithAutoReconnect(true)`

**事件处理**：
- `P2MessageReceiveV1` → 消息接收（私聊）
- `P2MessageReadV1` → 已读（静默忽略）

---

## 内部接口（cmd 之间不可调用，仅供内部参考）

### CLI 入口（无 HTTP）
| 可执行文件 | 命令示例 | 说明 |
|---|---|---|
| `cmd/claw` | `go run ./cmd/claw -prompt "任务" -dir . -session id` | 主 CLI 引擎 |
| `cmd/bench` | `go run ./cmd/bench` | Benchmark 评测 |
| `cmd/envcheck` | `go run ./cmd/envcheck` | 环境变量诊断 |

### 内部模块接口（关键接口定义）

#### LLMProvider 接口（internal/provider/interface.go）
```go
type LLMProvider interface {
    Generate(ctx context.Context, messages []schema.Message,
             availableTools []schema.ToolDefinition) (*schema.Message, error)
}
```
实现类：`OpenAIProvider`、`ClaudeProvider`、`CostTracker`（装饰器）

#### Registry 接口（internal/tools/registry.go）
```go
type Registry interface {
    Register(tool BaseTool)
    GetAvailableTools() []schema.ToolDefinition
    Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
    Use(mw MiddlewareFunc)  // 安全拦截中间件链
}
```

#### Reporter 接口（internal/engine/reporter.go）
```go
type Reporter interface {
    OnThinking(ctx context.Context)
    OnToolCall(ctx context.Context, toolName string, args string)
    OnToolResult(ctx context.Context, toolName string, result string, isError bool)
    OnMessage(ctx context.Context, content string)
}
```
实现类：`TerminalReporter`（终端）、`FeishuReporter`（飞书）

#### AgentRunner 接口（internal/tools/subagent.go）
```go
type AgentRunner interface {
    RunSub(ctx context.Context, taskPrompt string,
           readOnlyRegistry Registry, reporter interface{}) (string, error)
}
```
由 `AgentEngine.RunSub` 实现，打破 engine ↔ tools 循环依赖

#### Session Manager（internal/context/session.go）
```go
GlobalSessionMgr.GetOrCreate(id string, workDir string) *Session
// 线程安全，全局单例，按 sessionID 隔离多用户会话
```

#### ApprovalManager（internal/feishu/approval.go）
```go
GlobalApprovalMgr.WaitForApproval(taskID, toolName, args string,
                                   reporter *FeishuReporter) (bool, string)
// 阻塞协程，等待飞书回调 resolve
GlobalApprovalMgr.ResolveApproval(taskID string, allowed bool, reason string)
// 由 Webhook/WebSocket Handler 调用，解锁引擎协程
```

#### PromptComposer（internal/context/composer.go）
```go
NewPromptComposer(workDir string, planMode bool) *PromptComposer
composer.Build() schema.Message  // 动态拼装 System Prompt
```

#### Compactor（internal/context/compactor.go）
```go
NewCompactor(maxChars int=200000, retainLastMsgs int=6) *Compactor
Compact(msgs []schema.Message) []schema.Message
// 超阈值时：历史区全量掩码 + 保护区掐头去尾
```

#### SkillLoader（internal/context/skill.go）
```go
NewSkillLoader(workDir string) *SkillLoader
LoadAll() string  // 扫描 .claw/skills/*.md，解析 YAML Frontmatter
```

#### RecoveryManager（internal/context/recovery.go）
```go
AnalyzeAndInject(toolName string, rawError string) string
// 根据工具名 + 报错特征匹配自救指南
```

#### ReminderInjector（internal/engine/reminder.go）
```go
CheckAndInject(lastToolCall schema.ToolCall, lastResult schema.ToolResult) *schema.Message
// MD5 指纹检测连续失败 ≥3 次，注入打断指令
```

#### CostTracker（internal/observability/tracker.go）
```go
NewCostTracker(next LLMProvider, modelName string, session *Session) *CostTracker
// 装饰器：拦截 Usage，计算成本并累加到 session
```

#### Span / TraceExporter（internal/observability/trace.go）
```go
StartSpan(ctx context.Context, name string) (context.Context, *Span)
ExportTraceToFile(rootSpan *Span, workDir string, sessionID string)
// 导出到 .claw/traces/trace_<session>_<timestamp>.json
```

#### FeishuBot 工厂（internal/feishu/bot.go）
```go
NewFeishuBotWithFactory(factory AgentEngineFactory, workDir string) *FeishuBot
NewFeishuBot(eng *AgentEngine, sess *Session) *FeishuBot
// factory 允许每个会话独立创建 Engine + CostTracker
```

#### BenchmarkRunner（internal/eval/benchmark.go）
```go
NewBenchmarkRunner(modelName string) *BenchmarkRunner
RunSuite(ctx context.Context, testcases []TestCase)
// 沙箱隔离执行，ValidateScript exit 0 视为通过
```

---

## 数据流总览

```
飞书平台
  └─→ HTTP POST /webhook/event (cmd/feishu)
  └─→ WebSocket 长连接 (cmd/agentops / cmd/fei)
          ↓
     FeishuBot.GetEventDispatcher()
          ↓
     handleAgentRun(chatId, prompt)
          ↓
     GlobalSessionMgr.GetOrCreate(chatId, workDir)
          ↓
     AgentEngineFactory(sess) → 新 Engine + CostTracker(sess)
          ↓
     AgentEngine.Run(ctx, sess, FeishuReporter)
          ↓
     ├─ PromptComposer.Build() → System Prompt
     ├─ Compactor.Compact() → 上下文压缩
     ├─ LLMProvider.Generate() → 模型推理
     │     └─ CostTracker 计量
     ├─ Registry.Execute() → 工具执行
     │     └─ Middleware (高危挂起 → ApprovalManager)
     ├─ ReminderInjector.CheckAndInject() → 死循环检测
     └─ FeishuReporter → 飞书卡片推送
```