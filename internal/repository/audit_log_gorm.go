package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type auditLogGORMRepository struct{ database *gorm.DB }

func NewAuditLogRepository(database *gorm.DB) AuditLogRepository {
	return &auditLogGORMRepository{database: database}
}

func (repo *auditLogGORMRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return repo.database.WithContext(ctx).Create(log).Error
}

func (repo *auditLogGORMRepository) List(ctx context.Context, filter AuditLogFilter) ([]domain.AuditLog, int64, error) {
	query := repo.database.WithContext(ctx).Model(&domain.AuditLog{})
	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at < ?", *filter.To)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.AuditLog
	err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error
	return items, total, err
}
