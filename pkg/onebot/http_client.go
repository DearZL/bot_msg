package onebot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (a *Client) callHTTP(ctx context.Context, action string, payload any, dst any) error {
	ctx, cancel := a.withTimeout(ctx)
	defer cancel()

	// OneBot HTTP API 使用 action 作为路径，请求参数放在 JSON body 中。
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal onebot http %s request: %w", action, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.httpURL+"/"+action, bytes.NewReader(rawPayload))
	if err != nil {
		return fmt.Errorf("create onebot http %s request: %w", action, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.accessToken != "" {
		// 大多数 OneBot 实现支持 Bearer token。
		req.Header.Set("Authorization", "Bearer "+a.accessToken)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call onebot http %s: %w", action, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read onebot http %s response: %w", action, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("onebot http %s status %d: %s", action, resp.StatusCode, string(body))
	}

	// HTTP 和 WebSocket 返回体结构一致，这里复用同一个响应结构。
	var apiResp oneBotWSResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("decode onebot http %s response: %w", action, err)
	}
	if apiResp.Status != "ok" || apiResp.RetCode != 0 {
		return &Error{
			Action:  action,
			RetCode: apiResp.RetCode,
			Message: apiResp.Message,
			Wording: apiResp.Wording,
		}
	}
	if dst != nil && len(apiResp.Data) > 0 && string(apiResp.Data) != "null" {
		if err := json.Unmarshal(apiResp.Data, dst); err != nil {
			return fmt.Errorf("decode onebot http %s data: %w", action, err)
		}
	}
	return nil
}
