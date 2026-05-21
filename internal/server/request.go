package server

import "bot_msg/pkg/onebot"

type sendPrivateMsgRequest struct {
	UserID     int64          `json:"user_id" form:"user_id"`
	Message    onebot.Message `json:"message" form:"message"`
	AutoEscape bool           `json:"auto_escape" form:"auto_escape"`
	Echo       any            `json:"echo" form:"echo"`
}

type sendGroupMsgRequest struct {
	GroupID    int64          `json:"group_id" form:"group_id"`
	Message    onebot.Message `json:"message" form:"message"`
	AutoEscape bool           `json:"auto_escape" form:"auto_escape"`
	Echo       any            `json:"echo" form:"echo"`
}

type sendMsgRequest struct {
	MessageType string         `json:"message_type" form:"message_type"`
	UserID      int64          `json:"user_id" form:"user_id"`
	GroupID     int64          `json:"group_id" form:"group_id"`
	Message     onebot.Message `json:"message" form:"message"`
	AutoEscape  bool           `json:"auto_escape" form:"auto_escape"`
	Echo        any            `json:"echo" form:"echo"`
}

type messageIDRequest struct {
	MessageID int64 `json:"message_id" form:"message_id"`
	Echo      any   `json:"echo" form:"echo"`
}

type groupIDRequest struct {
	GroupID int64 `json:"group_id" form:"group_id"`
	Echo    any   `json:"echo" form:"echo"`
}
