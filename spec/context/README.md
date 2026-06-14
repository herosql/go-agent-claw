# 上下文工程体系

> 状态: 待扩展

---

## 模块概述

上下文工程体系是 go-agent-claw 的核心模块之一，负责管理 LLM 的"神经内存"，借鉴 OS 的内存管理策略，实现上下文的高效组装、压缩、持久化和自愈。

---

## 子模块索引

| 文档 | 状态 | 描述 |
|------|------|------|
| [prompt-composer.md](prompt-composer.md) | 待扩展 | 提示词组装：告别面条代码，动态加载 AGENTS.md 与外挂 Skills |
| [session-management.md](session-management.md) | 待扩展 | 会话管理：Session 物理隔离与 Working Memory 的底层实现 |
| [compactor.md](compactor.md) | 待扩展 | 突破内存：基于阶梯降级的 Context Compaction 策略 |
| [memory-persistence.md](memory-persistence.md) | 待扩展 | 记忆沉淀：状态外部化，基于文件系统的持久化记忆与待办管理 |
| [error-recovery.md](error-recovery.md) | 待扩展 | 错误自愈：上下文感知的 Error Recovery 提示模板注入机制 |

---

## 核心设计原则

1. **渐进式暴露（Progressive Disclosure）**：技能元数据常驻，技能正文按需加载
2. **阶梯降级**：多级压缩策略，从滑动窗口到 Token 水位线自适应
3. **状态外部化**：记忆沉淀至文件系统，防止上下文丢失
4. **双轨自愈**：硬编码优先 + 小模型兜底，覆盖 80% 高频坑 + 20% 长尾报错

---

## 关键技术方案

### 1. 提示词组装 - read_skill 原生工具

- **Discovery 阶段**：极简元数据常驻内存（Skill ID + Trigger + Capability）
- **Activation 阶段**：注入 `read_skill(skill_id)` 工具，按需加载技能正文
- **Thought Anchor**：Thinking System Prompt 强制自省，先查技能菜单再决定是否造轮子

### 2. 会话管理 - 三阶段压缩

- **第一阶段**：纵向裁剪，窗口粗筛（按条数）
- **第二阶段**：横向挤压，单条长消息外科手术（Head-Tail Truncation / Middle-Out）
- **第三阶段**：总体合规验证，安全水位线判定

### 3. Context Compaction - 自适应压缩

- **Provider 属性注入**：ContextWindowLimit、ModelName、TokenEstimator
- **动态水位线**：基于真实 API Usage 反馈进行系数自修正
- **闭环校准**：影子监控 + 缩放惩罚系数动态调整

### 4. 记忆沉淀 - 多层架构

- **Working Memory**：最近 N 轮会话消息
- **State Memory**：PLAN.md / TODO.md，任务级状态
- **Episodic Memory**：按日期的日志文件 + MEMORY.md 长期记忆
- **Hybrid Retrieval**：向量搜索 + BM25 关键词搜索

### 5. 错误自愈 - 双轨制

- **第一轨**：Domain Error Code 硬编码匹配，O(1) 微秒级
- **第二轨**：私有化小模型（如 GLM-4 Flash）Summary 泛化降级
- **死循环防御**：同一错误最多注入 2 次，阈值 3 触发人工审批

---

## 待扩展

### 待扩展

> 以下内容待补充详细设计方案

1. **read_skill 工具的 Tool Registry 集成细节**：如何与现有 Middleware 链配合
2. **多 Skill 并发激活的场景**：大模型同时需要多个技能时的上下文管理
3. **技能版本管理与缓存淘汰策略**

### 待扩展2

> 以下内容为未来规划方向

1. **跨会话的技能记忆**：技能使用频次统计，自动推荐高频技能
2. **技能的动态编排**：根据任务类型自动组合多个技能的 SOP