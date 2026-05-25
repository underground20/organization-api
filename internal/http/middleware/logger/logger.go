package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path

		logger.Info("HTTP request",
			slog.String("method", method),
			slog.Int("status", statusCode),
			slog.String("path", path),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
		)
	}
}
