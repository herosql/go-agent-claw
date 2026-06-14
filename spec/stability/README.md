# 稳定性控制与多智能体

> 状态: 待扩展

---

## 模块概述

稳定性控制与多智能体模块负责防止 Agent 陷入死循环、实现高危命令拦截、以及通过 Subagent 隔离复杂探索任务。

---

## 子模块索引

| 文档 | 状态 | 描述 |
|------|------|------|
| [system-reminders.md](system-reminders.md) | 待扩展 | 行为干预：防止 Agent 陷入"死循环"的 System Reminders 机制 |
| [middleware-permissions.md](middleware-permissions.md) | 待扩展 | 防御纵深：利用 Middleware 实现高危命令拦截与飞书人工审批 |
| [subagent-isolation.md](subagent-isolation.md) | 待扩展 | 任务委派：引入 Subagent 隔离复杂探索任务的上下文瓶颈 |

---

## 核心设计原则

1. **参数规范化**：剥离大模型伪装，捕捉"本质上"的死循环
2. **三态权限矩阵**：allow / ask / deny + 组织架构 + 资源范围
3. **非阻塞委派**：Subagent 升级为 OS 级"后台轻量级进程"

---

## 关键技术方案

### 1. System Reminders - 三步死循环检测

- **步骤一**：参数规范化（Canonical JSON、Path Clean、Command Sanitization）
- **步骤二**：相似性矩阵判断（Levenshtein Distance 莱文斯坦距离）
- **步骤三**：强力干预（Role=User 硬提醒，强制扭转乾坤）

### 2. Middleware 权限引擎 - 双缓冲热加载

- **Schema**：三态分类（allow/ask/deny）+ 组织架构矩阵
- **并发控制**：读写锁（sync.RWMutex）+ 防抖动（Debounce）
- **双缓冲**：指针赋值实现纳秒级配置切换

### 3. Subagent 非阻塞架构

- **解耦拉起**：spawn_subagent 瞬间返回，不死等
- **PID 句柄**：后台线程池分配唯一任务句柄
- **双模式**：Proactive Polling / Event-Driven Epoll 模式

---

## 待扩展

### 待扩展

> 待补充详细设计方案

1. **ReminderInjector 与 Session 的集成**：计数器机制在 Session 层实现
2. **多警长协同**：多个高危规则同时触发时的优先级判定

### 待扩展2

> 未来规划方向

1. **群体辩论与一致性投票**：多模型 Reviewer 达成共识
2. **点对点协商（P2P Negotiation）**：Subagent 间直接通信