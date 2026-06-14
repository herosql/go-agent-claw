// internal/loginit/loginit.go
// 统一日志初始化：用 slog 替代 log，
// 输出格式：2006/01/02 15:04:05 INFO message
package loginit

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// customHandler 自定义文本日志处理器
type customHandler struct {
	slog.TextHandler
}

// Handle 重写日志输出格式
func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	_ = time.Now() // 确保 time 包被使用
	timeStr := r.Time.Format("2006/01/02 15:04:05")
	levelStr := r.Level.String()
	msg := r.Message

	// 收集属性
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a.Key+"="+formatAttrValue(a.Value))
		return true
	})

	line := timeStr + " " + levelStr + " " + msg
	if len(attrs) > 0 {
		line += " " + attrs[0]
		for i := 1; i < len(attrs); i++ {
			line += " " + attrs[i]
		}
	}
	line += "\n"

	_, err := os.Stderr.WriteString(line)
	return err
}

func formatAttrValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return formatInt(v.Int64())
	case slog.KindFloat64:
		return formatFloat(v.Float64())
	default:
		return v.String()
	}
}

func formatInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

func formatFloat(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// Init 将全局 slog 设置为自定义格式（输出格式：2006/01/02 15:04:05 INFO message），
// 并把 standard log 包的输出也指向同一 writer，实现全统一。
func Init() {
	h := &customHandler{*slog.NewTextHandler(os.Stderr, nil)}
	slog.SetDefault(slog.New(h))
	log.SetOutput(os.Stderr)
}
