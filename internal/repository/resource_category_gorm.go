package repository

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
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
	tenantID, err := resourceCategoryMutationTenantID(ctx)
	if err != nil || (category.TenantID != "" && category.TenantID != tenantID) {
		return gorm.ErrRecordNotFound
	}
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := resourceCategoryParentInTenant(tx, tenantID, category.ParentID); err != nil {
			return err
		}
		category.TenantID = tenantID
		return tx.Create(category).Error
	})
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
	tenantID, err := resourceCategoryMutationTenantID(ctx)
	if err != nil || (category.TenantID != "" && category.TenantID != tenantID) {
		return gorm.ErrRecordNotFound
	}
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := resourceCategoryParentInTenant(tx, tenantID, category.ParentID); err != nil {
			return err
		}
		category.TenantID = tenantID
		return affected(tx.Model(&domain.ResourceCategory{}).
			Where("id = ? AND tenant_id = ?", category.ID, tenantID).
			Updates(map[string]interface{}{
				"name": category.Name, "parent_id": category.ParentID,
			}))
	})
}

func (repo *resourceCategoryGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := resourceCategoryMutationTenantID(ctx)
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

func resourceCategoryMutationTenantID(ctx context.Context) (string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return "", gorm.ErrRecordNotFound
	}
	return tenantID, nil
}

func resourceCategoryParentInTenant(tx *gorm.DB, tenantID string, parentID *string) error {
	if parentID == nil {
		return nil
	}
	var parent domain.ResourceCategory
	return tx.Select("id").Where(
		"id = ? AND tenant_id = ?", *parentID, tenantID,
	).First(&parent).Error
}
