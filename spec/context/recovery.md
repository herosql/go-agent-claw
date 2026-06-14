# 错误自愈机制

> 状态: 待扩展

---

## 1. Error Recovery 提示模板注入

当工具返回错误时，RecoveryManager 拦截并注入提示模板：

```go
type RecoveryManager struct {
    templates map[string]string  // error pattern -> recovery hint
}

func (rm *RecoveryManager) Inject(ctx context.Context, toolErr error) string {
    pattern := classifyError(toolErr)
    if hint, ok := rm.templates[pattern]; ok {
        return hint
    }
    // 未知错误：调用小模型翻译
    return rm.translateWithSmallModel(toolErr)
}
```

---

## 2. 用 AI 治愈 AI（未知错误处理）

### 方案

当拦截到未知报错时，悄悄调用便宜小模型（如 GLM-4 Flash）：
1. 将长篇堆栈翻译成简短的"人话"建议
2. 注入给主 Agent

### 评估

| 维度 | 评估 |
|------|------|
| 延迟 | 小模型调用增加 1-2 秒 |
| 成本 | 极低（约 0.001 元/次） |
| 收益 | 将未知错误转化为可操作建议 |

---

## 3. 典型错误模板

| 错误类型 | 模板提示 |
|----------|----------|
| 文件不存在 | 检查文件路径是否正确，确认文件是否已被移动或删除 |
| 权限不足 | 尝试使用 sudo 或检查文件权限设置 |
| 命令不存在 | 检查是否已安装该命令，或使用完整路径 |
| 网络超时 | 检查网络连接，或增加超时时间 |

---

## 4. 待实现细节

- [ ] RecoveryManager 结构设计
- [ ] 错误模式分类（classifyError）
- [ ] 小模型翻译集成
- [ ] 模板注入时机与位置