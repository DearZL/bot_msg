package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"bot_msg/internal/config"
	"bot_msg/internal/service"
	"bot_msg/pkg/onebot"
)

func newTestHTTPServer(accessToken string) *stdhttp.Server {
	logger := slog.New(slog.NewTextHandler(testDiscard{}, nil))
	client := &fakeOneBot{}
	msgService := service.New(config.BotConfig{
		SelfID:   42,
		Nickname: "tester",
	}, client)

	return New(config.HTTPConfig{
		Addr:        ":0",
		AccessToken: accessToken,
	}, msgService, logger)
}

type fakeOneBot struct {
	nextMessageID int64
}

func (a *fakeOneBot) SendPrivate(_ context.Context, _ int64, _ onebot.Message) (onebot.SendResult, error) {
	a.nextMessageID++
	return onebot.SendResult{MessageID: a.nextMessageID}, nil
}

func (a *fakeOneBot) SendGroup(_ context.Context, _ int64, _ onebot.Message) (onebot.SendResult, error) {
	a.nextMessageID++
	return onebot.SendResult{MessageID: a.nextMessageID}, nil
}

func (a *fakeOneBot) GetMessage(_ context.Context, messageID int64) (any, error) {
	return map[string]any{
		"message_id":  messageID,
		"raw_message": "hello",
	}, nil
}

func (a *fakeOneBot) Delete(_ context.Context, _ int64) error {
	return nil
}

func (a *fakeOneBot) GetFriendList(_ context.Context) (any, error) {
	return []any{}, nil
}

func (a *fakeOneBot) GetGroupList(_ context.Context) (any, error) {
	return []any{}, nil
}

func (a *fakeOneBot) GetGroupMemberList(_ context.Context, _ int64) (any, error) {
	return []any{}, nil
}

func (a *fakeOneBot) Healthy() bool {
	return true
}

func TestSendPrivateMsgAndGetMsg(t *testing.T) {
	server := newTestHTTPServer("")

	sendResp := performJSON(t, server, "/send_private_msg", map[string]any{
		"user_id": 12345,
		"message": "hello",
	})
	if sendResp.Code != stdhttp.StatusOK {
		t.Fatalf("send status = %d, body = %s", sendResp.Code, sendResp.Body.String())
	}

	var sendPayload struct {
		Status string `json:"status"`
		Data   struct {
			MessageID int64 `json:"message_id"`
		} `json:"data"`
	}
	decodeJSON(t, sendResp, &sendPayload)
	if sendPayload.Status != "ok" || sendPayload.Data.MessageID == 0 {
		t.Fatalf("unexpected send payload: %+v", sendPayload)
	}

	getResp := performJSON(t, server, "/get_msg", map[string]any{
		"message_id": sendPayload.Data.MessageID,
	})
	if getResp.Code != stdhttp.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.Code, getResp.Body.String())
	}

	var getPayload struct {
		Status string `json:"status"`
		Data   struct {
			MessageID  int64  `json:"message_id"`
			RawMessage string `json:"raw_message"`
		} `json:"data"`
	}
	decodeJSON(t, getResp, &getPayload)
	if getPayload.Data.MessageID != sendPayload.Data.MessageID || getPayload.Data.RawMessage != "hello" {
		t.Fatalf("unexpected get payload: %+v", getPayload)
	}
}

func TestAccessTokenRequired(t *testing.T) {
	server := newTestHTTPServer("secret")

	resp := performJSON(t, server, "/send_private_msg", map[string]any{
		"user_id": 12345,
		"message": "hello",
	})
	if resp.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.Code, stdhttp.StatusUnauthorized)
	}
}

func performJSON(t *testing.T, server *stdhttp.Server, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest(stdhttp.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	server.Handler.ServeHTTP(resp, req)
	return resp
}

func decodeJSON(t *testing.T, resp *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(resp.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode response %s: %v", resp.Body.String(), err)
	}
}

type testDiscard struct{}

func (testDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}
