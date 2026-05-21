package server

import (
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		s.logger.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func (s *Server) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.cfg.AccessToken == "" || c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		token := c.Query("access_token")
		if token == "" {
			token = c.GetHeader("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
		}

		if token != s.cfg.AccessToken {
			c.AbortWithStatusJSON(stdhttp.StatusUnauthorized, gin.H{
				"status":  "failed",
				"retcode": 10002,
				"message": "invalid access token",
				"wording": "invalid access token",
			})
			return
		}

		c.Next()
	}
}
