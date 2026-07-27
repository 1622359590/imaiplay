package server

import (
	"net/http"

	"github.com/1622359590/imaiplay/internal/api"
	"github.com/1622359590/imaiplay/internal/config"
	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthService   api.AuthService
	TenantService api.TenantService
	UserService   api.UserService
}

func New(
	cfg config.Config,
	dbCheck func() error,
	deps Dependencies,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health", middleware.Tenant(), func(c *gin.Context) {
		code, source := tenantcontext.TenantFromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"app_name": cfg.AppName,
			"version":  cfg.AppVersion,
			"tenant": gin.H{
				"code":   code,
				"source": source,
			},
		})
	})
	router.GET("/health/db", func(c *gin.Context) {
		if err := dbCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})
	registerRoutes(router, cfg, deps)
	return router
}

func registerRoutes(
	router *gin.Engine,
	cfg config.Config,
	deps Dependencies,
) {
	authHandler := api.NewAuthHandler(deps.AuthService)
	auth := router.Group("/api/v1/auth")
	auth.Use(middleware.Tenant())
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	backend := router.Group("/backend/v1")
	backend.Use(middleware.Tenant(), middleware.Auth(cfg.JWTSecret))
	tenantHandler := api.NewTenantHandler(deps.TenantService)
	backend.POST("/tenants", tenantHandler.Create)
	backend.GET("/tenants", tenantHandler.List)
	backend.GET("/tenants/:id", tenantHandler.Get)
	backend.PUT("/tenants/:id", tenantHandler.Update)
	backend.DELETE("/tenants/:id", tenantHandler.Delete)
	userHandler := api.NewUserHandler(deps.UserService)
	backend.POST("/users", userHandler.Create)
	backend.GET("/users", userHandler.List)
	backend.GET("/users/:id", userHandler.Get)
	backend.PUT("/users/:id", userHandler.Update)
	backend.DELETE("/users/:id", userHandler.Delete)
}

func Run(
	cfg config.Config,
	dbCheck func() error,
	deps Dependencies,
) error {
	return New(cfg, dbCheck, deps).Run(":" + cfg.ServerPort)
}
