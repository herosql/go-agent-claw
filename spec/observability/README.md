# 可观测性与科学度量

> 状态: 待扩展

---

## 模块概述

可观测性模块负责在 Harness 层拦截并记录 Token 消耗、执行耗时、链路追踪，为 Agent 提供"洞察黑盒"的能力。

---

## 子模块索引

| 文档 | 状态 | 描述 |
|------|------|------|
| [cost-tracking.md](cost-tracking.md) | 待扩展 | 成本与状态追踪：在 Harness 层拦截并记录 Token 消耗与执行耗时 |
| [tracing.md](tracing.md) | 待扩展 | 链路追踪：Span 树形结构与 Jaeger 可视化集成 |

---

## 核心设计原则

1. **装饰器拦截**：Decorator/Middleware 模式，零侵入监控
2. **流式时间分片**：Trace 数据分段流式上报，避免大文件
3. **OTel 兼容**：标准协议，无缝对接 Jaeger/Zipkin

---

## 关键技术方案

### 1. CostTracker - Provider 装饰器

- **Token 计量**：每次请求拦截 Usage，累加 PromptTokens + CompletionTokens
- **工具耗时拦截**：defer + time.Since 夹击，零侵入记录物理执行时间
- **影子监控**：估算值与真实值对比，动态调整 Token 估算系数

### 2. Tracing - 流式时间分片架构

- **Segmented Stream Tracing**：将大 Trace 拆分为可独立传输的分片
- **OTel 协议转换**：Span 数据转换为 OTLP 格式上报 Jaeger
- **并行 Span 渲染**：甘特图展示 LLM 推理与并发工具执行的时间重叠

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **OTLP 导出器实现**：Span 数据到 Jaeger 的协议转换细节
2. **Trace 采样策略**：高流量场景下的采样率控制

### 待扩展2

> 未来规划方向

1. **分布式 Trace**：跨 Agent 的 trace_id + parent_span_id 传递
2. **自定义指标仪表盘**：Grafana 集成，自定义 SLO 看板