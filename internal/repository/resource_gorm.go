package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type resourceGORMRepository struct{ database *gorm.DB }

func NewResourceRepository(database *gorm.DB) ResourceRepository {
	return &resourceGORMRepository{database: database}
}

func (repo *resourceGORMRepository) Create(
	ctx context.Context, resource *domain.Resource,
) error {
	return repo.database.WithContext(ctx).Create(resource).Error
}

func (repo *resourceGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.Resource, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var resource domain.Resource
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&resource).Error
	return &resource, err
}

func (repo *resourceGORMRepository) FindByTenant(
	ctx context.Context, tenantID string, offset, limit int,
) ([]domain.Resource, int64, error) {
	query := repo.database.WithContext(ctx).Model(&domain.Resource{}).
		Where("tenant_id = ?", tenantID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.Resource
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (repo *resourceGORMRepository) Update(
	ctx context.Context, resource *domain.Resource,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).Model(&domain.Resource{}).
		Where("id = ? AND tenant_id = ?", resource.ID, tenantID).
		Updates(map[string]interface{}{
			"category_id": resource.CategoryID,
			"name":        resource.Name, "resource_type": resource.ResourceType,
			"url": resource.URL, "size_bytes": resource.SizeBytes,
		}))
}

func (repo *resourceGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.Resource{}))
}

func (repo *resourceGORMRepository) TotalSizeByTenant(ctx context.Context, tenantID string) (int64, error) {
	var total int64
	err := repo.database.WithContext(ctx).Model(&domain.Resource{}).Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(size_bytes), 0)").Scan(&total).Error
	return total, err
}
