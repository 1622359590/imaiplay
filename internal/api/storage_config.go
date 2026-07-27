package api

import (
	"context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/storage"
	"github.com/gin-gonic/gin"
)

type StorageConfigService interface {
	Config() storage.Config
	Save(storage.Config) error
	Test(context.Context, storage.Config) error
}
type StorageConfigHandler struct{ service StorageConfigService }

func NewStorageConfigHandler(service StorageConfigService) *StorageConfigHandler {
	return &StorageConfigHandler{service: service}
}
func (h *StorageConfigHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	success(c, h.service.Config())
}
func (h *StorageConfigHandler) Save(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var req storage.Config
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := h.service.Save(req); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest(err.Error()))
		return
	}
	success(c, h.service.Config())
}
func (h *StorageConfigHandler) Test(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var req storage.Config
	if err := c.ShouldBindJSON(&req); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := h.service.Test(c.Request.Context(), req); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("storage connection failed"))
		return
	}
	success(c, gin.H{"ok": true})
}
