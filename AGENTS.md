# go-agent-claw AI Agent 协作指南

> 适用：所有 AI Agent，优先级低于 CLAUDE.md

## 1. 开发环境

- **语言**：Go >= 1.21（建议最新稳定版）
- **核心依赖**：
  - 智谱 GLM（通过 `ZHIPU_API_KEY` 环境变量注入）
  - 飞书 SDK（`lark/lark`）
  - 其他：`net/http`、`encoding/json`、`log/slog`
- **依赖管理**：`go mod tidy` / `go get <package_path>`
- **构建**：`go build ./...`

## 2. Go 标准工程布局

```
go-agent-claw/
├── cmd/                      # 可执行入口
│   ├── claw/                 # CLI 主程序
│   ├── feishu/               # 飞书 HTTP Bot
│   ├── agentops/             # 飞书 WebSocket Bot（含审批流）
│   ├── bench/                # Benchmark 评测
│   └── envcheck/              # 环境检查工具
├── internal/                 # 私有代码（不可被外部 import）
│   ├── engine/                # Agent 引擎（两阶段执行循环）
│   ├── tools/                 # 工具注册与 Middleware 链
│   ├── provider/              # LLM Provider 接口
│   ├── context/               # 会话、Compactor、PromptComposer
│   ├── feishu/                # 飞书 Bot、Approval、Bot WebSocket
│   ├── eval/                  # Benchmark Runner
│   └── observability/         # CostTracker、Trace、Span
├── docs/                      # 架构图、接口文档
├── scripts/                   # 辅助脚本
└── .claude/                   # Claude Code 配置
```

**约束**：
- 所有业务逻辑必须在 `internal/` 下，禁止在根目录放置业务代码
- `cmd/` 下每个子目录是一个独立可执行程序
- 外部依赖仅通过 `internal/` 子包间接引入

## 3. 运行命令

```bash
# CLI（必需 ZHIPU_API_KEY）
go run ./cmd/claw -prompt "任务" -dir . -session id

# 飞书 HTTP Bot（需 FEISHU_APP_ID/SECRET）
FEISHU_APP_ID=x FEISHU_APP_SECRET=x ZHIPU_API_KEY=x go run ./cmd/feishu

# 飞书 WebSocket Bot（带审批流）
FEISHU_APP_ID=x FEISHU_APP_SECRET=x ZHIPU_API_KEY=x go run ./cmd/agentops

# Benchmark 评测
go run ./cmd/bench

# 环境检查
go run ./cmd/envcheck
```

## 4. 测试与质量

- **全量测试**：`go test -v ./...`
- **单包测试**：`go test -v ./internal/<包名>`
- **覆盖率**：`go test -v -cover ./...`
- **代码检查**：`golangci-lint run`

## 5. Git 规范

- **分支**：`feature/<功能名>` / `fix/<问题名>`
- **Commit**：遵循 Conventional Commits，格式 `type(scope): 描述`
  - `feat`：新功能
  - `fix`：修复
  - `docs`：文档
  - `refactor`：重构
  - `test`：测试
- **PR 标题**：`[PROJ-XXX] type(scope): 描述`
- **提审前**：必须通过 `go build ./...` 和 `go test ./...`

## 6. 架构约定

- **接口设计**：小接口优先，消费者定义
- **错误处理**：`fmt.Errorf("上下文: %w", err)` 或 `errors.New`
- **日志**：使用 `log/slog` 结构化日志
- **并发安全**：必须标注并解释同步措施

## 7. AI 协作规则

- **新功能**：先读代码 → 提计划 → 确认后编码
- **复杂逻辑**：必须附核心逻辑注释
- **测试**：优先表格驱动测试（table-driven tests）
- **上下文压缩**：超 200k 字符时触发 Compactor，详见 `internal/context/compactor.go`
- **编码前先思考**：明确陈述假设，不确定时先问；存在多种解释时先呈现，不要默默选择；发现更简单方案时说出来
- **简单性优先**：最小化代码解决问题，不做投机性设计；不实现 spec 之外的功能；200 行能 50 行搞定则重写
- **精准修改**：只动必须动的；不"改进"相邻代码；匹配现有风格；修改导致的无用代码才删除，不动已有的死代码
- **目标驱动执行**：将任务转化为可验证目标；多步骤任务先陈述计划再执行；强成功标准支撑独立循环

## 8. 关键设计提醒

- **ReminderInjector**：MD5 指纹检测连续失败，超 3 次注入打断指令防死循环
- **Compactor**：上下文超阈值时对历史区全量掩码、对保护区掐头去尾
- **RecoveryManager**：报错特征匹配注入自救指南
- **CostTracker**：Provider 装饰器，计量 Token 成本

详见 `docs/module-deps.mmd`（模块依赖图）和 `docs/external-deps.mmd`（外部依赖图）。

## 9. 环境变量

| 变量 | 必需 | 说明 |
|------|------|------|
| `ZHIPU_API_KEY` | 是 | 智谱 GLM API Key |
| `FEISHU_APP_ID` | 飞书 Bot | 飞书应用 App ID |
| `FEISHU_APP_SECRET` | 飞书 Bot | 飞书应用 App Secret |

详见 `docs/api-list.md` 完整清单。