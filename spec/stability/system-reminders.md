# 行为干预：防止 Agent 陷入"死循环"的 System Reminders 机制

> 状态: 待扩展

---

## 背景与问题

大模型可能通过微小差异重试（如加空格、换参数顺序）来逃避严格哈希。如何精准"看穿"这种伪装，捕获"本质上"的死循环？

---

## 核心方案：三步死循环检测与干预

### 步骤一：参数规范化（Normalization）—— 剥离大模型的伪装

**Canonical JSON**：
```go
// 先反序列化为 map，再重新序列化
m := make(map[string]interface{})
json.Unmarshal([]byte(rawJSON), &m)
canonical, _ := json.Marshal(m)  // 键按字母顺序，无多余空格
```

**Path Clean**：
```go
cleaned := filepath.Clean(path)
cleaned = strings.TrimSpace(cleaned)
```

**Command Sanitization**：
```go
// 剔除命令行末尾多余空格、无意义的分号、&& echo ""等刷屏后缀
cmd = strings.TrimRight(cmd, " \t;")
cmd = regexp.MustCompile(`&&\s*echo\s*""?`).ReplaceAllString(cmd, "")
```

### 步骤二：相似性矩阵判断（Similarity）—— 捕捉"本质上"的死循环

**放弃单一哈希，改用"动态轨迹滑动窗口"**：

```go
type ReminderInjector struct {
    history []ToolCall  // 最近 5 轮工具调用滑动历史
}

type ToolCall struct {
    ToolName string
    NormalizedArgs string  // 经步骤一规范化后的参数
}
```

**Levenshtein Distance 相似度算法**：
```
Similarity = 1 - EditDistance(S1, S2) / max(Len(S1), Len(S2))
```

**相似度阈值拦截**：
- 工业级灰度红线：相似度 ≥ 85%
- 连续 3 次工具名称相同 + 规范化后相似度 ≥ 85%
- 判定："本质上陷入了探索螺旋（Doom Loop）"

### 步骤三：强力捕捉与干预（Behavioral Intervention）

**模拟人类"拍肩膀"**：

```go
// 以 Role=User 身份插入硬干预指令，利用近因效应
intervention := Message{
    Role:    RoleUser,
    Content: "[SYSTEM REMINDER 警告] 检测到你正在进行高相似度的无效重试！" +
             "你已经连续 3 次尝试了几乎相同的错误命令。" +
             "请立刻停止猜测参数，彻底改变解题策略，或者直接向用户申请人工帮助！",
}
appendToHistory(history, intervention)
```

---

## 核心数据结构

```go
type DoomLoopDetector struct {
    history       []ToolCall          // 滑动窗口
    windowSize    int                 // 默认 5
    similarityThresh float64          // 默认 0.85
    consecutiveThresh int             // 默认 3
}

type ToolCall struct {
    ToolName       string
    NormalizedArgs string
    Timestamp      time.Time
}
```

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **ReminderInjector 与 Session 的集成**：计数器机制在 Session 层如何实现
2. **多警长协同**：多个高危规则同时触发时的优先级判定

### 待扩展2

> 未来规划方向

1. **自适应阈值**：根据模型特性动态调整相似度阈值
2. **历史模式学习**：从历史会话中学习典型死循环模式