package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type TenantRegistrationService interface {
	Register(ctx context.Context, organizationName, adminEmail, adminName, password string) (*service.TenantRegistrationResult, error)
	ClearDemoData(ctx context.Context) error
}

type TenantRegistrationHandler struct{ service TenantRegistrationService }

func NewTenantRegistrationHandler(service TenantRegistrationService) *TenantRegistrationHandler {
	return &TenantRegistrationHandler{service: service}
}

func (handler *TenantRegistrationHandler) Register(c *gin.Context) {
	var request struct {
		OrganizationName string `json:"organization_name" binding:"required"`
		AdminEmail       string `json:"admin_email" binding:"required,email"`
		AdminName        string `json:"admin_name" binding:"required"`
		Password         string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	result, err := handler.service.Register(c.Request.Context(), request.OrganizationName, request.AdminEmail, request.AdminName, request.Password)
	respond(c, result, err)
}

func (handler *TenantRegistrationHandler) ClearDemoData(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	if err := handler.service.ClearDemoData(c.Request.Context()); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
