# 补测计划

---

## 批次安排

### 第一批：Compactor + fuzzyReplace（纯函数，最容易）
- P0-01：`Compactor.Compact` 未超阈值返回原数组
- P0-02：`Compactor.Compact` 超阈值时正确压缩
- P0-03：`Compactor.Compact` 不修改 ToolCalls
- P0-04：`fuzzyReplace` L1 精确匹配唯一替换
- P0-05：`fuzzyReplace` L1 匹配 0 处返回 error
- P0-06：`fuzzyReplace` L1 匹配 >1 处返回 error
- P0-07：`fuzzyReplace` L4 逐行去缩进匹配

**文件**：`internal/context/compactor_test.go`、`internal/tools/edit_file_test.go`

---

### 第二批：Session + Registry
- P0-08：`Session.Append` 并发安全
- P0-09：`Session.GetWorkingMemory` 跳过 ToolResult 孤儿
- P0-10：`Registry.Execute` 工具不存在返回 IsError

**文件**：`internal/context/session_test.go`、`internal/tools/registry_test.go`

---

### 第三批：P1 补充（可选，视时间）
- 见 `docs/test-gaps.md` P1-01 至 P1-10

---

## 断言原则

- **Characterization Test**：凭"实际行为"写断言，即先通过测试了解真实行为，再将断言建立在实际行为上
- 不凭"应该"：不假设函数应该如何，而先通过实验确认它实际如何
- 所有断言以实际运行结果为准