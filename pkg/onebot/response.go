package onebot

import "net/http"

type Response struct {
	Status  string `json:"status"`
	RetCode int    `json:"retcode"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Wording string `json:"wording,omitempty"`
	Echo    any    `json:"echo,omitempty"`
}

type EmptyData struct{}

type SendMessageResponse struct {
	Status  string `json:"status" example:"ok"`
	RetCode int    `json:"retcode" example:"0"`
	Data    struct {
		MessageID int64 `json:"message_id" example:"100001"`
	} `json:"data"`
	Echo any `json:"echo,omitempty"`
}

type EmptyResponse struct {
	Status  string    `json:"status" example:"ok"`
	RetCode int       `json:"retcode" example:"0"`
	Data    EmptyData `json:"data"`
	Echo    any       `json:"echo,omitempty"`
}

type AnyResponse struct {
	Status  string `json:"status" example:"ok"`
	RetCode int    `json:"retcode" example:"0"`
	Data    any    `json:"data,omitempty"`
	Echo    any    `json:"echo,omitempty"`
}

type ErrorResponse struct {
	Status  string `json:"status" example:"failed"`
	RetCode int    `json:"retcode" example:"10001"`
	Message string `json:"message,omitempty"`
	Wording string `json:"wording,omitempty"`
	Echo    any    `json:"echo,omitempty"`
}

const (
	RetOK          = 0
	RetBadParam    = 10001
	RetNotFound    = 10003
	RetUnsupported = 10004
	RetInternal    = 20000
)

func OK(data any, echo any) (int, Response) {
	return http.StatusOK, Response{
		Status:  "ok",
		RetCode: RetOK,
		Data:    data,
		Echo:    echo,
	}
}

func Failed(httpStatus int, retCode int, message string, echo any) (int, Response) {
	return httpStatus, Response{
		Status:  "failed",
		RetCode: retCode,
		Message: message,
		Wording: message,
		Echo:    echo,
	}
}
