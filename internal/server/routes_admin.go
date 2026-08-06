package server

import (
	"github.com/1622359590/imaiplay/internal/api"
	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(backend *gin.RouterGroup, deps Dependencies, h routeHandlers) {
	backend.GET("/plans", h.plan.List)
	backend.POST("/plans", h.plan.Create)
	backend.PUT("/plans/:id", h.plan.Update)
	backend.DELETE("/plans/:id", h.plan.Delete)
	backend.PUT("/tenant-plans/:tenantID", h.plan.Assign)
	backend.GET("/plan/current", h.plan.Current)
	backend.GET("/theme", h.theme.Get)
	backend.PUT("/theme", h.theme.Update)
	backend.POST("/tenants", h.tenant.Create)
	backend.POST("/admin/tenants", h.registration.CreateForSuperadmin)
	backend.GET("/tenants", h.tenant.List)
	backend.GET("/tenants/:id", h.tenant.Get)
	backend.PUT("/tenants/:id", h.tenant.Update)
	backend.DELETE("/tenants/:id", h.tenant.Delete)
	if deps.DomainBindService != nil {
		domain := api.NewDomainBindHandler(deps.DomainBindService)
		backend.POST("/domain-bind/verify", domain.Verify)
		backend.POST("/domain-bind", domain.Bind)
		backend.GET("/domain-bind/status", domain.Status)
		backend.DELETE("/domain-bind", domain.Unbind)
		backend.POST("/tenants/:id/domain-bind/verify", domain.VerifyForTenant)
		backend.POST("/tenants/:id/domain-bind", domain.BindForTenant)
		backend.GET("/tenants/:id/domain-bind/status", domain.StatusForTenant)
		backend.DELETE("/tenants/:id/domain-bind", domain.UnbindForTenant)
	}
	backend.POST("/users", h.user.Create)
	backend.GET("/users", h.user.List)
	backend.GET("/users/:id", h.user.Get)
	backend.PUT("/users/:id", h.user.Update)
	backend.DELETE("/users/:id", h.user.Delete)
	backend.PUT("/users/:id/password", h.user.ResetTenantAdminPassword)
	backend.GET("/dashboard", h.dashboard.Get)
	backend.DELETE("/tenants/demo-data", h.registration.ClearDemoData)
}
