# 慢思考与自省：Thinking 阶段

> 状态: 待扩展

---

## 1. 背景问题

如果模型在 Phase 1 想出的计划本身就是荒谬或错误的，直接在 Phase 2 执行会导致失败。

---

## 2. 审查微循环设计

### 独立审查提示词（Critic System Prompt）

```markdown
Role: 你是极度严苛的"Agent 行为审计专家"。你的职责是揪出 Agent 计划中的逻辑漏洞、幻觉以及安全隐患。

Input:
1. 用户的核心目标 (Root Task)
2. Agent 刚刚生成的执行计划 (Proposed Plan)

Audit Checklists:
1. 逻辑闭环性：计划中的每一步是否能合逻辑地推导到核心目标？
2. 幻觉检查：计划中提及的文件、参数或工具，是否凭空捏造？
3. 重复踩坑：该计划是否包含之前已经失败过的尝试？

Output Format (严格 JSON):
{
  "approved": true 或 false,
  "reason": "如果拒绝，写明具体的逻辑漏洞或安全隐患；如果通过，写 PASS"
}
```

### 审查微循环代码（参考实现）

```go
// internal/engine/loop.go
// ====================================================================
// Phase 1: 慢思考阶段 (Thinking)
// ====================================================================
var thinkResp *schema.Message
maxRetries := 3

for i := 0; i < maxRetries; i++ {
    // 1. 让模型生成初步计划
    thinkResp, err = e.provider.Generate(ctx, contextHistory, nil)
    if err != nil {
        return fmt.Errorf("Thinking 阶段生成失败: %w", err)
    }

    // 2. 物理插桩：将初步计划丢进"独立审查微循环"
    approved, reason := e.EvaluatePlan(ctx, userPrompt, thinkResp.Content)
    if approved {
        log.Println("[Engine][Critic] ✅ 计划通过审查")
        contextHistory = append(contextHistory, *thinkResp)
        break
    }

    // 拒绝理由临时注入，逼迫模型修正逻辑
    contextHistory = append(contextHistory, schema.Message{
        Role:    schema.RoleUser,
        Content: fmt.Sprintf("【系统拒绝提示】你刚才的计划存在以下漏洞：%s。请重新评估并给出一份更严谨的计划。", reason),
    })
}
// ====================================================================
// Phase 2: 行动阶段 (Action)
// ====================================================================
```

---

## 3. 自适应推理（Adaptive Reasoning）

- **简单任务**（如列出目录文件、查天气）：关闭 Thinking，享低 Token 成本
- **复杂任务**（如分析依赖关系并重构）：打开 Thinking，用算力换准确性

---

## 4. 流式响应支持（Streaming）

### StreamEvent 结构

```go
type StreamEvent struct {
    ContentDelta    string           // 文本增量
    ToolCallDeltas   []ToolCallDelta // 工具调用增量
    Error            error           // 传输错误
}

type ToolCallDelta struct {
    Index          int
    ToolCallID     string
    ToolName       string
    ArgumentsDelta string
}
```

### Provider 接口改造

```go
// 返回只读通道，实现异步流式
func (p *ClaudeProvider) GenerateStream(ctx context.Context, ...) <-chan StreamEvent {
    ch := make(chan StreamEvent)
    go func() {
        // 异步拉取 SSE，灌入 channel
        defer close(ch)
    }()
    return ch
}
```

### Main Loop 接收器

```go
// 先构造，后填充，终校验
msg := &schema.Message{}
buffers := map[int]*strings.Builder{}

for event := range stream {
    if event.ContentDelta != "" {
        fmt.Print(event.ContentDelta) // 打字机效果
        buffers["content"].WriteString(event.ContentDelta)
    }
    if event.ToolCallDeltas != nil {
        for _, delta := range event.ToolCallDeltas {
            buffers[delta.Index].WriteString(delta.ArgumentsDelta)
        }
    }
}
```

---

## 5. 待实现细节

- [ ] EvaluatePlan 审查微循环实现
- [ ] 独立审查 Session 的创建与销毁
- [ ] EnableThinking 硬开关
- [ ] 流式响应的完整实现