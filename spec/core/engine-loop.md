# 核心心脏：Agent Main Loop

> 状态: 待扩展

---

## 1. 架构演进：从 Framework 到 Harness

### 存储层级策略

当 LLM Context Window 逼近极限（如 128k Tokens）时，采取类似 OS 的策略：

| 存储层级 | 对应策略 | 存放内容 | Token 水位控制 |
|----------|----------|----------|----------------|
| L1 寄存器 (极速) | 滑动窗口截断 | 最近 3~5 轮的原始、精确的对话与工具调用日志 | 保持在 10k Token 以内 |
| L2 缓存 (常驻) | 低级别模型总结 | 实时更新的 PLAN.md 和 TODO.md | 保持在 2k Token 左右 |
| L3 远端磁盘 (冷备) | RAG 向量/文件归档 | 超过 20 轮的所有历史原始对话 | 0（不占模型当前窗口） |

---

## 2. ReAct 原理

参考: [https://arxiv.org/pdf/2210.03629](https://arxiv.org/pdf/2210.03629)

ReAct = Reason + Act，让大模型交替进行推理和行动。

---

## 3. 部分失败（Partial Failure）的弹性吸收

### 问题

假设并发执行 5 个文件读取，其中第 2 个文件因为权限问题报错了，另外 4 个成功了。

### 方案

**关键：绝对不要让一个工具报错导致整个引擎崩溃**

1. **标识只读工具**：在 ToolDefinition 中加入 `IsReadOnly bool` 标识
2. **并发审查**：纯只读工具（如 Read/Grep）可并发执行；修改工具（Edit/Bash）串行执行
3. **错误封装**：报错协程依然返回 `IsError: true` 和报错堆栈
4. **聚合返回**：`WaitGroup.Wait()` 结束后，Harness 带着 4 成功 + 1 失败结果一起喂给大模型
5. **自愈机制**：大模型拿到结果后会反思："第 2 个文件没权限，换个思路"

### 伪代码

```go
// 并发执行，但修改工具串行
for _, toolCall := range toolCalls {
    if toolCall.IsReadOnly {
        go executeTool(toolCall, resultChan)
    } else {
        executeToolSerially(toolCall)  // 修改工具串行
    }
}

// 收集结果（部分失败可接受）
waitGroup.Wait()
allResults := collectResults(resultChan)
```

---

## 4. 待实现细节

- [ ] 实现 L1/L2/L3 存储层级的具体代码
- [ ] ToolDefinition 添加 `IsReadOnly` 字段
- [ ] Dispatcher 层的读写分流逻辑
- [ ] 错误结果的标准化封装