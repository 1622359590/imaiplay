package middleware

import (
	"log/slog"
	"os"
	"strings"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/gin-gonic/gin"
)

func NewLogger(level, format string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	options := &slog.HandlerOptions{Level: slogLevel}
	if strings.EqualFold(strings.TrimSpace(format), "text") {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}

func Logging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		userID, tenantID, _, _, _ := usercontext.UserFromContext(c.Request.Context())
		logger.Info("http request",
			slog.String("request_id", RequestIDFromContext(c.Request.Context())),
			slog.String("method", c.Request.Method), slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()), slog.Duration("duration", time.Since(started)),
			slog.String("client_ip", c.ClientIP()), slog.String("tenant_id", tenantID), slog.String("user_id", userID),
		)
	}
}

func PanicLogging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http panic", slog.String("request_id", RequestIDFromContext(c.Request.Context())), slog.Any("panic", recovered))
				panic(recovered)
			}
		}()
		c.Next()
	}
}
