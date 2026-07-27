package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type resourceCategoryGORMRepository struct{ database *gorm.DB }

func NewResourceCategoryRepository(database *gorm.DB) ResourceCategoryRepository {
	return &resourceCategoryGORMRepository{database: database}
}

func (repo *resourceCategoryGORMRepository) Create(
	ctx context.Context, category *domain.ResourceCategory,
) error {
	return repo.database.WithContext(ctx).Create(category).Error
}

func (repo *resourceCategoryGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.ResourceCategory, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var category domain.ResourceCategory
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&category).Error
	return &category, err
}

func (repo *resourceCategoryGORMRepository) FindByTenant(
	ctx context.Context, tenantID string,
) ([]domain.ResourceCategory, error) {
	var categories []domain.ResourceCategory
	err := repo.database.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at ASC").Find(&categories).Error
	return categories, err
}

func (repo *resourceCategoryGORMRepository) Update(
	ctx context.Context, category *domain.ResourceCategory,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).Model(&domain.ResourceCategory{}).
		Where("id = ? AND tenant_id = ?", category.ID, tenantID).
		Updates(map[string]interface{}{
			"name": category.Name, "parent_id": category.ParentID,
		}))
}

func (repo *resourceCategoryGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var category domain.ResourceCategory
		if err := tx.Where(
			"id = ? AND tenant_id = ?", id, tenantID,
		).First(&category).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Resource{}).Where(
			"category_id = ? AND tenant_id = ?", id, tenantID,
		).Update("category_id", nil).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND tenant_id = ?", id, tenantID).
			Delete(&domain.ResourceCategory{}).Error
	})
}
