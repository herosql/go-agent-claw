// internal/loginit/loginit.go
// 统一日志初始化：用 slog 替代 log，保留 time/level/msg 前缀，
// 消息内容格式与原来 log.Printf 完全一致。
package loginit

import (
	"log"
	"log/slog"
	"os"
)

// Init 将全局 slog 设置为默认 TextHandler（输出格式：time level msg），
// 并把 standard log 包的输出也指向同一 writer，实现全统一。
func Init() {
	// 用默认 TextHandler，保留 time= + level= + msg= 前缀
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	// log 包也指向 stderr（slog 输出到 os.Stderr）
	log.SetOutput(os.Stderr)
}