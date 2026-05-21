package event

import (
	"context"
	"log/slog"
	"sync"

	"bot_msg/pkg/onebot"
)

type MessageHandler interface {
	// Name 返回处理器名称，用于日志和排查问题。
	Name() string
	// HandleMessage 处理一条消息事件，返回 Continue 或 Stop 控制后续分发。
	HandleMessage(ctx context.Context, event onebot.MessageEvent) (MessageResult, error)
}

type MessageResult int

const (
	// Continue 表示继续交给下一个处理器。
	Continue MessageResult = iota
	// Stop 表示当前消息已经处理完成，停止后续分发。
	Stop
)

// MessageHandlerFunc 方便用普通函数快速实现 MessageHandler。
type MessageHandlerFunc struct {
	name string
	fn   func(context.Context, onebot.MessageEvent) (MessageResult, error)
}

// NewMessageHandlerFunc 创建函数式消息处理器。
func NewMessageHandlerFunc(name string, fn func(context.Context, onebot.MessageEvent) (MessageResult, error)) MessageHandlerFunc {
	return MessageHandlerFunc{
		name: name,
		fn:   fn,
	}
}

func (h MessageHandlerFunc) Name() string {
	return h.name
}

func (h MessageHandlerFunc) HandleMessage(ctx context.Context, event onebot.MessageEvent) (MessageResult, error) {
	return h.fn(ctx, event)
}

type Dispatcher struct {
	logger *slog.Logger

	// messageHandlers 按注册顺序执行，适合实现日志、命令、webhook、AI 等流水线。
	mu              sync.RWMutex
	messageHandlers []MessageHandler
}

// NewDispatcher 创建消息分发器，并默认注册日志处理器。
func NewDispatcher(logger *slog.Logger, handlers ...MessageHandler) *Dispatcher {
	dispatcher := &Dispatcher{
		logger: logger,
	}
	dispatcher.UseMessageHandler(NewMessageLogHandler(logger))
	for _, handler := range handlers {
		dispatcher.UseMessageHandler(handler)
	}
	return dispatcher
}

// Register 将分发器注册到 SDK 的 OnMessage 回调。
func (d *Dispatcher) Register(client *onebot.Client) {
	client.OnMessage(d.DispatchMessage)
}

// UseMessageHandler 追加一个消息处理器。
func (d *Dispatcher) UseMessageHandler(handler MessageHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.messageHandlers = append(d.messageHandlers, handler)
}

// DispatchMessage 是所有消息事件进入业务侧后的统一分发入口。
func (d *Dispatcher) DispatchMessage(ctx context.Context, event onebot.MessageEvent) {
	handlers := d.snapshotMessageHandlers()
	for _, handler := range handlers {
		result, err := handler.HandleMessage(ctx, event)
		if err != nil {
			d.logger.ErrorContext(ctx, "onebot message handler failed",
				"handler", handler.Name(),
				"message_type", event.MessageType,
				"message_id", event.MessageID,
				"user_id", event.UserID,
				"group_id", event.GroupID,
				"error", err,
			)
			continue
		}
		if result == Stop {
			d.logger.DebugContext(ctx, "onebot message dispatch stopped",
				"handler", handler.Name(),
				"message_id", event.MessageID,
			)
			return
		}
	}
}

// snapshotMessageHandlers 复制当前处理器列表，避免分发时长时间持有锁。
func (d *Dispatcher) snapshotMessageHandlers() []MessageHandler {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return append([]MessageHandler(nil), d.messageHandlers...)
}
