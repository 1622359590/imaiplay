package server

import (
	"github.com/1622359590/imaiplay/internal/api"
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/gin-gonic/gin"
)

func registerInfrastructureRoutes(router *gin.Engine, backend *gin.RouterGroup, _ config.Config, deps Dependencies, h routeHandlers) {
	backend.POST("/resources/upload", h.resource.Upload)
	backend.POST("/resources/attachments/upload", h.resource.UploadAttachment)
	backend.GET("/resources", h.resource.List)
	backend.POST("/admin/resources/upload", h.resource.UploadPlatform)
	backend.POST("/admin/resources/attachments/upload", h.resource.UploadPlatformAttachment)
	backend.GET("/admin/resources", h.resource.ListPlatform)
	backend.GET("/admin/resources/:id/file", h.resource.File)
	backend.DELETE("/admin/resources/:id", h.resource.DeletePlatform)
	backend.GET("/resources/:id/file", h.resource.File)
	backend.DELETE("/resources/:id", h.resource.Delete)
	router.GET("/api/v1/platform-covers/:id", h.resource.PlatformCover)
	backend.POST("/resource-categories", h.resourceCategory.Create)
	backend.GET("/resource-categories", h.resourceCategory.List)
	backend.PUT("/resource-categories/:id", h.resourceCategory.Update)
	backend.DELETE("/resource-categories/:id", h.resourceCategory.Delete)
	backend.GET("/course-categories", h.courseCategory.List)
	backend.POST("/course-categories", h.courseCategory.Create)
	backend.PUT("/course-categories/:id", h.courseCategory.Update)
	backend.DELETE("/course-categories/:id", h.courseCategory.Delete)
	backend.GET("/admin/course-categories", h.courseCategory.ListPlatform)
	backend.POST("/admin/course-categories", h.courseCategory.CreatePlatform)
	backend.PUT("/admin/course-categories/:id", h.courseCategory.UpdatePlatform)
	backend.DELETE("/admin/course-categories/:id", h.courseCategory.DeletePlatform)
	if deps.SMSConfigService != nil {
		sms := api.NewSMSConfigHandler(deps.SMSConfigService)
		backend.GET("/sms-config", sms.Get)
		backend.PUT("/sms-config", sms.Save)
		backend.POST("/sms-config/test", sms.Test)
	}
	if deps.StorageConfigService != nil {
		storage := api.NewStorageConfigHandler(deps.StorageConfigService)
		backend.GET("/storage-config", storage.Get)
		backend.PUT("/storage-config", storage.Save)
		backend.POST("/storage-config/test", storage.Test)
	}
	backend.GET("/audit-logs", h.audit.ListTenant)
	backend.GET("/admin/audit-logs", h.audit.ListAdmin)
}
