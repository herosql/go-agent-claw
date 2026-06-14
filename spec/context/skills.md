# Skill 渐进式加载

> 状态: 待扩展

---

## 1. 渴望式加载 vs 渐进式暴露

### 问题

如果一次加载 50 个技能包，消耗几万个 Token。

### 解决：渐进式暴露

启动时只加载技能的元数据（YAML），正文按需加载。

---

## 2. 极简元数据常驻内存

```go
type SkillMeta struct {
    ID          string  // 技能 ID
    When        string  // 触发场景描述
    Capability  string  // 能力说明
    FilePath    string  // 技能文件路径
}
```

System Prompt 中只包含技能"菜单"：

```markdown
## 可用技能
- git-workflow: 当需要提交代码或执行 Git 操作时使用
- code-review: 当需要审查代码时使用
...
```

Token 消耗：50 个技能 ≈ 1k Token

---

## 3. read_skill 原生工具

```go
// 内置特权工具
type ReadSkillTool struct{}

func (t *ReadSkillTool) Name() string {
    return "read_skill"
}

func (t *ReadSkillTool) Execute(ctx context.Context, skillID string) (string, error) {
    // 加载 .claw/skills/{skillID}/SKILL.md
    content, err := loadSkillFile(skillID)
    if err != nil {
        return "", err
    }
    return content, nil
}
```

---

## 4. 动态挂载的流转闭包

```
Turn 1: 用户说"帮我提交代码"
  → 大模型检索菜单，发现命中 git-workflow
  → 调用 read_skill(skillID="git-workflow")
  → Harness 加载技能正文并返回

Turn 2: 大模型"习得"技能，正式执行 git commit
```

---

## 5. 意识前置（Thought Anchor）

在 Thinking 的 System Prompt 中：

```markdown
【第一准则】在制定任何 Plan 之前，必须进行自省：
1. 盘点可用技能菜单
2. 检查是否存在匹配技能

如果有，第一步必须是调用 read_skill；
如果没有，才允许使用底层原语（write_file/bash）
```

---

## 6. 待实现细节

- [ ] SkillMeta 结构与 YAML 解析
- [ ] `read_skill` 工具实现
- [ ] 技能菜单的 System Prompt 组装
- [ ] 意识前置提示词注入