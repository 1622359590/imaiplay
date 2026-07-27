package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type TenantRegistrationService interface {
	Register(ctx context.Context, organizationName, adminEmail, adminName, password string) (*service.TenantRegistrationResult, error)
	RegisterWithPhone(ctx context.Context, organizationName, adminEmail, phone, adminName, password string) (*service.TenantRegistrationResult, error)
	CreateForSuperadmin(ctx context.Context, organizationName, adminEmail, phone, adminName, password, planID string) (*service.TenantRegistrationResult, error)
	ClearDemoData(ctx context.Context) error
}

func (handler *TenantRegistrationHandler) CreateForSuperadmin(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		OrganizationName string `json:"organization_name" binding:"required"`
		AdminEmail       string `json:"admin_email" binding:"required,email"`
		AdminName        string `json:"admin_name" binding:"required"`
		Phone            string `json:"phone"`
		Password         string `json:"password" binding:"required,min=8"`
		PlanID           string `json:"plan_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	result, err := handler.service.CreateForSuperadmin(c.Request.Context(), request.OrganizationName, request.AdminEmail, request.Phone, request.AdminName, request.Password, request.PlanID)
	if result != nil {
		result.Token = ""
	}
	respond(c, result, err)
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
		Phone            string `json:"phone"`
		Password         string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	result, err := handler.service.RegisterWithPhone(c.Request.Context(), request.OrganizationName, request.AdminEmail, request.Phone, request.AdminName, request.Password)
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
