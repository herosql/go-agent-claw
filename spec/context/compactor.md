# 突破内存：基于阶梯降级的 Context Compaction 策略

> 状态: 待扩展

---

## 背景与问题

大模型 API 在每次 Response 时会在 Usage 中返回 PromptTokens。结合这一特性，如何将固定的"字符阈值"拦截改造为"基于真实 API Token 消耗水位线"的 Adaptive Compression？

---

## 现有方案分析

| 方案 | 优点 | 缺点 |
|------|------|------|
| LLM-based Summarization | 最大限度保留关键语义 | API 成本增加，有幻觉遗漏风险 |
| Adaptive Retrieval (Vector DB) | 按需换入相关片段 | 架构复杂度高，延迟增加 |
| Long Context Models | 大力出奇迹 | API 计费高昂，土豪专属 |

---

## 核心方案：自适应压缩三层次

### 1. 结构层：Provider 接口"属性注入"

```go
type LLMProvider interface {
    // 新增字段
    ContextWindowLimit() int      // 物理最大 Token 容量（如 128k）
    ModelName() string            // 表标识（ProviderID）
    
    // 新增方法
    EstimateTokens(text string) int  // Token 估算器
}
```

**Token 估算策略**：
- OpenAI 生态：封装 tiktoken
- 开源模型：本地 BPE 分词器
- 极致低延迟：自适应系数（英文 1 Token ≈ 4 字符，中文 1 Token ≈ 1.5 字符）

### 2. 策略层：动态计算与"自适应防线"

```go
// Compactor 动态向 Provider 询问水位
func (c *Compactor) Compact(ctx context.Context, history []Message, provider LLMProvider) []Message {
    // 动态获取水位
    totalTokens := provider.EstimateTokens(assembleContext(history))
    limit := provider.ContextWindowLimit()
    
    // 安全健康检查
    if totalTokens > limit * 0.8 {
        // 触发降级
        return c.adaptiveTruncation(history, totalTokens, limit)
    }
    return history
}
```

**降级策略**：
- **远期历史**：触发 Observation Masking，将老旧工具返回结果彻底擦除
- **近期记忆**：触发 Head-Tail Truncation，按需精准裁剪而非固定字符数

### 3. 闭环层：利用真实 Usage 账单"反馈校准"

```go
// 影子监控：每次请求后对比估算值与真实值
func (c *Compactor) calibrate(provider LLMProvider, estimated, actual int) {
    if actual > estimated {
        // 估算偏小，调大缩放惩罚系数
        c.penaltyFactor[provider.ModelName()] *= 1.1
    } else {
        c.penaltyFactor[provider.ModelName()] *= 0.95
    }
}
```

**闭环效果**：基于真实账单反馈的自适应算法，精准榨干每寸上下文红利，永不 OOM。

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **Provider 层的 Token 估算实现**： tiktoken / BPE 集成细节
2. **多轮压缩的累积误差控制**：多次压缩后语义损失最小化

### 待扩展2

> 未来规划方向

1. **语义压缩**：不只是截断，而是理解后重写
2. **跨 Provider 统一估算**：不同模型使用同一估算基准