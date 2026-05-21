package onebot

import (
	"io"
	"log/slog"
	"time"
)

type Config struct {
	// HTTPURL 是 OneBot HTTP/HTTPS API 地址，设置后普通 API 调用会优先走 HTTP。
	HTTPURL string
	// WebSocketURL 是 OneBot 正向 WS/WSS 地址，设置后可用于 API 调用和事件接收。
	WebSocketURL string
	// AccessToken 会同时用于 HTTP Authorization 头和 WebSocket 鉴权参数。
	AccessToken string
	// Timeout 控制单次 API 调用的最长等待时间。
	Timeout time.Duration
	// HeartbeatInterval 控制 WebSocket 模式下的心跳检测间隔。
	HeartbeatInterval time.Duration
	// ReconnectInterval 控制 WebSocket 断线后的重连间隔。
	ReconnectInterval time.Duration
	// EventBuffer 是事件通道缓冲大小，业务消费慢时超过该大小会丢弃事件。
	EventBuffer int
	// Logger 用于 SDK 内部日志，默认丢弃日志。
	Logger *slog.Logger
}

type Option func(*Config)

// DefaultConfig 返回 SDK 默认配置。
func DefaultConfig() Config {
	return Config{
		Timeout:           10 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		ReconnectInterval: 5 * time.Second,
		EventBuffer:       128,
		Logger:            slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// WithHTTP 配置 OneBot HTTP/HTTPS API 地址。
func WithHTTP(endpoint string) Option {
	return func(cfg *Config) {
		cfg.HTTPURL = endpoint
	}
}

// WithWebSocket 配置 OneBot 正向 WS/WSS 地址。
func WithWebSocket(endpoint string) Option {
	return func(cfg *Config) {
		cfg.WebSocketURL = endpoint
	}
}

// WithAccessToken 配置 OneBot 访问令牌。
func WithAccessToken(token string) Option {
	return func(cfg *Config) {
		cfg.AccessToken = token
	}
}

// WithTimeout 配置单次 API 调用超时时间。
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = timeout
	}
}

// WithHeartbeatInterval 配置 WebSocket 心跳间隔。
func WithHeartbeatInterval(interval time.Duration) Option {
	return func(cfg *Config) {
		cfg.HeartbeatInterval = interval
	}
}

// WithReconnectInterval 配置 WebSocket 重连间隔。
func WithReconnectInterval(interval time.Duration) Option {
	return func(cfg *Config) {
		cfg.ReconnectInterval = interval
	}
}

// WithEventBuffer 配置事件通道缓冲大小。
func WithEventBuffer(size int) Option {
	return func(cfg *Config) {
		cfg.EventBuffer = size
	}
}

// WithLogger 配置 SDK 日志输出。
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}
