package api

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
)

type AuditService interface {
	Record(context.Context, domain.AuditEvent) error
	List(context.Context, repository.AuditLogFilter) ([]domain.AuditLog, int64, error)
}

type AuditHandler struct{ service AuditService }

func NewAuditHandler(service AuditService) *AuditHandler { return &AuditHandler{service: service} }

func (handler *AuditHandler) ListTenant(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	handler.list(c)
}

func (handler *AuditHandler) ListAdmin(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.list(c)
}

func (handler *AuditHandler) list(c *gin.Context) {
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	filter := repository.AuditLogFilter{Action: c.Query("action"), UserID: c.Query("user_id"), TenantID: c.Query("tenant_id"), Offset: offset, Limit: limit}
	if value := c.Query("from"); value != "" {
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			errorsx.GinResponse(c, errorsx.BadRequest("invalid from"))
			return
		}
		filter.From = &parsed
	}
	if value := c.Query("to"); value != "" {
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			errorsx.GinResponse(c, errorsx.BadRequest("invalid to"))
			return
		}
		filter.To = &parsed
	}
	items, total, err := handler.service.List(c.Request.Context(), filter)
	if err != nil {
		errorsx.GinResponse(c, errorsx.Forbidden("permission denied"))
		return
	}
	success(c, gin.H{"items": items, "total": total})
}
