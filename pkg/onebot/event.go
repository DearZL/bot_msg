package onebot

import (
	"context"
	"encoding/json"
)

type EventHandler func(context.Context, Event)

// Event 是所有 OneBot 事件的基础结构，Raw 保留原始事件方便上层自行解析。
type Event struct {
	PostType string          `json:"post_type"`
	Time     int64           `json:"time"`
	SelfID   int64           `json:"self_id"`
	Raw      json.RawMessage `json:"-"`
}

// MessageEvent 表示 OneBot message 事件。
type MessageEvent struct {
	Event
	MessageType string  `json:"message_type"`
	SubType     string  `json:"sub_type"`
	MessageID   int64   `json:"message_id"`
	UserID      int64   `json:"user_id"`
	GroupID     int64   `json:"group_id,omitempty"`
	Message     Message `json:"message"`
	RawMessage  string  `json:"raw_message"`
}

// Events 返回事件通道，适合业务方自己消费所有事件。
func (a *Client) Events() <-chan Event {
	return a.events
}

// OnEvent 注册通用事件回调。
func (a *Client) OnEvent(handler EventHandler) {
	a.handlersMu.Lock()
	a.handlers = append(a.handlers, handler)
	a.handlersMu.Unlock()
}

// OnMessage 注册消息事件回调，只处理 post_type=message 的事件。
func (a *Client) OnMessage(handler func(context.Context, MessageEvent)) {
	a.OnEvent(func(ctx context.Context, event Event) {
		if event.PostType != "message" {
			return
		}
		var msgEvent MessageEvent
		if err := json.Unmarshal(event.Raw, &msgEvent); err != nil {
			return
		}
		msgEvent.Event.Raw = event.Raw
		handler(ctx, msgEvent)
	})
}

// dispatchEvent 将 WebSocket 收到的无 echo 消息视为事件并分发。
func (a *Client) dispatchEvent(raw []byte) {
	var event Event
	if err := json.Unmarshal(raw, &event); err != nil || event.PostType == "" {
		return
	}
	event.Raw = append(json.RawMessage(nil), raw...)

	select {
	case a.events <- event:
	default:
		a.logger.Warn("onebot event dropped because buffer is full", "post_type", event.PostType)
	}

	a.handlersMu.RLock()
	handlers := append([]EventHandler(nil), a.handlers...)
	a.handlersMu.RUnlock()
	for _, handler := range handlers {
		go handler(context.Background(), event)
	}
}
