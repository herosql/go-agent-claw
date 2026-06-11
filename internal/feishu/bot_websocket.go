package feishu

import (
	"context"
	"log"
	"strings"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// GetEventDispatcher 构建事件调度器（WebSocket 版本）
func (b *FeishuBot) GetEventDispatcherSocket() *dispatcher.EventDispatcher {
	// ========== 改动点1：WebSocket 不需要加密密钥、验签 Token，直接传 nil ==========
	handler := dispatcher.NewEventDispatcher("", "").
		// 私聊消息接收回调（原有业务逻辑完全复用，一行不改）
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			contentStr := *event.Event.Message.Content
			contentStr = strings.TrimPrefix(contentStr, `{"text":"`)
			contentStr = strings.TrimSuffix(contentStr, `"}`)

			chatId := *event.Event.Message.ChatId
			log.Printf("[Feishu] 收到会话 %s 消息: %s\n", chatId, contentStr)

			// 原有：起 Goroutine 跑 Agent，避免阻塞，保留不变
			go b.handleAgentRun(chatId, contentStr)

			return nil
		}).
		// 消息已读事件，保留不变
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			return nil
		})

	return handler
}
