# 成本与状态追踪

> 状态: 待扩展

## 模块职责

在 Harness 层拦截并记录 Token 消耗与执行耗时，为成本控制和性能分析提供数据支撑。

---

## 待实现

### 物理工具耗时拦截器方案

#### 1. 结构层：设计工具执行上下文（ToolContext）

为了让拦截器能优雅地拿到当前是哪个工具在执行、参数是什么，我们需要一个轻量级的中间传递对象：

**输入载体**：包含当前要执行的工具名称（ToolName）、大模型给出的原始参数（Arguments）以及全链路的 `context.Context`。

**输出载体**：包含工具执行后的物理输出（Output）以及是否报错的标识（IsError）。

#### 2. 行为层：编写工具级装饰器中间件（Tool Runtime Middleware）

在 `internal/observability` 目录下，设计一个符合 `MiddlewareFunc` 签名的核心拦截器。利用 Go 语言经典的 `defer` 与 `time.Since` 组合，完美实现物理时间夹击：

**拦截器执行核心算法**：

1. **切入前置（Before Execute）**：当 Main Loop 触发某物理工具（如 bash）时，流量首先被拦截器扣下。拦截器立刻在内存中打下一个起始时间戳（`start := time.Now()`）。

2. **流转下发（Next）**：拦截器调用 `next()` 指针，把控制权交还给真正的物理工具（例如让 `bash.go` 去物理编译一个巨型 Go 项目 5 分钟）。此时，拦截器处于同步挂起状态。

3. **切后拦截（After Execute）**：物理工具执行完毕，控制权弹回拦截器。拦截器通过 `time.Since(start)` 瞬间精准算出这次物理长跑到底消耗了多少毫秒（Duration）。

4. **指标落盘与审计**：将工具名、耗时、状态（成功/失败）打包，异步投递给监控系统，或者直接调用当前 `Session.RecordToolDuration(name, duration)` 进行内存累加。

在终端或 Trace 日志里打印一行醒目的审计：
```
[Tracker][Tool-Audit] 🔧 工具 [bash] 执行完毕 | 物理耗时: 302,150 ms | 状态: SUCCESS
```

#### 3. 挂载到 Engine

在 `cmd/claw/main.go` 初始化 ToolRegistry 时，像套娃一样把这个耗时拦截器挂上去：

```go
registry := tools.NewRegistry()
// 注入你设计的可观测性工具拦截器
registry.Use(observability.NewToolDurationTracker())
// 注册那些纯粹的、没有任何监控代码的物理工具
registry.Register(tools.NewBashTool())
registry.Register(tools.NewReadFileTool())
```

**这是最体现"驾驭工程（Harness Engineering）"美学的地方**：bash.go 根本不知道自己被监控了，Main Loop 也不知道工具被拦截了。

---

## 验收标准

- [ ] ToolContext 结构体定义完整
- [ ] MiddlewareFunc 签名符合 Registry 要求
- [ ] `time.Since` 精准计算物理耗时
- [ ] 日志输出格式统一，包含工具名、耗时、状态
- [ ] 内存累加到 Session 的指标中