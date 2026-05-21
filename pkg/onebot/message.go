package onebot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Message []Segment

// Segment 表示 OneBot 消息段，例如 text、image、at 等。
type Segment struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// TextMessage 创建一个纯文本消息。
func TextMessage(text string) Message {
	return Message{{
		Type: "text",
		Data: map[string]any{"text": text},
	}}
}

// UnmarshalJSON 兼容 OneBot 常见消息格式：字符串、单个消息段、消息段数组。
func (m *Message) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*m = nil
		return nil
	}

	if data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*m = TextMessage(text)
		return nil
	}

	var segments []Segment
	if err := json.Unmarshal(data, &segments); err == nil {
		*m = Message(segments)
		return nil
	}

	var single Segment
	if err := json.Unmarshal(data, &single); err == nil && single.Type != "" {
		*m = Message{single}
		return nil
	}

	return fmt.Errorf("message must be a string, segment object, or segment array")
}

// UnmarshalText 用于支持 query/form 里的文本消息。
func (m *Message) UnmarshalText(text []byte) error {
	*m = TextMessage(string(text))
	return nil
}

// PlainText 提取所有 text 消息段中的文本，非文本消息段会被忽略。
func (m Message) PlainText() string {
	var builder strings.Builder
	for _, segment := range m {
		if segment.Type != "text" || segment.Data == nil {
			continue
		}
		if text, ok := segment.Data["text"].(string); ok {
			builder.WriteString(text)
		}
	}
	return builder.String()
}

// Empty 判断消息是否为空。
func (m Message) Empty() bool {
	return len(m) == 0
}
