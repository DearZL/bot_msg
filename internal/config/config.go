package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"bot_msg/pkg/onebot"
)

type Config struct {
	// HTTP 是 bot_msg 服务自身暴露给调用方的 HTTP 配置。
	HTTP HTTPConfig
	// Bot 是服务本地展示用的机器人信息，不代表 OneBot 实现端一定返回相同信息。
	Bot BotConfig
	// OneBot 是 SDK 配置，负责连接 OneBot HTTP/WS 实现端。
	OneBot onebot.Config
	// LogLevel 控制服务日志级别。
	LogLevel slog.Level
}

type HTTPConfig struct {
	// Addr 是本服务监听地址。
	Addr string
	// AccessToken 是调用本服务 API 时的鉴权 token。
	AccessToken string
}

type BotConfig struct {
	// SelfID 是本服务展示的机器人账号。
	SelfID int64
	// Nickname 是本服务展示的机器人昵称。
	Nickname string
}

// Load 从环境变量加载服务配置。
func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr:        getenv("BOTMSG_HTTP_ADDR", ":8080"),
			AccessToken: os.Getenv("BOTMSG_ACCESS_TOKEN"),
		},
		Bot: BotConfig{
			SelfID:   getenvInt64("BOTMSG_SELF_ID", 10000),
			Nickname: getenv("BOTMSG_NICKNAME", "bot_msg"),
		},
		OneBot: onebot.Config{
			HTTPURL:           strings.TrimRight(os.Getenv("BOTMSG_ONEBOT_HTTP_URL"), "/"),
			WebSocketURL:      loadOneBotWSURL(),
			AccessToken:       os.Getenv("BOTMSG_ONEBOT_ACCESS_TOKEN"),
			Timeout:           time.Duration(getenvInt64("BOTMSG_ONEBOT_TIMEOUT_SECONDS", 10)) * time.Second,
			HeartbeatInterval: time.Duration(getenvInt64("BOTMSG_ONEBOT_HEARTBEAT_SECONDS", 30)) * time.Second,
			ReconnectInterval: time.Duration(getenvInt64("BOTMSG_ONEBOT_RECONNECT_SECONDS", 5)) * time.Second,
		},
		LogLevel: parseLogLevel(getenv("BOTMSG_LOG_LEVEL", "info")),
	}
}

// loadOneBotWSURL 解析 OneBot WebSocket 地址。
// 如果显式配置了 HTTP 地址且没有配置 WS 地址，则不启用 WS 后台连接。
func loadOneBotWSURL() string {
	if value := strings.TrimSpace(os.Getenv("BOTMSG_ONEBOT_WS_URL")); value != "" {
		return value
	}
	if strings.TrimSpace(os.Getenv("BOTMSG_ONEBOT_HTTP_URL")) != "" {
		return ""
	}

	if value := strings.TrimSpace(os.Getenv("BOTMSG_ONEBOT_WS_ENDPOINT")); value != "" {
		endpoint := value
		value = strings.TrimPrefix(value, "http://")
		value = strings.TrimPrefix(value, "https://")
		if strings.HasPrefix(endpoint, "https://") {
			return "wss://" + value
		}
		return "ws://" + value
	}

	return "ws://127.0.0.1:3001/api"
}

// getenv 读取字符串环境变量，空值时返回默认值。
func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// getenvInt64 读取 int64 环境变量，解析失败时返回默认值。
func getenvInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

// parseLogLevel 将字符串日志级别转换为 slog.Level。
func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
