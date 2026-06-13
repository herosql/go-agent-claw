# 核心链路清单

共 8 条，覆盖 Agent 引擎、工具层、飞书审批、上下文管理、可观测性。

---

## 链路 1：CLI Agent 执行循环

```
用户 -prompt → cmd/claw
  → GlobalSessionMgr.GetOrCreate(sessionID, workDir)
  → AgentEngine.Run(ctx, sess, TerminalReporter)
    → PromptComposer.Build() → System Prompt
    → for turn loop:
        → Compactor.Compact(contextHistory)   [链路2]
        → LLMProvider.Generate(thinking)     [EnableThinking=true]
        → LLMProvider.Generate(action)        [工具调用决策]
        → Registry.Execute(toolCall)         [链路3]
        → ReminderInjector.CheckAndInject()  [死循环检测]
        → Observation 注入 Session
    → Session 历史追加 + 成本累加
  → 输出最终回复
```

**关键文件**：`internal/engine/loop.go`、`cmd/claw/main.go`
**核心断言点**：Thinking 和 Action 都生成 ToolCall → 最后无 ToolCall 时退出

---

## 链路 2：上下文压缩（Compactor）

```
Compactor.Compact(msgs []Message)
  → 估算总字符数
  → 未超 200k：直接返回原数组
  → 超阈值：
      → System Prompt：原样保留
      → 历史区（< 保护区）：超 200 字符 → 全量掩码 "...[折叠]..."
      → 保护区：单条超 1000 字符 → 掐头去尾各 500 字符
      → Assistant Thinking：历史区超 200 字符 → 折叠
  → 返回压缩后数组
```

**关键文件**：`internal/context/compactor.go`
**核心断言点**：压缩后总字符数 < MaxChars；ToolCalls 字段不被修改

---

## 链路 3：工具执行链（Registry + Middleware）

```
Registry.Execute(ctx, ToolCall)
  → StartSpan("Tool.Execute")
  → 路由查找 BaseTool
  → 依次执行 MiddlewareFunc 链
    → 任一返回 allowed=false → 返回 IsError=true + rejectReason
  → BaseTool.Execute(ctx, args)
    → read_file: 限制 workDir + 8k 截断
    → write_file: 自动 mkdir + workDir 限制
    → edit_file: 4级模糊匹配替换   [链路4]
    → bash: 30s 超时 + 8k 截断
    → spawn_subagent: 调用 AgentEngine.RunSub   [链路5]
  → EndSpan + 返回 ToolResult
```

**关键文件**：`internal/tools/registry.go`
**核心断言点**：危险工具（write_file/edit_file）在 Middleware 层被拦截；正常工具返回 IsError=false

---

## 链路 4：编辑文件四级模糊匹配

```
edit_file.fuzzyReplace(content, oldText, newText)
  → L1 精确匹配: strings.Count == 1 → 直接 Replace
  → L2 归一化: \r\n → \n 再匹配
  → L3 TrimSpace: 去除首尾空白后匹配
  → L4 逐行去缩进: 滑动窗口按行匹配
    → 匹配 0 处 → error "在文件中未找到 old_text"
    → 匹配 >1 处 → error "模糊匹配到 N 处"
    → 匹配 1 处 → 执行替换
```

**关键文件**：`internal/tools/edit_file.go`
**核心断言点**：L1/L2/L3/L4 各层独立降级；精确匹配唯一时绝不触发模糊匹配

---

## 链路 5：子智能体委派

```
spawn_subagent.Execute(ctx, args)
  → 解析 TaskPrompt
  → AgentEngine.RunSub(ctx, prompt, readOnlyRegistry, reporter)
    → 构建 Explorer System Prompt（强制只读）
    → for turn ≤ 10:
        → LLMProvider.Generate(action)  [关闭 Thinking]
        → readOnlyRegistry.Execute()     [仅 read_file / bash]
        → 无 ToolCall → 返回 Content 作为 Summary
    → 超 10 轮 → error "探索过于深入"
  → 返回 【子智能体探索报告】: <summary>
```

**关键文件**：`internal/tools/subagent.go`、`internal/engine/loop.go:RunSub`
**核心断言点**：readOnlyRegistry 不含 write_file/edit_file；子智能体不返回 ToolCall 时退出

---

## 链路 6：飞书审批流程

```
cmd/agentops 启动
  → FeishuBot.GetEventDispatcher() 注册事件处理
  → larkws.NewClient() 建立 WebSocket 长连接

飞书收到用户消息
  → WebSocket 推送至 OnP2MessageReceiveV1
  → "approve <taskID>" → GlobalApprovalMgr.ResolveApproval(taskID, true, ...)
  → "reject <taskID>"  → GlobalApprovalMgr.ResolveApproval(taskID, false, ...)
  → 普通文本 → handleAgentRun(chatId, prompt)
      → factory(sess) 创建 Engine + CostTracker
      → eng.Run(ctx, sess, FeishuReporter)

Registry.Middleware 检测到危险操作（write_file / edit_file / 危险 bash）
  → ApprovalManager.WaitForApproval()
      → 创建容量 1 的 Channel
      → 挂起当前 Goroutine
      → 向飞书发送审批请求卡片
  → 等待 Channel 收到结果
  → allowed=true → 放行；allowed=false → 返回 IsError=true
```

**关键文件**：`internal/feishu/approval.go`、`internal/feishu/bot.go`、`internal/feishu/bot_websocket.go`
**核心断言点**：WaitForApproval 在收到 resolve 前持续阻塞；resolve 唤醒后正确传递 allowed 值

---

## 链路 7：成本追踪（CostTracker）

```
CostTracker.Generate(ctx, msgs, tools)
  → 记录 startTime
  → 调用 nextProvider.Generate()   [真实 LLM 调用]
  → 计算 latency
  → 解析 respMsg.Usage（PromptTokens + CompletionTokens）
  → 从 PricingModel 查找 modelName 对应单价
  → cost = (prompt * inputPrice + completion * outputPrice) / 1_000_000
  → session.RecordUsage(prompt, completion, cost)
  → 累加 session.TotalCostCNY / TotalPromptTokens / TotalCompletionTokens
  → 返回 respMsg
```

**关键文件**：`internal/observability/tracker.go`
**核心断言点**：Usage 为 nil 时不计费；成本累加为 Session 级别（非全局）

---

## 链路 8：会话管理与断点续传

```
GlobalSessionMgr.GetOrCreate(sessionID, workDir)
  → 查 map[sessions]
  → 不存在 → NewSession() → 存入 map
  → 存在 → 直接返回

Session.GetWorkingMemory(limit)
  → 复制最后 limit 条消息
  → 处理截断边缘 ToolResult 孤儿（ToolCallID 非空的 User 消息跳过）
  → 返回 []Message

PlanMode PromptComposer.Build()
  → 检查 PLAN.md + TODO.md 是否存在
  → 不存在 → 强制创建，写入架构设计
  → 存在 → 读取并注入 System Prompt
  → 强制 Agent 每步打勾（edit TODO.md - [ ] → - [x]）
```

**关键文件**：`internal/context/session.go`、`internal/context/composer.go`
**核心断言点**：ToolResult 孤儿在截断时被正确跳过；PlanMode 下 PLAN.md / TODO.md 必须存在