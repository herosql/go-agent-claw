# 洞察黑盒：链路追踪机制

> 状态: 待扩展

## 模块职责

为 Agent 引入 Tracing 机制，复盘失败决策路径，将运行时链路追踪与因果/认知层追踪结合。

---

## 工程界的 Agent Trace 思路与方法

### 面向 LLM 的"增强型 Trace"

LangSmith（LangChain 生态）等平台在 Span Tree 的基础上进一步扩展，将 Token 消耗、Prompt 版本、模型参数、评分反馈等 LLM 特有的元数据融入每个 Span，形成面向 LLM 的"增强型 Trace"。

### 多 Agent 协作追踪

在 Multi-Agent 系统（如 AutoGen、CrewAI）中，单棵 Trace Tree 已不足以描述跨 Agent 的消息传递。工程界的做法是引入 **Distributed Trace**，为每条跨 Agent 消息注入 `trace_id` + `parent_span_id`，将多个 Agent 的独立树拼接为一张有向无环图（DAG）。

### 异步与并行 Span

当 Agent 并发调用多个工具时，子 Span 的时间区间会出现重叠。现代框架会在可视化层将这类并行 Span 渲染为甘特图（Gantt-style），而非串行树，以便直观定位延迟瓶颈。

---

## 待实现

### 问题

我们目前实现的这个 `ExportTraceToFile`，仅仅是生成了一个庞大而扁平的 JSON 文件。虽然它是结构化的，但如果你的 Agent 跑了一天，这个 JSON 文件可能会有几十 MB，用文本编辑器打开肉眼阅读极其困难。

业界在处理微服务的链路追踪时，通常会使用标准的 **OpenTelemetry (OTel) 协议**，并将数据上报给 **Jaeger** 或 **Zipkin** 这样的前端可视化看板。

**结合你对云原生监控体系的理解**，如果要把我们现有的 Span 数据结构转换并发送到 Jaeger 系统中，使我们能够在浏览器里看到极其漂亮的"甘特图（Gantt Chart）"（横向时间轴展示 LLM 推理和并发工具执行的时间重叠），你认为在代码架构上我们需要如何扩展当前的 observability 包？

### 方案：流式时间分片 Trace 架构（Segmented Stream Tracing）

#### 1. 结构解耦：从"内存树"降维为"磁盘流（Append-Only Stream）"

放弃在 Go 内存中用指针去维系复杂的 `Children []*Span` 树状层级。

**扁平化结构**：每个 Span 变成一个完全独立的、扁平的日志单体（包含自身 ID 与 ParentID），类似于：

```json
{"traceid": "tx001", "spanid": "span02", "parentid": "span01", "name": "LLM.Action", "start_time": "..."}
```

**按序追加**：当代码触发 `StartSpan` 或 `EndSpan` 时，拦截器立刻、同步将这一行结构化数据以 Append 模式写入到当前时间片的局部物理文件中。内存中只保留轻量级的 TraceID 上下文计数器。

#### 2. 行为借鉴：引入 Nginx 式的"时间轴时间片轮转（Rotation）"

正如 Nginx 可以配置 `access.log` 按天、按小时或按文件大小切分，Agent Harness 的追踪器也引入"时间/轮次切分窗口（Time-Window Splitter）"：

**切分策略**：

- **按时间（Time-Based）**：比如每过 5 分钟自动关闭当前的 `tracesegment1.json`，后续数据直接物理写入 `tracesegment2.json`。
- **按逻辑轮次（Turn-Based）**：对于 Agent 更实用——以 ReAct 的 Turn 逻辑轮次作为物理切分边界。Turn-1 的所有 Action/Thinking/Tool 调用全部写入 `turn1.json`；一旦进入新循环，自动物理切换到 `turn2.json`。

**物理红利**：这样任何一个临时切分文件的大小都被死死锁在几百 KB 以内。开发者想排查第 8 轮为什么报错，直接用轻量级文本编辑器秒开 `turn_8.json` 即可，完全不需要给几十 MB 的巨型文件"收尸"。

#### 3. 收尾进化：无损懒拼接（Lazy Map-Reduce）与前端看板还原

**静态索引（Manifest）**：在切分的同时，Harness 维护一个极小的 `.manifest` 索引文件，记录这个 Session 一共切出了哪些时间片碎片。

**按需拼接（On-Demand Stitching）**：

- 当人类工程师点击复盘按钮时，后端后台（或 CLI 离线工具）启动一个轻量级的 Merge 脚本。
- 它像 MapReduce 一样，读取 Manifest，按照 `parent_id` 的拓扑关系，在磁盘流式读取的同时，异步拼装出一棵标准的 OpenTelemetry (OTel) 兼容树。

**直投可视化看板**：拼接完成的格式不直接保存为文件，而是直接通过标准的 HTTP/GRPC 协议，伪装成 OTel 标准流量，"刷"地一声投递给本地私有化部署的 Jaeger 或 Zipkin 看板。

---

## 验收标准

- [ ] Span 数据结构支持扁平化 JSON Append 写入
- [ ] 支持按时间或按 Turn 轮次切分文件
- [ ] Manifest 索引文件维护完整
- [ ] OTel 格式转换与 Jaeger/Zipkin 兼容
- [ ] 并行 Span 可视化为甘特图样式