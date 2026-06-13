---
name: docs-auto-sync
description: 当代码发生变化后，自动检查并报告哪些文档需要同步更新。适用于代码重构、接口变更、架构调整后，确保文档与实现保持一致。只读不写，不自动修正。
---

# 文档同步检查技能 (docs-auto-sync)

## 触发条件

当用户报告以下场景时，强制加载此技能：
- 代码重构后文档可能过时
- 架构图或接口发生变更
- 用户明确要求"检查文档一致性"
- 新增/删除/修改了 `cmd/`、`internal/` 下的模块

## 执行步骤

### 第一步：确定代码变更范围

1. 使用 `git diff --name-only HEAD~5` 确认最近 5 次提交中改动的文件
2. 使用 `git log --oneline -10` 了解最近的变更背景
3. 分类改动：
   - `cmd/` 下新增/删除入口 → 必须检查 `docs/api-list.md` 入口部分
   - `internal/` 下新增/修改接口类型（interface 定义）→ 必须检查 `docs/api-list.md` 内部接口部分
   - `internal/schema/` 下修改结构体 → 必须检查 `docs/data-model.md`
   - 新增/删除/重命名 `internal/` 子包 → 必须更新 `docs/module-deps.mmd`
   - 新增外部依赖（go.mod 变更）→ 必须更新 `docs/external-deps.mmd`
   - 新增工具（tools 包新文件）→ 检查 `docs/architecture.mmd` 工具层

### 第二步：读取最新代码实现

按以下优先级读取变更文件：

**接口变更（高优先级）**：
- 读取 `internal/provider/interface.go` — LLMProvider 接口签名
- 读取 `internal/tools/registry.go` — Registry 接口签名
- 读取 `internal/engine/reporter.go` — Reporter 接口签名
- 读取 `internal/tools/subagent.go` — AgentRunner 接口签名
- 读取 `internal/feishu/bot.go` — FeishuBot 工厂接口

**数据模型变更（中优先级）**：
- 读取 `internal/schema/messge.go` — Message、ToolCall、ToolResult、ToolDefinition
- 读取 `internal/context/session.go` — Session 结构体
- 读取 `internal/observability/trace.go` — Span 结构体
- 读取 `internal/feishu/approval.go` — ApprovalManager 结构体

**入口变更（低优先级）**：
- 读取 `cmd/*/main.go` — 各入口启动逻辑

### 第三步：逐项对比检查

对每个变更的模块，回答以下问题并在报告中说明：

| 检查项 | 检查方法 | 通过标准 |
|---|---|---|
| 接口签名一致 | 读取 .go 接口定义 vs `docs/api-list.md` 接口段落 | 接口名、参数、返回值类型一致 |
| 数据模型字段一致 | 读取 .go 结构体定义 vs `docs/data-model.md` 字段列表 | 字段名、类型、说明一致 |
| 模块依赖关系一致 | `go list -f '{{.ImportPath}} → {{join .Imports ","}}' ./internal/...` vs `docs/module-deps.mmd` | import 关系与图一致 |
| 外部依赖一致 | `cat go.mod` vs `docs/external-deps.mmd` | 直接依赖（不含间接）与图一致 |
| CLI 入口命令一致 | 读取 cmd/*/main.go flag 定义 vs `docs/api-list.md` CLI 段落 | flag 名、默认值、说明一致 |

### 第四步：生成报告

报告格式：

```
## 文档同步检查报告

**检查时间**：[时间]
**代码变更范围**： [git diff 列表]

### ✅ 一致的部分
（列出文档与代码一致的模块）

### ⚠️ 需要更新的部分
（列出不一致项，格式：）
- [文件] 第 X 行：[文档描述] vs [实际代码描述]
  建议：[应改为...]

### 📋 待人工确认
（列出无法自动判断、需要业务知识的地方）
```

## 约束

- **只读不写**：禁止使用 Write/Edit 工具修改任何文件
- **不自动修正**：仅报告问题，不执行修复
- **不臆测**：如果无法确认某项，使用 Grep 在整个代码库搜索佐证
- **优先代码**：文档与代码不一致时，以代码实际行为为准，在报告中标注"代码优先"