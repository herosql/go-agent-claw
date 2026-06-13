# go-agent-claw

AI Agent Harness（驾驭引擎），通过工具调用（Tool Use）驱动大模型在工作区执行代码任务。

## 核心能力

- **CLI 交互**：通过 prompt 驱动 Agent 在指定工作区执行代码任务
- **飞书 Bot**：HTTP Webhook / WebSocket 两种模式，支持审批流
- **自动化评测**：Benchmark 沙箱隔离执行，ValidateScript 断言验收

## 核心架构

详见 `docs/architecture.mmd`（分层全景图）。

**两阶段执行循环**（`internal/engine/loop.go`）：
1. **Thinking Phase**（可选）：模型生成推理，Content 拼入上下文
2. **Action Phase**：模型决定调用工具，并行执行后结果注入上下文，循环直到无 ToolCall

**关键设计**：
- `ReminderInjector`：MD5 指纹检测连续失败，超 3 次注入打断指令防死循环
- `Compactor`：上下文超 200k 字符时对历史区全量掩码、对保护区掐头去尾
- `RecoveryManager`：报错特征匹配注入自救指南
- `CostTracker`：装饰器拦截 Usage，按模型定价表累加账单

## 快速开始

```bash
# 构建
go build ./...

# CLI
ZHIPU_API_KEY=xxx go run ./cmd/claw -prompt "帮我改个 bug" -dir .

# 飞书 Bot（HTTP）
ZHIPU_API_KEY=xxx FEISHU_APP_ID=xxx FEISHU_APP_SECRET=xxx go run ./cmd/feishu

# 飞书 Bot（WebSocket + 审批）
ZHIPU_API_KEY=xxx FEISHU_APP_ID=xxx FEISHU_APP_SECRET=xxx go run ./cmd/agentops

# 评测
ZHIPU_API_KEY=xxx go run ./cmd/bench
```

## 项目结构

```
cmd/
  claw/        CLI 主程序
  feishu/      飞书 HTTP Bot
  agentops/    飞书 WebSocket Bot（带审批流）
  bench/       Benchmark 评测
  envcheck/    环境变量诊断

internal/
  engine/      两阶段执行循环
  tools/       工具注册与 Middleware 链
  provider/    LLM Provider 接口（OpenAI/Claude 双协议）
  context/     Session、Compactor、PromptComposer、RecoveryManager
  feishu/      飞书 Bot、ApprovalManager
  eval/        BenchmarkRunner
  observability/ CostTracker、Span 链路追踪

docs/
  architecture.mmd   分层架构图
  module-deps.mmd    模块依赖图
  external-deps.mmd  外部依赖图
  api-list.md        接口清单
  data-model.md     数据模型
```

## 环境变量

| 变量 | 必需 | 说明 |
|---|---|---|
| `ZHIPU_API_KEY` | 是 | 智谱 GLM API Key |
| `FEISHU_APP_ID` | 飞书 Bot | 飞书应用 App ID |
| `FEISHU_APP_SECRET` | 飞书 Bot | 飞书应用 App Secret |

详见 `docs/api-list.md`。

## 工程规范

详见 `AGENTS.md`（项目协作规范）和 `constitution.md`（项目宪法）。