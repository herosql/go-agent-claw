// internal/tools/bash.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/herosql/go-agent-claw/internal/schema"
)

type BashTool struct {
	workDir string // 工作区约束
}

func NewBashTool(workDir string) *BashTool {
	return &BashTool{workDir: workDir}
}

func (t *BashTool) Name() string {
	return "bash"
}

func (t *BashTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        t.Name(),
		Description: "在当前工作区执行任意的 bash 命令。支持链式命令(如 &&)。返回标准输出(stdout)和标准错误(stderr)。",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "要执行的 bash 命令，例如: ls -la 或 go test ./...",
				},
			},
			"required": []string{"command"},
		},
	}
}

type bashArgs struct {
	Command string `json:"command"`
}

// progressReporter is the interface used by BashTool to send running updates.
type progressReporter interface {
	OnToolRunning(ctx context.Context, toolName string, elapsedSecs int)
}

// progressKey is a private context key type to avoid collisions.
type progressKey struct{}

// reporterFromContext extracts a progressReporter from context, if present.
func reporterFromContext(ctx context.Context) progressReporter {
	if r := ctx.Value(progressKey{}); r != nil {
		return r.(progressReporter)
	}
	return nil
}

// SetProgressReporter returns a new context with the reporter attached.
func SetProgressReporter(ctx context.Context, r progressReporter) context.Context {
	return context.WithValue(ctx, progressKey{}, r)
}

func (t *BashTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input bashArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	// 【驾驭底线 1】：Time Budgeting (时间预算与超时控制)
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 尝试从 context 中获取进度 reporter
	reporter := reporterFromContext(ctx)

	// 启动进度通知 goroutine：每 5 秒发送一次"执行中"状态
	done := make(chan struct{})
	if reporter != nil {
		go func() {
			for i := 5; ; i += 5 {
				select {
				case <-done:
					return
				case <-time.After(5 * time.Second):
					reporter.OnToolRunning(ctx, t.Name(), i)
				}
			}
		}()
	}

	// 执行 bash 命令
	cmd := exec.CommandContext(timeoutCtx, "bash", "-c", input.Command)
	cmd.Dir = t.workDir

	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	// 通知 goroutine 退出
	close(done)

	// 超时处理
	if timeoutCtx.Err() == context.DeadlineExceeded {
		return outputStr + "\n[警告: 命令执行超时(30s)，已被系统强制终止。如果是启动常驻服务，请尝试将其转入后台。]", nil
	}

	// 错误处理
	if err != nil {
		return fmt.Sprintf("执行报错: %v\n输出:\n%s", err, outputStr), nil
	}

	// 空输出
	if outputStr == "" {
		return "命令执行成功，无终端输出。", nil
	}

	// 长度截断保护
	const maxLen = 8000
	if len(outputStr) > maxLen {
		return fmt.Sprintf("%s\n\n...[终端输出过长，已截断至前 %d 字节]...", outputStr[:maxLen], maxLen), nil
	}

	return outputStr, nil
}