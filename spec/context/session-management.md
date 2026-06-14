# 会话管理：Session 物理隔离与 Working Memory

> 状态: 待扩展

---

## 背景与问题

引擎全局共享一个 Registry 实例，绑定全局唯一的 WorkDir。业务若需让同一引擎进程服务多个互不干扰的项目目录，必须通过 context.Context 将当前 Session 的动态 WorkDir 透传给工具。

---

## 核心方案：三阶段上下文压缩

### 第一阶段：纵向裁剪 —— 窗口粗筛（按条数）

保留朴素滑动窗口机制，从后往前截取最近 $N$ 条消息。

- **目的**：快速缩小战线，保留最新、最相关的连续剧情
- **边界对齐**：应用"孤儿工具响应剔除逻辑"，若截断后首条是 ToolResult，顺延舍弃直到首条是 User 或 Assistant 消息

### 第二阶段：横向挤压 —— 单条长消息"外科手术"（按长度/Token）

对粗筛后的 $N$ 条消息内部进行扫描与横向压缩：

**动态检测与分级处理**：
- 设定单条消息预估大小阈值 $M$（如 10k Token）
- 遍历发现超越阈值的长消息，立即触发智能无损压缩：

| 策略 | 适用场景 | 方式 |
|------|----------|------|
| 中间塌陷（Middle-Out） | 长日志、长代码 | 保留头部 2000 字 + 尾部 2000 字，中间用掩码替换 |
| 结构化大纲化（Summary Fallback） | 重要长消息 | 调用廉价小模型提取 Metadata 大纲替换正文 |

### 第三阶段：最终合规验证 —— 总体容量"死刑宣判"（安全水位线）

- **计算总预算**：累加 System Prompt + $N$ 条精炼消息的总 Token
- **安全水位线**：检查总 Token 是否小于 `ContextWindowLimit * 0.8`
- **自适应降级（Loop Truncation）**：
  - 符合 → 完美放行
  - 超标 → 自动将条数限制由 $N$ 降为 $N-2$，剔除较老消息
  - 循环执行直到总体积降到水位线以下

---

## 核心数据结构

```go
type Session struct {
    ID           string
    WorkDir      string                    // 动态 WorkDir，通过 Context 透传
    History      []Message                 // 全量历史
    WorkingMemory func() []Message         // 最近 N 轮
    Compactor    *Compactor                // 压缩器
}

type Message struct {
    Role    Role
    Content string
    Tokens  int  // 预估 Token 数
}
```

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **Context 透传机制**：BaseTool.Execute 中如何从 Context 提取 Session WorkDir
2. **多项目并发隔离**：多个 Session 同时运行时的资源隔离策略

### 待扩展2

> 未来规划方向

1. **会话迁移与恢复**：Session 快照与断点续传
2. **跨会话上下文共享**：多个项目间的公共知识复用