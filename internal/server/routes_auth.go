package server

import (
	"time"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(router *gin.Engine, cfg config.Config, deps Dependencies, h routeHandlers) {
	router.GET("/api/v1/portal", h.portal.Get)
	portalSession := router.Group("/api/v1/portal")
	portalSession.Use(middleware.Auth(cfg.JWTSecret))
	portalSession.GET("/session", h.portal.GetSession)
	theme := router.Group("/api/v1")
	theme.Use(middleware.TenantWithRepository(deps.TenantRepository))
	theme.GET("/theme", h.theme.Get)
	router.POST("/api/v1/bootstrap/superadmin", h.auth.BootstrapSuperadmin)
	auth := router.Group("/api/v1/auth")
	auth.Use(middleware.TenantWithRepositoryForAdminHost(deps.TenantRepository, cfg.AdminHost))
	limiter := middleware.NewRateLimiter(cfg.AuthRateLimit, time.Duration(cfg.AuthRateWindowSeconds)*time.Second)
	auth.POST("/register", limiter.Handler(), h.auth.Register)
	auth.POST("/login", limiter.Handler(), h.auth.Login)
	auth.POST("/select-tenant", limiter.Handler(), h.auth.SelectTenant)
	auth.POST("/login-code/send", limiter.Handler(), h.auth.SendLoginCode)
	auth.POST("/login-code", limiter.Handler(), h.auth.LoginWithCode)
	auth.POST("/refresh", h.auth.Refresh)
	auth.POST("/forgot-password", limiter.Handler(), h.auth.ForgotPassword)
	auth.POST("/reset-password", h.auth.ResetPassword)
	authProtected := router.Group("/api/v1/auth")
	authProtected.Use(middleware.TenantWithRepositoryForAdminHost(deps.TenantRepository, cfg.AdminHost), middleware.Auth(cfg.JWTSecret), middleware.TenantMatch(deps.TenantRepository), middleware.TenantAccess(deps.TenantRepository))
	authProtected.GET("/me", h.auth.Me)
	authProtected.POST("/logout", h.auth.Logout)
	registration := router.Group("/api/v1/tenants")
	registration.Use(middleware.TenantWithRepository(deps.TenantRepository))
	registration.POST("/register", limiter.Handler(), h.registration.Register)
}
