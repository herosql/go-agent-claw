# Edit 工具模糊匹配

> 状态: 待扩展

---

## 1. 行级模糊匹配算法

当前 `lineByLineReplace` 算法已实现：
- 去除首尾空格干扰
- 定位匹配块
- 唯一性校验

---

## 2. 缩进自动补全（待实现）

### 问题

如果目标代码在 12 层嵌套中，而 newText 只有 4 空格缩进，替换后格式会很难看。

### 解决方案

#### 第一步：提取基准缩进

```go
func extractBaseIndentation(contentLines []string, matchStartIndex int) string {
    line := contentLines[matchStartIndex]
    // 统计左侧空格/制表符数量
    indent := ""
    for _, ch := range line {
        if ch == ' ' {
            indent += " "
        } else if ch == '\t' {
            indent += "\t"
        } else {
            break
        }
    }
    return indent
}
```

#### 第二步：对 newText 进行结构化处理

```go
func normalizeNewLines(newText string) []string {
    lines := strings.Split(newText, "\n")
    normalized := make([]string, 0, len(lines))
    for _, line := range lines {
        // 去除原有的缩进（左对齐）
        trimmed := strings.TrimLeft(line, " \t")
        normalized = append(normalized, trimmed)
    }
    return normalized
}
```

#### 第三步：注入缩进并缝合

```go
func injectIndentation(baseIndent string, newLines []string) []string {
    result := make([]string, len(newLines))
    for i, line := range newLines {
        result[i] = baseIndent + line
    }
    return result
}
```

---

## 3. 待实现细节

- [ ] `extractBaseIndentation` 函数
- [ ] `normalizeNewLines` 函数
- [ ] `injectIndentation` 函数
- [ ] 在 `lineByLineReplace` 中集成缩进补全逻辑