package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type ResourceCategoryService interface {
	Create(
		ctx context.Context, name string, parentID *string,
	) (*domain.ResourceCategory, error)
	List(ctx context.Context) ([]domain.ResourceCategory, error)
	Update(
		ctx context.Context, id, name string, parentID *string,
	) (*domain.ResourceCategory, error)
	Delete(ctx context.Context, id string) error
}

type ResourceCategoryHandler struct {
	service ResourceCategoryService
}

func NewResourceCategoryHandler(
	service ResourceCategoryService,
) *ResourceCategoryHandler {
	return &ResourceCategoryHandler{service: service}
}

type resourceCategoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	ParentID *string `json:"parent_id"`
}

func (handler *ResourceCategoryHandler) Create(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request resourceCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	category, err := handler.service.Create(
		c.Request.Context(), request.Name, request.ParentID,
	)
	respond(c, category, err)
}

func (handler *ResourceCategoryHandler) List(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	items, err := handler.service.List(c.Request.Context())
	respond(c, items, err)
}

func (handler *ResourceCategoryHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request resourceCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	category, err := handler.service.Update(
		c.Request.Context(), c.Param("id"), request.Name, request.ParentID,
	)
	respond(c, category, err)
}

func (handler *ResourceCategoryHandler) Delete(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	if err := handler.service.Delete(
		c.Request.Context(), c.Param("id"),
	); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
