package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AdminHost(allowedHost string) gin.HandlerFunc {
	allowedHost = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(allowedHost)), ".")
	return func(c *gin.Context) {
		if allowedHost == "" {
			c.Next()
			return
		}
		if requestHost(c.Request.Host) != allowedHost {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}
