#!/bin/bash

# Commit message rewrite script - maps old messages to new messages

declare -A messages

# From 88e2ae3 to HEAD (oldest first)
messages["88e2ae3"]="feat: 构建主循环"
messages["947ccba"]="feat: 添加 slog 日志和 Thinking"
messages["5a268f3"]="feat: 添加 OpenAI 和 Claude Provider"
messages["bc18a01"]="feat: 添加 OpenAI 和 Claude Provider（Provider 实现）"
messages["3ba2eaf"]="feat: 添加工具注册表"
messages["add9a3b"]="feat: 添加 Write & Bash 工具"
messages["5f5a914"]="feat: 实现健壮的 Edit 工具，支持多级模糊匹配"
messages["1886434"]="feat: 并发调用工具"
messages["10041fe"]="feat: 并发调用工具（测试文件）"
messages["0e28c0f"]="feat: 加入飞书"
messages["7a41061"]="feat: 动态加载 AGENTS.md 和外部 Skills"
messages["b7650f2"]="feat: 飞书机器人 WebSocket 长连接模式上线"
messages["41a21ed"]="fix: 慢思考"
messages["3de4720"]="fix: 加入 session"
messages["41f155f"]="fix: 测试写代码能力"
messages["3f2e15b"]="feat: 重构为 Session 会话模型，支持多用户/多终端并发隔离"
messages["282c02b"]="fix: 压缩上下文内存，防止大模型发生 OOM"
messages["bdd290f"]="fix: 修复 Zhipu API 1214 错误的多重根因"
messages["8884635"]="feat: 错误自愈：上下文感知的 Error Recovery 机制"
messages["c02a4d8"]="feat: 行为干预：防止 Agent 陷入死循环的 System Reminders 机制"
messages["140fecd"]="feat: 防御纵深：利用 Middleware 实现高危命令拦截与飞书人工审批"
messages["e1bd5d6"]="feat: 任务委派：引入 Subagent 隔离复杂探索任务的上下文瓶颈"
messages["d88937d"]="feat: 成本追踪：在 Harness 层拦截并记录 Token 消耗与执行耗时"
messages["b210e28"]="feat: Benchmark 自动化评估脚本，科学量化 Harness 引擎性能"
messages["990bc63"]="feat: 拼装完整 CLI 引擎，完成未知项目的文件探索与重构"
messages["dd6f50f"]="feat: 打造 AgentOps 小助手，在飞书中触发日志分析与故障修复审批，删除测试文件"
messages["c4afae8"]="ci: 添加 GitHub Actions 测试工作流"
messages["a284519"]="feat: 全面重构日志系统 + 飞书消息显示优化 + 开发工具链"
messages["055a6e2"]="docs: 补充 AGENTS.md 和 README.md 项目说明文档"

# Get the short hash (7 chars) of current commit
short=$(git rev-parse --short $GIT_COMMIT 2>/dev/null || echo "")

# If we have a mapping, use it, otherwise keep original
if [[ -n "$short" && -n "${messages[$short]}" ]]; then
    echo "${messages[$short]}"
else
    # Output original message
    cat
fi