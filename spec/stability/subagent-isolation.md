# 任务委派：引入 Subagent 隔离复杂探索任务的上下文瓶颈

> 状态: 待扩展

---

## 背景与问题

如果大模型在一次请求中同时吐出 3 个 `spawn_subagent` 工具调用，当前的引擎架构能否自动支持这种"多路侦察兵并行出发"的炫酷操作？

---

## 核心方案：非阻塞多兵种异步协同架构

### 1. 架构改造：Subagent 升级为"后台轻量级进程"

**解耦拉起**：
```go
// spawn_subagent 工具的物理行为改变
func (t *SpawnSubagentTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    taskPrompt := args["taskprompt"].(string)
    
    // 瞬间返回，不调用 RunSub 死等
    go func() {
        workerPool.Submit(func() {
            runSubagent(taskPrompt)
        })
    }()
    
    // 生成唯一任务句柄
    pid := generatePID()
    
    return fmt.Sprintf("探路者小队（PID: %s）已成功派往目标，正在后台搜集信息，请稍后用 poll 工具查看其汇报。", pid), nil
}
```

### 2. 调度机制：两种回报模式

**方式 A：主动轮询模式（Proactive Polling）**

```go
// 为主 Agent 补充原生工具
checksubagentstatus(pids []int) -> []SubagentStatus

type SubagentStatus struct {
    PID     string
    State   string  // RUNNING / COMPLETED / FAILED
    Summary string  // 跑完后吐出 Summary 报告
}
```

**方式 B：内核事件驱动（Epoll 模式）**

```
主 Agent 派出 3 个子智能体（PID: 101, 102, 103）
  ↓
主 Agent 进入 epoll_wait 静默挂起状态
  ↓
3 个子智能体在后台并行狂飙
  ↓
谁先跑完，谁触发"就绪中断信号"
  ↓
3 个就绪事件全部集齐（或超时熔断）
  ↓
Harness 唤醒主 Agent，合成大礼包作为 Observation 统一喂给主 Agent
```

### 3. 终极进化：主执行器的"时间片轮转"

```
Turn 1: 处理任务 A，发现需要等审批，挂起任务 A，保存其 contextHistory（内存现场）
Turn 2: 切到任务 B，拉起 spawn_subagent，丢进线程池，挂起任务 B
Turn 3: 检查事件队列，发现任务 C 的子智能体回报了，唤醒任务 C 的上下文，继续往前推进
```

---

## 前沿形态：Team 协作模式

### Blackboard Architecture（黑板架构）

主 Agent（架构师）把大任务拆解后写入共享任务列表，多个专业子智能体（前端、后端、DBA）像"消息订阅者"一样各自认领任务并并行工作。

### P2P Negotiation（点对点协商）

子智能体发现缺少后端 API 接口时，可直接向后端子智能体发起请求：
> "老兄，帮我补个 /user 接口"

前端子智能体协程被挂起，直到后端子智能体完成并返回 Event Signal。

### 群体辩论与一致性投票

极度敏感场景（如线上核心代码 Review）下，同时拉起 3 个不同大模型驱动的 Reviewer Subagent，分别输出意见后互相审查，直到达成多数票共识。

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **Subagent 的生命周期管理**：超时、熔断、资源清理
2. **跨 Subagent 的上下文共享**：P2P 通信的 Channel 设计

### 待扩展2

> 未来规划方向

1. **Subagent 池化复用**：预热常用专长的 Subagent 实例
2. **动态 Agent 团队编排**：根据任务类型自动组合专业子智能体