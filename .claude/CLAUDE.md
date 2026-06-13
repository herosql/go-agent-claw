# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 核心原则导入（最高优先级）

- **项目宪法**：`./constitution.md` — 具有最高法律效力，所有规则不得与宪法冲突
- **通用规则**：`./AGENTS.md` — 项目工程规范（命令、环境、Git、测试、架构约定）
- **冲突兜底**：宪法优先；AGENTS.md 与本文件不得与宪法冲突

---

## 核心使命与角色设定

你是一个资深的 Go 语言工程师，正在协助开发 **go-agent-claw** AI Agent Harness。所有行动必须严格遵守项目宪法。

---

## Claude Code 专属配置

### Sub-agent
- 需要安全审查时，调用 `security-reviewer` sub-agent

### Hooks
- 每次代码编辑后自动运行 `gofmt`

### 个人偏好
- 注释优先中文
- 输出简洁：代码 + 极简解释，不冗余
- 实现计划：收到需求后先输出计划，不直接写代码
- 导入配置：`~/.claude/my-personal-go-prefs.md`（**不存在，待补充**）

---

## 项目概述

详见 `README.md`。关键架构图见 `docs/architecture.mmd`。

---

## 关键模块

详见 `docs/module-deps.mmd`（模块依赖图）和 `docs/external-deps.mmd`（外部依赖图）。

| 模块 | 路径 | 一句话职责 |
|---|---|---|
| AgentEngine | `internal/engine/loop.go` | 两阶段执行循环 + 工具并行调度 |
| Registry | `internal/tools/registry.go` | 工具注册与 Middleware 链式拦截 |
| LLMProvider | `internal/provider/interface.go` | 统一大模型通信契约（OpenAI/Claude 双协议）|
| CostTracker | `internal/observability/tracker.go` | Provider 装饰器，计量 Token 成本 |
| Session | `internal/context/session.go` | 维护历史，支持断点续传 |
| Compactor | `internal/context/compactor.go` | 上下文超阈值压缩 |
| PromptComposer | `internal/context/composer.go` | 动态拼装 System Prompt（PlanMode 强制 PLAN.md/TODO.md）|
| ApprovalManager | `internal/feishu/approval.go` | Channel 阻塞引擎协程，等待飞书 approve/reject |
| FeishuBot | `internal/feishu/bot.go` | 事件分发（HTTP/WebSocket），按会话工厂创建 Engine |
| BenchmarkRunner | `internal/eval/benchmark.go` | 沙箱隔离执行，ValidateScript 断言验收 |
| Span/Trace | `internal/observability/trace.go` | 树形链路追踪，任务结束导出 JSON |

---

## 怎么跑（详见 `./AGENTS.md` §3）

运行命令、环境变量等详细说明见 `./AGENTS.md` 第 3 节和第 9 节。

**快速命令**：
```bash
go build ./...          # 构建
go test ./...            # 测试
go run ./cmd/claw ...    # CLI
```

---

## 禁区

待补充。

---

## 历史包袱

待补充。