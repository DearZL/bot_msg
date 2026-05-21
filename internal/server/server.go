package server

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"bot_msg/internal/config"
	"bot_msg/internal/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg     config.HTTPConfig
	service *service.Service
	logger  *slog.Logger
}

func New(cfg config.HTTPConfig, service *service.Service, logger *slog.Logger) *stdhttp.Server {
	gin.SetMode(gin.ReleaseMode)

	api := &Server{
		cfg:     cfg,
		service: service,
		logger:  logger,
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(api.requestLogger())
	router.Use(api.auth())

	router.GET("/healthz", api.healthz)
	api.registerActionRoutes(router)
	router.NoRoute(api.noRoute)

	return &stdhttp.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// registerActionRoutes 同时注册根路径和 /api 前缀，兼容常见 OneBot HTTP 调用方式。
func (s *Server) registerActionRoutes(router *gin.Engine) {
	for _, route := range s.actionRoutes() {
		router.Any("/"+route.action, route.handler)
		router.Any("/api/"+route.action, route.handler)
	}
}
