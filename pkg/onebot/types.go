package onebot

import (
	"encoding/json"
	"fmt"
)

type SendResult struct {
	MessageID int64
}

type Error struct {
	Action  string
	RetCode int
	Message string
	Wording string
}

func (e *Error) Error() string {
	return fmt.Sprintf("onebot action %s failed: retcode=%d message=%s wording=%s", e.Action, e.RetCode, e.Message, e.Wording)
}

type oneBotWSRequest struct {
	Action string `json:"action"`
	Params any    `json:"params,omitempty"`
	Echo   string `json:"echo"`
}

type oneBotWSResponse struct {
	Status  string          `json:"status"`
	RetCode int             `json:"retcode"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Wording string          `json:"wording"`
	Echo    string          `json:"echo"`
}

type oneBotWSResult struct {
	resp oneBotWSResponse
	err  error
}

type oneBotSendData struct {
	MessageID int64 `json:"message_id"`
}
