package event

import (
	"context"
	"log/slog"

	"bot_msg/pkg/onebot"
)

type MessageLogHandler struct {
	logger *slog.Logger
}

// NewMessageLogHandler 创建只打印日志、不做业务动作的默认消息处理器。
func NewMessageLogHandler(logger *slog.Logger) *MessageLogHandler {
	return &MessageLogHandler{
		logger: logger,
	}
}

func (h *MessageLogHandler) Name() string {
	return "message_log"
}

// HandleMessage 记录收到的消息摘要，避免默认处理器产生回复、落库或转发等副作用。
func (h *MessageLogHandler) HandleMessage(ctx context.Context, event onebot.MessageEvent) (MessageResult, error) {
	h.logger.InfoContext(ctx, "onebot message received",
		"message_type", event.MessageType,
		"sub_type", event.SubType,
		"message_id", event.MessageID,
		"user_id", event.UserID,
		"group_id", event.GroupID,
		"self_id", event.SelfID,
		"time", event.Time,
		"raw_message", event.RawMessage,
		"plain_text", event.Message.PlainText(),
		"segments", len(event.Message),
	)
	return Continue, nil
}
