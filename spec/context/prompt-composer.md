# 提示词组装：动态加载 AGENTS.md 与外挂 Skills

> 状态: 待扩展

---

## 背景与问题

如果一个复杂项目下有 50 个高阶技能包，"渴望式加载（Eager Loading）" 必然导致开局消耗几万 Token。结合 "渐进式暴露（Progressive Disclosure）" 理念，需要重构 SkillLoader 和 Tool Registry。

---

## 核心方案：三阶段渐进式暴露

### 阶段一：Discovery - 极简元数据常驻内存

每个技能被精简为三个核心字段，构建"技能菜单索引表"：

| 字段 | 说明 | 示例 |
|------|------|------|
| Trigger | 什么时候该用它 | 当人类要求提交代码、执行 Git 操作时 |
| Capability | 能解决什么 | 规范化 Git 提交，支持 Emoji 前缀与分步审查 |
| Skill ID | 物理文件路径或唯一标识 | `git-workflow` |

**Token 预算**：50 个技能 × ~20 Token ≈ 1k Token，彻底解决开局爆内存问题。

### 阶段二：Activation - read_skill 原生工具注入

在 ToolRegistry 中外挂高特权内置工具：

```go
// 工具签名
read_skill(skill_id: string) -> SkillBody
```

**System Prompt 物理红线**：
> "你拥有一个专业技能库。当你判定当前任务符合某项技能的场景（Trigger）时，你必须且只能先调用 read_skill(skill_id) 工具来获取该技能的详细执行指南（SOP）。禁止凭空盲目编写无规范的代码。"

**动态挂载流转闭包**：
```
Turn 1: 用户说"帮我把修改提交到 Git"
  → 大模型检索"菜单"，命中 git-workflow
  → 调用 read_skill(skill_id="git-workflow")
  → Harness 拦截，加载 .claw/skills/git-workflow/SKILL.md
  → 作为 Observation 返回给大模型
Turn 2: 大模型正式"习得"技能，严格按 Emoji 规范执行
```

### 阶段三：Thought Anchor - 意识前置

在 Thinking System Prompt（元认知层）强行加入审判逻辑：

> "【第一准则】在为任务制定任何 Plan 之前，你必须进行自省：盘点可用专业技能菜单，检查是否存在与当前任务场景匹配的 Skill？如果有，你的 Plan 的第一步必须是调用 read_skill；如果没有任何技能匹配，你才被允许调用底层原语去自主探索和'造轮子'。"

---

## 文件结构

```
.claw/skills/
└── git-workflow/
    ├── SKILL.md        # 技能正文（按需加载）
    └── METADATA.yaml   # 元数据（常驻内存）
```

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **Tool Registry 集成细节**：read_skill 如何与 Middleware 链配合
2. **多 Skill 并发激活**：大模型同时需要多个技能时的上下文管理
3. **技能版本管理与缓存淘汰策略**

### 待扩展2

> 未来规划方向

1. **跨会话技能记忆**：统计技能使用频次，自动推荐高频技能
2. **技能动态编排**：根据任务类型自动组合多个技能的 SOP