package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type DomainBindService interface {
	Verify(context.Context, string) (service.DomainBindStatus, error)
	Bind(context.Context, string) (service.DomainBindStatus, error)
	Status(context.Context) (service.DomainBindStatus, error)
	Unbind(context.Context) (service.DomainBindStatus, error)
}

type DomainBindHandler struct {
	service DomainBindService
}

func NewDomainBindHandler(service DomainBindService) *DomainBindHandler {
	return &DomainBindHandler{service: service}
}

func (handler *DomainBindHandler) Verify(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("请求参数无效"))
		return
	}
	status, err := handler.service.Verify(c.Request.Context(), request.Domain)
	respond(c, status, err)
}

func (handler *DomainBindHandler) Bind(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("请求参数无效"))
		return
	}
	status, err := handler.service.Bind(c.Request.Context(), request.Domain)
	respond(c, status, err)
}

func (handler *DomainBindHandler) Status(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	status, err := handler.service.Status(c.Request.Context())
	respond(c, status, err)
}

func (handler *DomainBindHandler) Unbind(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	status, err := handler.service.Unbind(c.Request.Context())
	respond(c, status, err)
}
