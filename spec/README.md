# go-agent-claw 架构设计规格书

> 生成时间: 2026-06-14
> 状态: 持续更新中

---

## 模块索引

### 1. 认知和核心引擎
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/core/engine-loop.md](core/engine-loop.md) | 待扩展 | Agent Main Loop 与 ReAct 循环机制 |
| [spec/core/thinking.md](core/thinking.md) | 待扩展 | 慢思考与自省机制 |
| [spec/core/provider.md](core/provider.md) | 待扩展 | Provider 接口抽象 |

### 2. 极简工具与物理交互
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/tools/registry.md](tools/registry.md) | 待扩展 | Tool Registry 与分发机制 |
| [spec/tools/parallel-exec.md](tools/parallel-exec.md) | 待扩展 | 并行工具执行与流控 |
| [spec/tools/edit-fuzzy.md](tools/edit-fuzzy.md) | 待扩展 | Edit 工具模糊匹配 |
| [spec/tools/bash-background.md](tools/bash-background.md) | 待扩展 | Bash 后台任务管理 |

### 3. 上下文工程体系
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/context/README.md](spec/context/README.md) | 待扩展 | 上下文工程体系总览 |
| [spec/context/prompt-composer.md](spec/context/prompt-composer.md) | 待扩展 | 提示词组装：read_skill 渐进式加载 |
| [spec/context/session-management.md](spec/context/session-management.md) | 待扩展 | 会话管理：三阶段上下文压缩 |
| [spec/context/compactor.md](spec/context/compactor.md) | 待扩展 | 突破内存：自适应 Token 水位线压缩 |
| [spec/context/memory-persistence.md](spec/context/memory-persistence.md) | 待扩展 | 记忆沉淀：多层记忆架构与 Hybrid Retrieval |
| [spec/context/error-recovery.md](spec/context/error-recovery.md) | 待扩展 | 错误自愈：双轨制 Error Recovery 机制 |

### 4. 稳定性控制与多智能体
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/stability/README.md](spec/stability/README.md) | 待扩展 | 稳定性控制与多智能体总览 |
| [spec/stability/system-reminders.md](spec/stability/system-reminders.md) | 待扩展 | 行为干预：System Reminders 死循环检测 |
| [spec/stability/middleware-permissions.md](spec/stability/middleware-permissions.md) | 待扩展 | 防御纵深：Middleware 权限引擎与热加载 |
| [spec/stability/subagent-isolation.md](spec/stability/subagent-isolation.md) | 待扩展 | 任务委派：Subagent 非阻塞异步架构 |

### 5. 飞书集成
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/feishu/scheduler.md](feishu/scheduler.md) | 待扩展 | 工作区任务调度队列 |
| [spec/feishu/agentops.md](feishu/agentops.md) | 待扩展 | AgentOps 小助手与意图拦截 |

### 6. 可观测性与科学度量
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/observability/tracker.md](observability/tracker.md) | 待扩展 | 成本与状态追踪、工具耗时拦截器 |
| [spec/observability/tracing.md](observability/tracing.md) | 待扩展 | 链路追踪与 Jaeger/Zipkin 可视化 |
| [spec/observability/benchmark.md](observability/benchmark.md) | 待扩展 | Benchmark 自动化评估体系 |

### 7. 扩展
| 文档 | 状态 | 描述 |
|------|------|------|
| [spec/extensions/shallow.md](extensions/shallow.md) | 待扩展3 | 浅层扩展：AST、LSP、Computer Use、MCP |
| [spec/extensions/deep.md](extensions/deep.md) | 待扩展3 | 深层扩展：沙箱、权限、插件、自进化 |

---

## 核心设计原则

1. **OS 类比**：将 LLM Context 视作"神经内存"，借鉴 OS 的内存管理策略
2. **弹性容错**：部分失败不等于全部失败，工具执行支持并发与优雅降级
3. **自适应算力**：根据任务复杂度动态分配 Thinking 资源
4. **安全第一**：路径穿越防护、Workspace 隔离、沙箱隔离

---

## 层级架构对照表

| 层级 | OS 对应 | go-agent-claw 组件 |
|------|---------|-------------------|
| L1 寄存器 | 极速寄存器 | Working Memory (最近 N 轮对话) |
| L2 缓存 | 常驻内存 | PLAN.md / TODO.md 状态文件 |
| L3 磁盘 | 冷备存储 | Vector DB / 文件归档 |
| 调度器 | OOM Killer | Compactor / Context Truncation |
| 沙箱 | 进程隔离 | Docker / Workspace Mutex |