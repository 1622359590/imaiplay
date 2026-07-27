package api

import (
	"net/http"
	"strconv"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0, "message": "ok", "data": data,
	})
}

func requireHandlerRole(c *gin.Context, expected string) bool {
	return requireHandlerRoles(c, expected)
}

func requireHandlerRoles(c *gin.Context, expected ...string) bool {
	_, _, _, role, ok := usercontext.UserFromContext(c.Request.Context())
	if ok {
		for _, candidate := range expected {
			if role == candidate {
				return true
			}
		}
	}
	errorsx.GinResponse(c, errorsx.Forbidden("permission denied"))
	return false
}

func paginationQuery(c *gin.Context) (int, int, error) {
	offset, err := queryInt(c, "offset", 0)
	if err != nil || offset < 0 {
		return 0, 0, errorsx.BadRequest("invalid offset")
	}
	limit, err := queryInt(c, "limit", 20)
	if err != nil || limit <= 0 || limit > 100 {
		return 0, 0, errorsx.BadRequest("invalid limit")
	}
	return offset, limit, nil
}

func queryInt(c *gin.Context, key string, fallback int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
