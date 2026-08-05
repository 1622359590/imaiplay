package repository

import (
	"context"
	"errors"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type tenantDemoRecordGORMRepository struct{ database *gorm.DB }

func NewTenantDemoRecordRepository(database *gorm.DB) TenantDemoRecordRepository {
	return &tenantDemoRecordGORMRepository{database: database}
}

func (repo *tenantDemoRecordGORMRepository) RegisterBatch(
	ctx context.Context, records []domain.TenantDemoRecord,
) error {
	if len(records) == 0 {
		return nil
	}
	tenantID, batchID := records[0].TenantID, records[0].BatchID
	if tenantID == "" || batchID == "" {
		return errors.New("demo record tenant and batch are required")
	}
	for _, record := range records {
		if record.TenantID != tenantID || record.BatchID != batchID ||
			record.RecordType == "" || record.RecordID == "" {
			return errors.New("demo record batch is inconsistent")
		}
	}
	return repo.database.WithContext(ctx).Create(&records).Error
}

func (repo *tenantDemoRecordGORMRepository) HasRecords(
	ctx context.Context, tenantID string,
) (bool, error) {
	var count int64
	err := repo.database.WithContext(ctx).Model(&domain.TenantDemoRecord{}).
		Where("tenant_id = ?", tenantID).Limit(1).Count(&count).Error
	return count > 0, err
}

func (repo *tenantDemoRecordGORMRepository) ListByTenant(
	ctx context.Context, tenantID string,
) ([]domain.TenantDemoRecord, error) {
	var records []domain.TenantDemoRecord
	err := repo.database.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at ASC").Order("id ASC").Find(&records).Error
	return records, err
}

func (repo *tenantDemoRecordGORMRepository) DeleteBatch(
	ctx context.Context, tenantID, batchID string,
) error {
	return repo.database.WithContext(ctx).
		Where("tenant_id = ? AND batch_id = ?", tenantID, batchID).
		Delete(&domain.TenantDemoRecord{}).Error
}
