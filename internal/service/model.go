package service

type SendResult struct {
	MessageID int64 `json:"message_id" example:"100001"`
}

type BotInfo struct {
	UserID   int64  `json:"user_id" example:"10000"`
	Nickname string `json:"nickname" example:"bot_msg"`
}

type Status struct {
	Online bool `json:"online" example:"true"`
	Good   bool `json:"good" example:"true"`
}

type VersionInfo struct {
	AppName         string `json:"app_name" example:"bot_msg"`
	AppVersion      string `json:"app_version" example:"0.1.0"`
	ProtocolVersion string `json:"protocol_version" example:"v11"`
}
