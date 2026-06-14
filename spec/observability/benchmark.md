# 科学度量：Benchmark 自动化评估体系

> 状态: 待扩展

## 模块职责

构建 Benchmark 自动化评估脚本，科学量化 Harness 引擎性能，建立四维一体的"架构调优雷达"。

---

## 评估方法论

### LLM-as-Judge 与 Agent-as-Judge

### 结果评估（Outcome-Based）vs 轨迹评估（Trajectory-Based）

### 子目标完成率

### 持续评估与防过拟合

- **SWE-bench-Live**

### 与 CI/CD 流水线集成

将评估嵌入 CI/CD 流水线，已经成为让 Agent 真正走向生产可信的基础设施。每一次 Prompt 模板变更、工具函数修改或模型版本升级，都自动触发一次 Benchmark 跑分，并对比历史基线，才能真正做到"每次改动心中有数"。

---

## 待实现

### 问题

在当前的 Benchmark Runner 中，如果某个任务非常复杂（例如：重构 10 个代码文件），大模型在我们的 go-tiny-claw 引擎中可能会发生 20 个 Turn 的循环交互，然后才完成任务。

如果在这期间，它有一次 bash 敲错了命令（比如漏了一个路径斜杠），触发了底层报错。但得益于我们在第 14 讲中写的 Error Recovery 机制，它在下一轮自己纠正了错误，并最终通过了 ValidateScript。

对于这种"中途摔了一跤，但最终依然完成任务"的情况，我们目前的计分板只是简单地标记为 `Passed: true`。但在极其苛刻的性能调优（Performance Tuning）中，"一发入魂完成"和"重试了 5 次才完成"，其对系统架构（Prompt、工具设计）好坏的评判价值是完全不同的。

**基于我们之前在第 18 讲（Cost Tracker）和第 19 讲（Tracing 链路追踪）中沉淀的数据收集能力**，如果让你在跑分的 `TestResult` 结构体中，增加两个能够精准度量这种"试错成本"或"驾驭顺滑度"的新指标，你会添加哪两个关键指标？为什么？

### 方案

#### 1. 数据结构拓扑（Schema 升级）

在 `TestResult` 结构体中，彻底废弃单薄的 `Passed` 和 `TotalCost`，演进为以下工业级审计单：

```go
type TestResult struct {
    TestCaseID             string
    Passed                 bool    // 物理断言：ValidateScript 是否 exit 0
    TotalTurns             int     // 指标 1：主循环 ReAct 迭代次数
    ToolRetryCount         int     // 指标 2：Error Recovery 触发计数
    TotalPromptTokens      int64   // 指标 3.1：累积输入 Token (反映内存压缩 GC 效率)
    TotalCompletionTokens  int64   // 指标 3.2：累积输出 Token (反映慢思考推理密度)
    TotalCostUSD           float64 // 综合财务账单
}
```

#### 2. 拦截器流式打账（数据收集机制）

- **Turns 计数**：在手写的 Main Loop 顶层计数器累加，每经历一次 `Thinking -> Action -> Execution`，`TotalTurns++`。
- **Retry 计数**：挂载在第 14 讲的 RecoveryManager 内部。一旦触发了 `switch-case` 匹配或私有化小模型的 Summary 注入，立即给当前 Session 的 `ToolRetryCount++`。
- **Token 动态累加**：利用第 18 讲实现的 Harness 拦截器（Middleware），在每次 Provider 发起 HTTP 请求的 Response 返回体中，流式抽取出 Usage 数据，同步追加到当前测试例的 Token 账本中。

#### 3. 终极防线：四维一体的"架构调优雷达"

当你在本地或公司内网运行 Benchmark 自动化跑分脚本时，每一项改动都将通过下面这张雷达图进行物理审判，直接导出核心诊断方向：

| 模式 | 特征 | 诊断 | 药方 |
|------|------|------|------|
| **A. 笨拙挣扎型** | 高 Turns + 高 Retry | 频繁撞南墙，靠自愈机制死撑活拉 | 必须重构工具的 Description，别让模型盲猜参数；或使用模糊匹配降低工具报错率 |
| **B. 唐僧念经型** | 高 Turns + 低 Retry | 工具调用一发入魂，但 Completion Tokens 极其巨大。模型在 Thinking 阶段疯狂自己跟自己碎碎念，迟迟不肯退出 | 削减 System Prompt 里多余的自省规则，加大对 Thinking 阶段的"物理催促"，限制最大生成长度 |
| **C. 上下文肥胖症** | 低 Turns + 高 Input Token | 步数很少，但 Prompt Tokens 像滚雪球一样失控 | 滑动窗口太宽，或 Context Compactor 阶梯降级策略卡得不够狠，必须强行擦除远期的巨型工具报错日志 |

#### 4. 终极闭环：接入自动化 CI/CD 质量红线

这套方案最大的商业价值，在于将其彻底降维打击为传统的、确定性的"全自动软件压测"：

在项目的 GitHub Actions 或 GitLab CI 流水线中挂载该跑分脚本。每当团队有人修改了 AGENTS.md 提示词或合入新工具代码时，CI/CD 强行跑一遍 Benchmark：

- 如果物理断言失败（`Passed == false`），直接拒绝合入。
- 如果断言通过，但只要 `TotalTurns` 或 `TotalCostUSD` 相比于 Main 主分支的基线（Baseline）产生了超过 15% 的异常劣化（Regression），流水线同样触发物理熔断，打回重写。

---

## 验收标准

- [ ] `TestResult` 结构体包含 `TotalTurns` 和 `ToolRetryCount` 指标
- [ ] Middleware 流式采集 Token 消耗
- [ ] RecoveryManager 触发时累加 `ToolRetryCount`
- [ ] 四维雷达图诊断逻辑实现
- [ ] CI/CD 流水线质量红线集成