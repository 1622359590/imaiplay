package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
)

type AuditLogFilter struct {
	TenantID string
	Action   string
	UserID   string
	From     *time.Time
	To       *time.Time
	Offset   int
	Limit    int
}

type AuditLogRepository interface {
	Create(context.Context, *domain.AuditLog) error
	List(context.Context, AuditLogFilter) ([]domain.AuditLog, int64, error)
}
