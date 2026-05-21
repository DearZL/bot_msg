package server

import (
	"errors"
	stdhttp "net/http"

	"bot_msg/internal/service"
	"bot_msg/pkg/onebot"

	"github.com/gin-gonic/gin"
)

// healthz godoc
// @Summary Health check
// @Description Returns service health.
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (s *Server) healthz(c *gin.Context) {
	c.JSON(stdhttp.StatusOK, gin.H{
		"status": "ok",
	})
}

func (s *Server) noRoute(c *gin.Context) {
	s.writeFailed(c, stdhttp.StatusNotFound, onebot.RetUnsupported, "unsupported path: "+c.Request.URL.Path, nil)
}

// sendPrivateMsg godoc
// @Summary Send private message
// @Description Sends a private message through the configured OneBot v11 endpoint.
// @Tags messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body sendPrivateMsgRequest true "Private message payload"
// @Success 200 {object} onebot.Response{data=service.SendResult}
// @Failure 400 {object} onebot.ErrorResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /send_private_msg [post]
func (s *Server) sendPrivateMsg(c *gin.Context) {
	var req sendPrivateMsgRequest
	if !s.bind(c, &req, &req.Echo) {
		return
	}

	result, err := s.service.SendPrivate(c.Request.Context(), req.UserID, req.Message)
	if err != nil {
		s.writeServiceError(c, err, req.Echo)
		return
	}
	s.writeOK(c, result, req.Echo)
}

// sendGroupMsg godoc
// @Summary Send group message
// @Description Sends a group message through the configured OneBot v11 endpoint.
// @Tags messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body sendGroupMsgRequest true "Group message payload"
// @Success 200 {object} onebot.Response{data=service.SendResult}
// @Failure 400 {object} onebot.ErrorResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /send_group_msg [post]
func (s *Server) sendGroupMsg(c *gin.Context) {
	var req sendGroupMsgRequest
	if !s.bind(c, &req, &req.Echo) {
		return
	}

	result, err := s.service.SendGroup(c.Request.Context(), req.GroupID, req.Message)
	if err != nil {
		s.writeServiceError(c, err, req.Echo)
		return
	}
	s.writeOK(c, result, req.Echo)
}

// sendMsg godoc
// @Summary Send message
// @Description Sends a private or group message. If message_type is omitted, group_id takes precedence.
// @Tags messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body sendMsgRequest true "Message payload"
// @Success 200 {object} onebot.Response{data=service.SendResult}
// @Failure 400 {object} onebot.ErrorResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /send_msg [post]
func (s *Server) sendMsg(c *gin.Context) {
	var req sendMsgRequest
	if !s.bind(c, &req, &req.Echo) {
		return
	}

	switch req.MessageType {
	case "private":
		result, err := s.service.SendPrivate(c.Request.Context(), req.UserID, req.Message)
		if err != nil {
			s.writeServiceError(c, err, req.Echo)
			return
		}
		s.writeOK(c, result, req.Echo)
	case "group":
		result, err := s.service.SendGroup(c.Request.Context(), req.GroupID, req.Message)
		if err != nil {
			s.writeServiceError(c, err, req.Echo)
			return
		}
		s.writeOK(c, result, req.Echo)
	default:
		if req.GroupID > 0 {
			result, err := s.service.SendGroup(c.Request.Context(), req.GroupID, req.Message)
			if err != nil {
				s.writeServiceError(c, err, req.Echo)
				return
			}
			s.writeOK(c, result, req.Echo)
			return
		}
		result, err := s.service.SendPrivate(c.Request.Context(), req.UserID, req.Message)
		if err != nil {
			s.writeServiceError(c, err, req.Echo)
			return
		}
		s.writeOK(c, result, req.Echo)
	}
}

// deleteMsg godoc
// @Summary Delete message
// @Description Deletes a message by message_id.
// @Tags messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body messageIDRequest true "Message ID payload"
// @Success 200 {object} onebot.EmptyResponse
// @Failure 400 {object} onebot.ErrorResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /delete_msg [post]
func (s *Server) deleteMsg(c *gin.Context) {
	var req messageIDRequest
	if !s.bind(c, &req, &req.Echo) {
		return
	}

	if err := s.service.DeleteMessage(c.Request.Context(), req.MessageID); err != nil {
		s.writeServiceError(c, err, req.Echo)
		return
	}
	s.writeOK(c, gin.H{}, req.Echo)
}

// getMsg godoc
// @Summary Get message
// @Description Gets a message by message_id.
// @Tags messages
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body messageIDRequest true "Message ID payload"
// @Success 200 {object} onebot.AnyResponse
// @Failure 400 {object} onebot.ErrorResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /get_msg [post]
func (s *Server) getMsg(c *gin.Context) {
	var req messageIDRequest
	if !s.bind(c, &req, &req.Echo) {
		return
	}

	record, err := s.service.GetMessage(c.Request.Context(), req.MessageID)
	if err != nil {
		s.writeServiceError(c, err, req.Echo)
		return
	}
	s.writeOK(c, record, req.Echo)
}

// getFriendList godoc
// @Summary Get friend list
// @Description Gets the friend list from the configured OneBot v11 endpoint.
// @Tags lists
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} onebot.AnyResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /get_friend_list [get]
func (s *Server) getFriendList(c *gin.Context) {
	data, err := s.service.GetFriendList(c.Request.Context())
	if err != nil {
		s.writeServiceError(c, err, nil)
		return
	}
	s.writeOK(c, data, nil)
}

// getGroupList godoc
// @Summary Get group list
// @Description Gets the group list from the configured OneBot v11 endpoint.
// @Tags lists
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} onebot.AnyResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /get_group_list [get]
func (s *Server) getGroupList(c *gin.Context) {
	data, err := s.service.GetGroupList(c.Request.Context())
	if err != nil {
		s.writeServiceError(c, err, nil)
		return
	}
	s.writeOK(c, data, nil)
}

// getGroupMemberList godoc
// @Summary Get group member list
// @Description Gets group members by group_id.
// @Tags lists
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body groupIDRequest true "Group ID payload"
// @Success 200 {object} onebot.AnyResponse
// @Failure 400 {object} onebot.ErrorResponse
// @Failure 500 {object} onebot.ErrorResponse
// @Router /get_group_member_list [post]
func (s *Server) getGroupMemberList(c *gin.Context) {
	var req groupIDRequest
	if !s.bind(c, &req, &req.Echo) {
		return
	}

	data, err := s.service.GetGroupMemberList(c.Request.Context(), req.GroupID)
	if err != nil {
		s.writeServiceError(c, err, req.Echo)
		return
	}
	s.writeOK(c, data, req.Echo)
}

func (s *Server) bind(c *gin.Context, dst any, echo *any) bool {
	var err error
	if c.Request.Method == stdhttp.MethodGet {
		err = c.ShouldBindQuery(dst)
	} else {
		err = c.ShouldBind(dst)
	}
	if err != nil {
		s.writeFailed(c, stdhttp.StatusBadRequest, onebot.RetBadParam, err.Error(), derefEcho(echo))
		return false
	}
	return true
}

func (s *Server) writeServiceError(c *gin.Context, err error, echo any) {
	switch {
	case errors.Is(err, service.ErrInvalidMessage):
		s.writeFailed(c, stdhttp.StatusBadRequest, onebot.RetBadParam, err.Error(), echo)
	default:
		var onebotErr *onebot.Error
		if errors.As(err, &onebotErr) {
			message := onebotErr.Message
			if message == "" {
				message = onebotErr.Wording
			}
			s.writeFailed(c, stdhttp.StatusOK, onebotErr.RetCode, message, echo)
			return
		}
		s.logger.Error("action failed", "error", err)
		s.writeFailed(c, stdhttp.StatusInternalServerError, onebot.RetInternal, "internal server error", echo)
	}
}

func (s *Server) writeOK(c *gin.Context, data any, echo any) {
	status, payload := onebot.OK(data, echo)
	c.JSON(status, payload)
}

func (s *Server) writeFailed(c *gin.Context, httpStatus int, retCode int, message string, echo any) {
	status, payload := onebot.Failed(httpStatus, retCode, message, echo)
	c.JSON(status, payload)
}

func derefEcho(echo *any) any {
	if echo == nil {
		return nil
	}
	return *echo
}
