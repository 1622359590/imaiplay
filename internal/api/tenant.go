package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type TenantService interface {
	Create(ctx context.Context, code, name string) (*domain.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]domain.Tenant, int64, error)
	Get(ctx context.Context, id string) (*domain.Tenant, error)
	Update(ctx context.Context, id, name string, status int) (*domain.Tenant, error)
	Delete(ctx context.Context, id string) error
}

type TenantHandler struct {
	service TenantService
}

func NewTenantHandler(service TenantService) *TenantHandler {
	return &TenantHandler{service: service}
}

func (handler *TenantHandler) Create(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	tenant, err := handler.service.Create(c.Request.Context(), request.Code, request.Name)
	respond(c, tenant, err)
}

func (handler *TenantHandler) List(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.List(c.Request.Context(), offset, limit)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}

func (handler *TenantHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	tenant, err := handler.service.Get(c.Request.Context(), c.Param("id"))
	respond(c, tenant, err)
}

func (handler *TenantHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		Name   string `json:"name" binding:"required"`
		Status *int   `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	tenant, err := handler.service.Update(
		c.Request.Context(), c.Param("id"), request.Name, *request.Status,
	)
	respond(c, tenant, err)
}

func (handler *TenantHandler) Delete(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}

func respond(c *gin.Context, data interface{}, err error) {
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, data)
}
