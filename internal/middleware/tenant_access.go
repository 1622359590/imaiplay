package middleware

import (
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
)

func TenantAccess(tenants repository.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tenants == nil {
			c.Next()
			return
		}
		_, tenantID, _, role, ok := usercontext.UserFromContext(c.Request.Context())
		if !ok || role == "superadmin" {
			c.Next()
			return
		}
		tenant, err := tenants.FindByID(c.Request.Context(), tenantID)
		if err != nil {
			errorsx.GinResponse(c, errorsx.Forbidden("tenant is unavailable"))
			c.Abort()
			return
		}
		if accessible, reason := tenantAccessible(tenant, time.Now().UTC()); !accessible {
			errorsx.GinResponse(c, errorsx.Forbidden(reason))
			c.Abort()
			return
		}
		c.Next()
	}
}

func tenantAccessible(tenant *domain.Tenant, now time.Time) (bool, string) {
	status := tenant.LifecycleStatus
	if status == "" {
		if tenant.Status == 0 {
			return false, "tenant is suspended"
		}
		return true, ""
	}
	if status == "suspended" || status == "deleted" {
		return false, "tenant is " + status
	}
	if status == "trial" && tenant.TrialEndsAt != nil && !now.Before(*tenant.TrialEndsAt) {
		return false, "tenant trial expired"
	}
	return true, ""
}
