package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type requestIDKey struct{}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if id == "" {
			id = uuid.NewString()
		}
		c.Header("X-Request-ID", id)
		c.Request = c.Request.WithContext(WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
