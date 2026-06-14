# Provider 接口抽象

> 状态: 待扩展

---

## 1. 双协议支持

| Provider | 协议 | 模型示例 |
|----------|------|----------|
| ClaudeProvider | Anthropic | Claude 3.5 Sonnet |
| OpenAIProvider | OpenAI | GPT-4o, GPT-4 |

---

## 2. 接口定义

```go
type LLMProvider interface {
    // 阻塞式生成
    Generate(ctx context.Context, history []schema.Message, systemPrompt string) (*schema.Message, error)

    // 流式生成（待实现）
    GenerateStream(ctx context.Context, history []schema.Message, systemPrompt string) <-chan StreamEvent

    // Token 估算（待实现）
    EstimateTokens(text string) int

    // 获取模型上下文上限（待实现）
    ContextWindowLimit() int
}
```

---

## 3. 属性注入（待实现）

```go
type ProviderConfig struct {
    ModelName           string  // 模型标识
    ContextWindowLimit  int     // 上下文硬上限（如 128000）
    TokenEstimator      func(string) int  // BPE 估算器
}
```

---

## 4. 自适应 Token 估算

- **OpenAI**: 使用 tiktoken
- **开源模型**: 使用本地 BPE 分词器
- **简化估算**: 英文 1 Token ≈ 4 字符，中文 1 Token ≈ 1.5 字符

### 反馈校准

每次 API 返回真实 `usage.PromptTokens` 后，对估算系数进行自我修正。

---

## 5. 待实现细节

- [ ] Provider 接口添加 `ContextWindowLimit()`
- [ ] 实现 `EstimateTokens()` 方法
- [ ] 实现 `GenerateStream()` 流式接口
- [ ] 添加反馈校准机制