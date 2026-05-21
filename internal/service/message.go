package service

import (
	"context"
	"errors"
	"fmt"

	"bot_msg/internal/config"
	"bot_msg/pkg/onebot"
)

var ErrInvalidMessage = errors.New("invalid message")

type Service struct {
	bot    config.BotConfig
	onebot OneBotClient
}

type OneBotClient interface {
	SendPrivate(ctx context.Context, userID int64, msg onebot.Message) (onebot.SendResult, error)
	SendGroup(ctx context.Context, groupID int64, msg onebot.Message) (onebot.SendResult, error)
	GetMessage(ctx context.Context, messageID int64) (any, error)
	Delete(ctx context.Context, messageID int64) error
	GetFriendList(ctx context.Context) (any, error)
	GetGroupList(ctx context.Context) (any, error)
	GetGroupMemberList(ctx context.Context, groupID int64) (any, error)
	Healthy() bool
}

func New(bot config.BotConfig, onebotClient OneBotClient) *Service {
	return &Service{
		bot:    bot,
		onebot: onebotClient,
	}
}

func (s *Service) SendPrivate(ctx context.Context, userID int64, msg onebot.Message) (SendResult, error) {
	if userID <= 0 {
		return SendResult{}, fmt.Errorf("%w: user_id must be greater than zero", ErrInvalidMessage)
	}
	if msg.Empty() {
		return SendResult{}, fmt.Errorf("%w: message is required", ErrInvalidMessage)
	}

	result, err := s.onebot.SendPrivate(ctx, userID, msg)
	if err != nil {
		return SendResult{}, err
	}

	return SendResult{MessageID: result.MessageID}, nil
}

func (s *Service) SendGroup(ctx context.Context, groupID int64, msg onebot.Message) (SendResult, error) {
	if groupID <= 0 {
		return SendResult{}, fmt.Errorf("%w: group_id must be greater than zero", ErrInvalidMessage)
	}
	if msg.Empty() {
		return SendResult{}, fmt.Errorf("%w: message is required", ErrInvalidMessage)
	}

	result, err := s.onebot.SendGroup(ctx, groupID, msg)
	if err != nil {
		return SendResult{}, err
	}

	return SendResult{MessageID: result.MessageID}, nil
}

func (s *Service) GetMessage(ctx context.Context, messageID int64) (any, error) {
	if messageID <= 0 {
		return nil, fmt.Errorf("%w: message_id must be greater than zero", ErrInvalidMessage)
	}
	return s.onebot.GetMessage(ctx, messageID)
}

func (s *Service) DeleteMessage(ctx context.Context, messageID int64) error {
	if messageID <= 0 {
		return fmt.Errorf("%w: message_id must be greater than zero", ErrInvalidMessage)
	}
	return s.onebot.Delete(ctx, messageID)
}

func (s *Service) GetFriendList(ctx context.Context) (any, error) {
	return s.onebot.GetFriendList(ctx)
}

func (s *Service) GetGroupList(ctx context.Context) (any, error) {
	return s.onebot.GetGroupList(ctx)
}

func (s *Service) GetGroupMemberList(ctx context.Context, groupID int64) (any, error) {
	if groupID <= 0 {
		return nil, fmt.Errorf("%w: group_id must be greater than zero", ErrInvalidMessage)
	}
	return s.onebot.GetGroupMemberList(ctx, groupID)
}

func (s *Service) LoginInfo() BotInfo {
	return BotInfo{
		UserID:   s.bot.SelfID,
		Nickname: s.bot.Nickname,
	}
}

func (s *Service) Status() Status {
	healthy := s.onebot.Healthy()
	return Status{
		Online: healthy,
		Good:   healthy,
	}
}

func (s *Service) VersionInfo() VersionInfo {
	return VersionInfo{
		AppName:         "bot_msg",
		AppVersion:      "0.1.0",
		ProtocolVersion: "v11",
	}
}
