package middleware

import (
	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
)

func TenantMatch(tenants repository.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tenants == nil {
			c.Next()
			return
		}
		_, sessionTenantID, _, role, authenticated :=
			usercontext.UserFromContext(c.Request.Context())
		portalCode, _ := usercontext.TenantFromContext(c.Request.Context())
		if !authenticated || role == "superadmin" ||
			portalCode == usercontext.UnknownTenant {
			c.Next()
			return
		}
		portalTenant, err := tenants.FindByCode(
			c.Request.Context(),
			portalCode,
		)
		if err != nil || portalTenant.ID != sessionTenantID {
			errorsx.GinResponse(
				c,
				errorsx.Forbidden("tenant context does not match session"),
			)
			c.Abort()
			return
		}
		c.Next()
	}
}
