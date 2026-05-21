package onebot

import (
	"encoding/json"
	"testing"
)

func TestMessageUnmarshalString(t *testing.T) {
	var msg Message
	if err := json.Unmarshal([]byte(`"hello"`), &msg); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}

	if got := msg.PlainText(); got != "hello" {
		t.Fatalf("PlainText() = %q, want %q", got, "hello")
	}
}

func TestMessageUnmarshalSegments(t *testing.T) {
	var msg Message
	raw := []byte(`[{"type":"text","data":{"text":"hello"}},{"type":"image","data":{"file":"a.png"}}]`)
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}

	if len(msg) != 2 {
		t.Fatalf("len(msg) = %d, want 2", len(msg))
	}
	if got := msg.PlainText(); got != "hello" {
		t.Fatalf("PlainText() = %q, want %q", got, "hello")
	}
}
