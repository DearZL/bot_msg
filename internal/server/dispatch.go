package server

import (
	_ "bot_msg/internal/service"
	_ "bot_msg/pkg/onebot"

	"github.com/gin-gonic/gin"
)

// getLoginInfo godoc
// @Summary Get login info
// @Description Gets configured bot login information exposed by this service.
// @Tags system
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} onebot.Response{data=service.BotInfo}
// @Router /get_login_info [get]
func (s *Server) getLoginInfo(c *gin.Context) {
	s.writeOK(c, s.service.LoginInfo(), nil)
}

// getStatus godoc
// @Summary Get status
// @Description Gets current service status. In WS mode, status reflects whether the OneBot WebSocket is connected.
// @Tags system
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} onebot.Response{data=service.Status}
// @Router /get_status [get]
func (s *Server) getStatus(c *gin.Context) {
	s.writeOK(c, s.service.Status(), nil)
}

// getVersionInfo godoc
// @Summary Get version info
// @Description Gets service version and protocol information.
// @Tags system
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} onebot.Response{data=service.VersionInfo}
// @Router /get_version_info [get]
func (s *Server) getVersionInfo(c *gin.Context) {
	s.writeOK(c, s.service.VersionInfo(), nil)
}
