package server

import (
	"net/http"

	"github.com/1622359590/imaiplay/internal/config"
	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

func New(cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.Tenant())
	router.GET("/health", func(c *gin.Context) {
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
	return router
}

func Run(cfg config.Config) error {
	return New(cfg).Run(":" + cfg.ServerPort)
}
