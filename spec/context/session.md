# Session 与记忆管理

> 状态: 待扩展

---

## 1. 多层记忆体系

### 短期工作记忆（Working Memory）

`GetWorkingMemory()` 只保留最近 N 轮会话消息，保证 API 请求不 OOM。

### 任务级状态记忆（State Memory）

PLAN.md 和 TODO.md，伴随任务生灭。

### 情景记忆沉淀池（Episodic Memory）

在 `~/.openclaw/workspace/memory/` 下：
- 按日期生成 `2026-04-12.md` 日志
- 维护 `MEMORY.md` 汇总长期事实

### 长程记忆检索（Hybrid Retrieval）

- 向量搜索（语义相似性）
- BM25 关键词搜索
- 合并排序

---

## 2. 动态算力分配（触发条件）

### 宏观触发

外部记忆文件（PLAN.md）检测到任务目标变更时，在下一个 Turn 开始前开启慢思考。

### 微观触发

工具调用返回非预期结果（Error 或严重偏差）时，在当前 Turn 内动态开启慢思考。

---

## 3. Session 物理隔离

每个工作区绑定独立的 Registry 实例：

```go
// 引擎按工作区隔离
engine, ok := engines[workDir]
if !ok {
    engine = NewAgentEngine(workDir, registry)
    engines[workDir] = engine
}
```

通过 `context.Context` 透传当前 Session 的 WorkDir。

---

## 4. 待实现细节

- [ ] 情景记忆沉淀池写入逻辑
- [ ] Hybrid Retrieval 实现
- [ ] 宏观/微观触发条件检测
- [ ] Session 级别的 Context 传递