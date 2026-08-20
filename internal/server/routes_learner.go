package server

import (
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerLearnerRoutes(router *gin.Engine, cfg config.Config, deps Dependencies, h routeHandlers) {
	student := router.Group("/api/v1")
	student.Use(middleware.TenantWithRepositoryForPlatformHost(deps.TenantRepository, cfg.AdminHost), middleware.Auth(cfg.JWTSecret), middleware.TenantMatch(deps.TenantRepository), middleware.TenantAccess(deps.TenantRepository))
	student.GET("/courses", h.course.PublishedList)
	student.GET("/courses/:id", h.course.PublishedDetail)
	student.GET("/course-materials/:id/download", h.material.Download)
	student.GET("/resources/:id/file", h.resource.File)
	student.GET("/resources/:id/playback-url", h.resource.PlaybackURL)
	student.POST("/lessons/:id/progress", h.progress.Report)
	student.GET("/lessons/:id/progress", h.progress.Get)
	student.GET("/learner/overview", h.learnerOverview.Get)
	student.GET("/learner/motivation", h.learnerMotivation.Get)
	student.POST("/learner/motivation/ack", h.learnerMotivation.Acknowledge)
	student.GET("/recent-learning", h.learnerOverview.Recent)
	router.GET("/api/v1/resource-playback/:id", h.resource.Playback)
}
