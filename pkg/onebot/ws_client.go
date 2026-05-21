package onebot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

type Client struct {
	// HTTP 地址存在时，API 调用优先走 HTTP；WebSocket 主要用于收事件。
	httpURL           string
	wsURL             string
	accessToken       string
	timeout           time.Duration
	heartbeatInterval time.Duration
	reconnectInterval time.Duration
	logger            *slog.Logger

	echoSeq atomic.Uint64
	writeMu sync.Mutex

	mu          sync.Mutex
	conn        *websocket.Conn
	ready       chan struct{}
	reconnectCh chan struct{}
	pending     map[string]chan oneBotWSResult
	events      chan Event
	handlersMu  sync.RWMutex
	handlers    []EventHandler
	httpClient  *http.Client
}

// NewClient 使用 Option 构建 SDK 客户端。
func NewClient(opts ...Option) *Client {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewClientWithConfig(cfg)
}

// NewClientWithConfig 使用完整配置构建 SDK 客户端。
func NewClientWithConfig(cfg Config) *Client {
	cfg = normalizeConfig(cfg)
	return &Client{
		httpURL:           strings.TrimRight(cfg.HTTPURL, "/"),
		wsURL:             cfg.WebSocketURL,
		accessToken:       cfg.AccessToken,
		timeout:           cfg.Timeout,
		heartbeatInterval: cfg.HeartbeatInterval,
		reconnectInterval: cfg.ReconnectInterval,
		logger:            cfg.Logger,
		ready:             make(chan struct{}),
		reconnectCh:       make(chan struct{}, 1),
		pending:           make(map[string]chan oneBotWSResult),
		events:            make(chan Event, cfg.EventBuffer),
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// normalizeConfig 补齐未设置的默认值，避免调用方必须填满所有字段。
func normalizeConfig(cfg Config) Config {
	defaults := DefaultConfig()
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaults.Timeout
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = defaults.HeartbeatInterval
	}
	if cfg.ReconnectInterval <= 0 {
		cfg.ReconnectInterval = defaults.ReconnectInterval
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = defaults.EventBuffer
	}
	if cfg.Logger == nil {
		cfg.Logger = defaults.Logger
	}
	return cfg
}

// Start 启动 WebSocket 连接管理和心跳；仅配置 HTTP 时无需启动后台任务。
func (a *Client) Start(ctx context.Context) {
	if a.wsURL == "" {
		return
	}
	go a.connectLoop(ctx)
	go a.heartbeatLoop(ctx)
}

// SendPrivate 是 SendPrivateMessage 的兼容别名。
func (a *Client) SendPrivate(ctx context.Context, userID int64, msg Message) (SendResult, error) {
	return a.SendPrivateMessage(ctx, userID, msg)
}

// SendPrivateMessage 发送私聊消息。
func (a *Client) SendPrivateMessage(ctx context.Context, userID int64, msg Message) (SendResult, error) {
	payload := map[string]any{
		"user_id": userID,
		"message": msg,
	}

	var data oneBotSendData
	if err := a.Call(ctx, "send_private_msg", payload, &data); err != nil {
		return SendResult{}, err
	}

	a.logger.Info("private message sent by onebot websocket", "user_id", userID, "message_id", data.MessageID)
	return SendResult{
		MessageID: data.MessageID,
	}, nil
}

// SendGroup 是 SendGroupMessage 的兼容别名。
func (a *Client) SendGroup(ctx context.Context, groupID int64, msg Message) (SendResult, error) {
	return a.SendGroupMessage(ctx, groupID, msg)
}

// SendGroupMessage 发送群聊消息。
func (a *Client) SendGroupMessage(ctx context.Context, groupID int64, msg Message) (SendResult, error) {
	payload := map[string]any{
		"group_id": groupID,
		"message":  msg,
	}

	var data oneBotSendData
	if err := a.Call(ctx, "send_group_msg", payload, &data); err != nil {
		return SendResult{}, err
	}

	a.logger.Info("group message sent by onebot websocket", "group_id", groupID, "message_id", data.MessageID)
	return SendResult{
		MessageID: data.MessageID,
	}, nil
}

// SendMessage 调用 OneBot 通用 send_msg。
func (a *Client) SendMessage(ctx context.Context, messageType string, userID int64, groupID int64, msg Message) (SendResult, error) {
	payload := map[string]any{
		"message_type": messageType,
		"user_id":      userID,
		"group_id":     groupID,
		"message":      msg,
	}
	var data oneBotSendData
	if err := a.Call(ctx, "send_msg", payload, &data); err != nil {
		return SendResult{}, err
	}
	return SendResult{MessageID: data.MessageID}, nil
}

// GetMessage 查询消息详情。
func (a *Client) GetMessage(ctx context.Context, messageID int64) (any, error) {
	var data any
	err := a.Call(ctx, "get_msg", map[string]any{
		"message_id": messageID,
	}, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Delete 删除消息。
func (a *Client) Delete(ctx context.Context, messageID int64) error {
	err := a.Call(ctx, "delete_msg", map[string]any{
		"message_id": messageID,
	}, nil)
	return err
}

// DeleteMessage 是 Delete 的语义化别名。
func (a *Client) DeleteMessage(ctx context.Context, messageID int64) error {
	return a.Delete(ctx, messageID)
}

// GetLoginInfo 获取登录账号信息。
func (a *Client) GetLoginInfo(ctx context.Context) (any, error) {
	var data any
	err := a.Call(ctx, "get_login_info", nil, &data)
	return data, err
}

// GetStatus 获取 OneBot 实现端状态。
func (a *Client) GetStatus(ctx context.Context) (any, error) {
	var data any
	err := a.Call(ctx, "get_status", nil, &data)
	return data, err
}

// GetFriendList 获取好友列表。注意：这是敏感数据，调用方应避免无必要落盘或打印完整内容。
func (a *Client) GetFriendList(ctx context.Context) (any, error) {
	var data any
	err := a.Call(ctx, "get_friend_list", nil, &data)
	return data, err
}

// GetGroupList 获取群列表。注意：这是敏感数据，调用方应避免无必要落盘或打印完整内容。
func (a *Client) GetGroupList(ctx context.Context) (any, error) {
	var data any
	err := a.Call(ctx, "get_group_list", nil, &data)
	return data, err
}

// GetGroupMemberList 获取群成员列表。注意：这是敏感数据，调用方应避免无必要落盘或打印完整内容。
func (a *Client) GetGroupMemberList(ctx context.Context, groupID int64) (any, error) {
	var data any
	err := a.Call(ctx, "get_group_member_list", map[string]any{
		"group_id": groupID,
	}, &data)
	return data, err
}

// Healthy 表示当前 SDK 是否具备可用传输层。
func (a *Client) Healthy() bool {
	return a.httpURL != "" || a.currentConn() != nil
}

// Call 调用任意 OneBot action；优先使用 HTTP，未配置 HTTP 时使用 WebSocket。
func (a *Client) Call(ctx context.Context, action string, payload any, dst any) error {
	if a.httpURL != "" {
		return a.callHTTP(ctx, action, payload, dst)
	}
	if a.wsURL != "" {
		_, err := a.callWS(ctx, action, payload, dst)
		return err
	}
	return fmt.Errorf("onebot client has no transport configured")
}

// callWS 通过 WebSocket 调用 OneBot action，并使用 echo 匹配响应。
func (a *Client) callWS(ctx context.Context, action string, payload any, dst any) (oneBotWSResponse, error) {
	ctx, cancel := a.withTimeout(ctx)
	defer cancel()

	conn, err := a.waitConn(ctx)
	if err != nil {
		return oneBotWSResponse{}, err
	}

	echo := a.nextEcho(action)
	resultCh := make(chan oneBotWSResult, 1)
	a.registerPending(echo, resultCh)
	defer a.unregisterPending(echo)

	// OneBot WebSocket API 请求格式为 action + params + echo。
	req := oneBotWSRequest{
		Action: action,
		Params: payload,
		Echo:   echo,
	}
	rawReq, err := json.Marshal(req)
	if err != nil {
		return oneBotWSResponse{}, fmt.Errorf("marshal onebot websocket %s request: %w", action, err)
	}

	if err := a.send(conn, rawReq); err != nil {
		a.dropConn(conn)
		return oneBotWSResponse{}, fmt.Errorf("send onebot websocket %s request: %w", action, err)
	}

	select {
	case <-ctx.Done():
		return oneBotWSResponse{}, fmt.Errorf("onebot websocket %s timeout: %w", action, ctx.Err())
	case result := <-resultCh:
		if result.err != nil {
			return oneBotWSResponse{}, result.err
		}
		if result.resp.Status != "ok" || result.resp.RetCode != 0 {
			return oneBotWSResponse{}, &Error{
				Action:  action,
				RetCode: result.resp.RetCode,
				Message: result.resp.Message,
				Wording: result.resp.Wording,
			}
		}
		if dst != nil && len(result.resp.Data) > 0 && string(result.resp.Data) != "null" {
			if err := json.Unmarshal(result.resp.Data, dst); err != nil {
				return oneBotWSResponse{}, fmt.Errorf("decode onebot websocket %s data: %w", action, err)
			}
		}
		return result.resp, nil
	}
}

// connectLoop 维护一个常驻 WebSocket 连接，断线后自动重连。
func (a *Client) connectLoop(ctx context.Context) {
	firstDial := true
	for {
		if ctx.Err() != nil {
			a.closeCurrentConn(fmt.Errorf("onebot websocket client stopped"))
			return
		}

		if a.currentConn() != nil {
			select {
			case <-ctx.Done():
				a.closeCurrentConn(fmt.Errorf("onebot websocket client stopped"))
				return
			case <-a.reconnectCh:
				continue
			}
		}

		if firstDial {
			firstDial = false
		} else {
			a.waitReconnectDelay(ctx)
			if ctx.Err() != nil {
				a.closeCurrentConn(fmt.Errorf("onebot websocket client stopped"))
				return
			}
		}

		conn, err := a.dial()
		if err != nil {
			a.logger.Warn("onebot websocket connect failed",
				"url", a.wsURL,
				"error", err,
				"retry_after", a.reconnectInterval.String(),
			)
			continue
		}

		a.setConn(conn)
	}
}

// waitReconnectDelay 控制重连节奏，避免服务端异常时疯狂重连。
func (a *Client) waitReconnectDelay(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(a.reconnectInterval):
	}
}

// heartbeatLoop 定期调用 get_status，确认 WebSocket 链路仍可用。
func (a *Client) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(a.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, a.timeout)
			err := a.Call(pingCtx, "get_status", nil, nil)
			cancel()
			if err != nil {
				a.logger.Warn("onebot websocket heartbeat failed", "error", err)
				a.closeCurrentConn(fmt.Errorf("onebot websocket heartbeat failed: %w", err))
				continue
			}
			a.logger.Debug("onebot websocket heartbeat ok")
		}
	}
}

// withTimeout 在调用方没有设置 deadline 时补一个默认超时。
func (a *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, a.timeout)
}

// waitConn 等待 WebSocket 连接可用，直到连接成功或上下文超时。
func (a *Client) waitConn(ctx context.Context) (*websocket.Conn, error) {
	for {
		a.mu.Lock()
		if a.conn != nil {
			conn := a.conn
			a.mu.Unlock()
			return conn, nil
		}
		ready := a.ready
		a.mu.Unlock()

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("onebot websocket is not connected: %w", ctx.Err())
		case <-ready:
		}
	}
}

func (a *Client) currentConn() *websocket.Conn {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.conn
}

func (a *Client) dial() (*websocket.Conn, error) {
	wsURL, err := a.urlWithAccessToken()
	if err != nil {
		return nil, err
	}

	cfg, err := websocket.NewConfig(wsURL, "http://localhost/")
	if err != nil {
		return nil, fmt.Errorf("create onebot websocket config: %w", err)
	}
	if a.accessToken != "" {
		cfg.Header = http.Header{}
		cfg.Header.Set("Authorization", "Bearer "+a.accessToken)
	}

	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect onebot websocket %s: %w", wsURL, err)
	}
	return conn, nil
}

func (a *Client) urlWithAccessToken() (string, error) {
	if a.accessToken == "" {
		return a.wsURL, nil
	}

	parsed, err := url.Parse(a.wsURL)
	if err != nil {
		return "", fmt.Errorf("parse onebot websocket url: %w", err)
	}
	query := parsed.Query()
	if query.Get("access_token") == "" {
		query.Set("access_token", a.accessToken)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func (a *Client) setConn(conn *websocket.Conn) {
	a.mu.Lock()
	if a.conn != nil {
		a.mu.Unlock()
		_ = conn.Close()
		return
	}
	a.conn = conn
	select {
	case <-a.ready:
	default:
		close(a.ready)
	}
	a.mu.Unlock()

	a.logger.Info("onebot websocket connected", "url", a.wsURL)
	go a.readLoop(conn)
}

func (a *Client) send(conn *websocket.Conn, rawReq []byte) error {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()

	if err := conn.SetWriteDeadline(time.Now().Add(a.timeout)); err != nil {
		return err
	}
	if err := websocket.Message.Send(conn, string(rawReq)); err != nil {
		return err
	}
	return conn.SetWriteDeadline(time.Time{})
}

func (a *Client) readLoop(conn *websocket.Conn) {
	for {
		var raw string
		if err := websocket.Message.Receive(conn, &raw); err != nil {
			a.failConn(conn, fmt.Errorf("onebot websocket read failed: %w", err))
			return
		}

		rawBytes := []byte(raw)
		resp, ok := decodeOneBotWSResponse(rawBytes)
		if !ok {
			a.dispatchEvent(rawBytes)
			continue
		}
		a.completePending(resp.Echo, oneBotWSResult{resp: resp})
	}
}

func decodeOneBotWSResponse(raw []byte) (oneBotWSResponse, bool) {
	var envelope struct {
		Echo json.RawMessage `json:"echo"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Echo) == 0 {
		return oneBotWSResponse{}, false
	}

	var resp oneBotWSResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return oneBotWSResponse{}, false
	}
	if resp.Echo == "" {
		resp.Echo = strings.Trim(string(envelope.Echo), `"`)
	}
	return resp, resp.Echo != ""
}

func (a *Client) registerPending(echo string, ch chan oneBotWSResult) {
	a.mu.Lock()
	a.pending[echo] = ch
	a.mu.Unlock()
}

func (a *Client) unregisterPending(echo string) {
	a.mu.Lock()
	delete(a.pending, echo)
	a.mu.Unlock()
}

func (a *Client) completePending(echo string, result oneBotWSResult) {
	a.mu.Lock()
	ch := a.pending[echo]
	a.mu.Unlock()
	if ch == nil {
		return
	}

	select {
	case ch <- result:
	default:
	}
}

func (a *Client) failConn(conn *websocket.Conn, err error) {
	a.mu.Lock()
	if a.conn != conn {
		a.mu.Unlock()
		return
	}
	a.conn = nil
	a.ready = make(chan struct{})
	pending := a.pending
	a.pending = make(map[string]chan oneBotWSResult)
	a.mu.Unlock()

	_ = conn.Close()
	a.logger.Warn("onebot websocket disconnected", "error", err)
	a.notifyReconnect()

	for _, ch := range pending {
		select {
		case ch <- oneBotWSResult{err: err}:
		default:
		}
	}
}

func (a *Client) dropConn(conn *websocket.Conn) {
	a.failConn(conn, fmt.Errorf("onebot websocket connection dropped"))
}

func (a *Client) closeCurrentConn(err error) {
	conn := a.currentConn()
	if conn == nil {
		return
	}
	a.failConn(conn, err)
}

func (a *Client) notifyReconnect() {
	select {
	case a.reconnectCh <- struct{}{}:
	default:
	}
}

func (a *Client) nextEcho(action string) string {
	seq := a.echoSeq.Add(1)
	return "bot_msg:" + action + ":" + strconv.FormatInt(time.Now().UnixNano(), 10) + ":" + strconv.FormatUint(seq, 10)
}
