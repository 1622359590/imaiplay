package api

import (
	"context"
	"errors"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type PortalResolver interface {
	Resolve(context.Context, string, string) (*service.Portal, error)
	ResolveByTenantID(context.Context, string) (*service.Portal, error)
}

type PortalHandler struct {
	service PortalResolver
}

func NewPortalHandler(service PortalResolver) *PortalHandler {
	return &PortalHandler{service: service}
}

func (handler *PortalHandler) Get(c *gin.Context) {
	if handler.service == nil {
		errorsx.GinResponse(c, errorsx.Internal("portal service is unavailable"))
		return
	}
	portal, err := handler.service.Resolve(
		c.Request.Context(),
		c.Query("tenant_code"),
		c.Request.Host,
	)
	if err != nil {
		portalErrorResponse(c, err)
		return
	}
	success(c, portal)
}

func (handler *PortalHandler) GetSession(c *gin.Context) {
	if handler.service == nil {
		errorsx.GinResponse(c, errorsx.Internal("portal service is unavailable"))
		return
	}
	_, tenantID, _, role, ok := usercontext.UserFromContext(
		c.Request.Context(),
	)
	if !ok || role != "learner" || strings.TrimSpace(tenantID) == "" {
		errorsx.GinResponse(c, errorsx.Forbidden("learner portal session required"))
		return
	}
	portal, err := handler.service.ResolveByTenantID(
		c.Request.Context(),
		tenantID,
	)
	if err != nil {
		portalErrorResponse(c, err)
		return
	}
	success(c, portal)
}

func portalErrorResponse(c *gin.Context, err error) {
	errorCode := ""
	var appErr *errorsx.AppError
	if errors.As(err, &appErr) {
		switch {
		case appErr.Code == 40400:
			errorCode = "portal_not_found"
		case appErr.Message == "tenant is suspended":
			errorCode = "portal_suspended"
		case appErr.Message == "tenant trial expired":
			errorCode = "portal_trial_expired"
		}
	}
	errorsx.GinResponseWithErrorCode(c, err, errorCode)
}
