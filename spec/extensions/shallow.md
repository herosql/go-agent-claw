# 扩展：浅层

> 状态: 待扩展3

## 模块职责

轻量级功能扩展，不涉及核心架构改动，可渐进式集成。

---

## 待扩展3

### 基于 AST 的终极外科手术刀

目前我们的 `edit_file` 工具依赖的是多级字符串模糊匹配（Fuzzy Match）。这在多数的场景下可能工作良好，但在修改极其复杂的嵌套代码时，大模型依然容易在缩进上产生幻觉。

如果想把 go-tiny-claw 打造成一个纯粹用于代码重构的终极利器，可以尝试把 Go 语言的 **AST（抽象语法树）解析器**，或者 **LSP（Language Server Protocol）客户端**集成到底层工具中。

这样，大模型就能通过更加语义化的指令（如："把 main 包下 User 结构体的 Name 字段改为小写"）来精准重构代码，彻底消灭基于文本替换导致的语法破坏。

**待实现**：
- [ ] AST 解析器封装
- [ ] LSP 客户端集成
- [ ] 语义化指令到代码变更的转换层

---

### Computer Use：从终端走向 GUI 桌面

你可能已经注意到了 Anthropic 等厂商推出的 Computer Use 能力（让大模型直接控制鼠标键盘、看屏幕截图）。在驾驭工程的视角下，这其实并没有改变 Main Loop 的核心架构！

你只需要在 go-tiny-claw 的 Tool Registry 中新增几个极简的 GUI 原语工具：
- `take_screenshot`（截图并转为 base64 发给多模态大模型）
- `click_mouse(x, y)`
- `type_text(string)`

然后，辅以一个稍微复杂点的 Prompt Composer，你的 Agent 就能瞬间从一个"只会敲代码的终端极客"，进化为一个"能帮你点外卖、做 Excel 报表的桌面全能管家"。

**待实现**：
- [ ] `take_screenshot` 工具注册
- [ ] `click_mouse` 工具注册
- [ ] `type_text` 工具注册
- [ ] 多模态模型对接（如果支持）

---

### 接入标准化协议：MCP（Model Context Protocol）

虽然在专栏中，为了贯彻极简主义，我们没有引入臃肿的 MCP。但在复杂的企业内部环境中，如果你的 AgentOps 助手需要去查询线上的 MySQL 数据库、对接 Jira 系统抓取工单，你不可能为所有的内部系统都在 Go 引擎里重写一遍 BaseTool。

未来，你完全可以在你的 Tool Registry 中引入一个轻量级的 **MCP Client**，让 go-tiny-claw 能够按需（Lazy Loading）发现并接入公司内部的标准化 MCP Server。

这会让你的 Agent 瞬间拥有跨越部门边界、操纵整个公司基建的能力。

**待实现**：
- [ ] MCP Client 库选型与集成
- [ ] 工具的 Lazy Loading 机制
- [ ] MCP Server 自动发现协议
- [ ] 公司内部 MCP Server 对接示例