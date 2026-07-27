package middleware

import (
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/gin-gonic/gin"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			errorsx.GinResponse(c, errorsx.Unauthorized("missing or invalid token"))
			c.Abort()
			return
		}
		claims, err := security.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			errorsx.GinResponse(c, errorsx.Unauthorized("missing or invalid token"))
			c.Abort()
			return
		}
		ctx := usercontext.WithUser(
			c.Request.Context(),
			claims.UserID,
			claims.TenantID,
			claims.Email,
			claims.Role,
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
