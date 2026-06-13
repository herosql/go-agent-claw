# Lint 问题修复待办清单

> 生成时间: 2026-06-13
> 剩余问题数: 113

---

## P0 - 必须修复（功能/安全问题）

### G304 路径穿越漏洞 (1)

| 文件 | 行号 | 问题 |
|------|------|------|
| `internal/tools/edit_file.go` | 166 | `os.ReadFile(fullPath)` 未做路径安全验证 |

**修复方案**: 添加 `isPathSafe` 检查，参考 `read_file.go` 和 `write_file.go` 的实现。

---

### G204/G501/G401 安全相关 (4)

| 文件 | 行号 | 问题 |
|------|------|------|
| `internal/tools/bash.go` | 59 | `exec.CommandContext` 潜在命令注入 |
| `internal/engine/reminder.go` | 5 | `crypto/md5` 弱加密算法 |
| `internal/engine/reminder.go` | 27 | `md5.New()` 弱加密原语 |

**修复方案**:
- `bash.go`: 添加输入验证或 `#nosec G204` 注释（工具本身需要执行命令）
- `reminder.go`: 考虑使用 `crypto/sha256` 替代 MD5

---

### errcheck 错误检查 (1)

| 文件 | 行号 | 问题 |
|------|------|------|
| `internal/feishu/bot.go` | 193 | `r.client.Im.Message.Create` 返回值未检查 |

**修复方案**: 检查错误并记录日志

---

## P1 - 重要（代码质量）

### fieldalignment 结构体内存对齐 (10)

| 文件 | 结构体 | 当前 | 建议 |
|------|--------|------|------|
| `internal/schema/messge.go:21` | `Message` | 80 bytes | 64 bytes |
| `internal/schema/messge.go:51` | `ToolDefinition` | 48 bytes | 40 bytes |
| `internal/observability/trace.go:18` | `Span` | 88 bytes | 80 bytes |
| `internal/observability/tracker.go:24` | `CostTracker` | 40 bytes | 32 bytes |
| `internal/feishu/approval.go:18` | `ApprovalManager` | 32 bytes | 8 bytes |
| `internal/eval/benchmark.go:31` | `TestResult` | 48 bytes | 24 bytes |
| `internal/context/composer.go:26` | `PromptComposer` | 32 bytes | 16 bytes |
| `internal/context/session.go:12` | `Session` | 152 bytes | 88 bytes |

**修复方案**: 按指针大小重排字段顺序（较小类型在前）

---

### prealloc 预分配 (1)

| 文件 | 行号 | 问题 |
|------|------|------|
| `internal/eval/benchmark.go:63` | `results = append(results, res)` 返回值未使用 |

**修复方案**: 移除无用变量或正确使用返回值

---

## P2 - 代码文档（注释缺失）

### revive 导出函数/类型缺少注释 (~80)

**批量修复方案**: 为以下导出的函数和类型添加 GoDoc 注释：

#### internal/tools
- `WriteFileTool` (type)
- `NewWriteFileTool` (func)
- `WriteFileTool.Name` (method)
- `WriteFileTool.Definition` (method)
- `WriteFileTool.Execute` (method)
- `EditFileTool` (type) - 已有部分注释
- `ReadFileTool` (type) - 已有部分注释
- `NewReadFileTool` (func) - 已有部分注释
- `ReadFileTool.Name` (method) - 已有部分注释
- `ReadFileTool.Execute` (method) - 已有部分注释
- `BashTool` (type) - 已有部分注释
- `NewBashTool` (func) - 已有部分注释
- `BashTool.Name` (method) - 已有部分注释
- `BashTool.Definition` (method) - 已有部分注释
- `BashTool.Execute` (method) - 已有部分注释
- `SubagentTool` (type)
- `SubagentTool.Name` (method)
- `SubagentTool.Execute` (method)
- `NewRegistry` (func) - 已有部分注释

#### internal/context
- `NewCompactor` (func) - 已有部分注释
- `NewPromptComposer` (func) - 已有部分注释
- `PromptComposer.Build` (method) - 已有部分注释
- `NewSession` (func) - 已有部分注释
- `Session.GetWorkingMemory` (method) - 已有部分注释
- `SessionManager` (type)
- `GlobalSessionMgr` (var) - 已有注释，需修改格式

#### internal/engine
- `AgentEngine` (type) - 已有部分注释
- `NewAgentEngine` (func) - 已有部分注释
- `AgentEngine.Run` (method) - 已有部分注释
- `NewReminderInjector` (func)
- `TerminalReporter` (type) - 已有部分注释
- `NewTerminalReporter` (func) - 已有部分注释
- `TerminalReporter.OnThinking` (method)
- `TerminalReporter.OnToolCall` (method)
- `TerminalReporter.OnToolResult` (method)
- `TerminalReporter.OnMessage` (method)

#### internal/provider
- `OpenAIProvider` (type)
- `NewZhipuOpenAIProvider` (func)
- `OpenAIProvider.Generate` (method)
- `ClaudeProvider` (type)
- `NewZhipuClaudeProvider` (func)
- `ClaudeProvider.Generate` (method)

#### internal/feishu
- `FeishuBot` (type)
- `NewFeishuBotWithFactory` (func)
- `FeishuBot.GetEventDispatcher` (method)
- `FeishuReporter` (type)
- `FeishuReporter.OnThinking` (method)
- `FeishuReporter.OnToolCall` (method)
- `FeishuReporter.OnToolResult` (method)
- `FeishuReporter.OnMessage` (method)

#### internal/eval
- `BenchmarkRunner` (type)
- `NewBenchmarkRunner` (func)

#### 其他
- `internal/observability/tracker.go` - package comment
- `internal/observability/trace.go` - package comment
- `internal/context/compactor.go` - package comment
- `internal/context/recovery.go` - package comment
- `internal/context/composer.go` - package comment
- `internal/context/session.go` - package comment
- `internal/context/skill.go` - package comment
- `internal/tools/write_file.go` - package comment
- `internal/tools/read_file.go` - package comment
- `internal/tools/bash.go` - package comment
- `internal/tools/edit_file.go` - package comment
- `internal/tools/subagent.go` - package comment
- `internal/tools/registry.go` - package comment
- `internal/engine/terminal_reporter.go` - package comment
- `internal/engine/reporter.go` - package comment
- `internal/engine/reminder.go` - package comment
- `internal/feishu/bot.go` - package comment
- `internal/feishu/approval.go` - package comment
- `internal/eval/benchmark.go` - package comment
- `internal/schema/messge.go` - package comment, const comment

---

## 已修复清单

| 日期 | 类别 | 数量 |
|------|------|------|
| 2026-06-13 | errorlint | 1 |
| 2026-06-13 | shadow | 6 |
| 2026-06-13 | G304 | 5 |
| 2026-06-13 | fieldalignment | 2 |
| 2026-06-13 | prealloc | 5 |
| 2026-06-13 | 目录权限 0755→0750 | 4 |
| 2026-06-13 | 文件权限 0644→0600 | 3 |
| 2026-06-13 | http timeout | 1 |
| 2026-06-13 | chatId→chatID | 4 |
| 2026-06-13 | errcheck | 2 |
| 2026-06-13 | gofmt/goimports | 8+ |