package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type courseCategoryGORMRepository struct{ database *gorm.DB }

func NewCourseCategoryRepository(database *gorm.DB) CourseCategoryRepository {
	return &courseCategoryGORMRepository{database: database}
}

func (repo *courseCategoryGORMRepository) Create(
	ctx context.Context, category *domain.CourseCategory,
) error {
	requestedStatus := category.Status
	err := repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(category).Error; err != nil {
			return err
		}
		if requestedStatus == 0 {
			if err := tx.Model(&domain.CourseCategory{}).
				Where("id = ? AND tenant_id = ?", category.ID, category.TenantID).
				UpdateColumn("status", 0).Error; err != nil {
				return err
			}
			category.Status = 0
		}
		return nil
	})
	return mapCourseCategoryConstraintError(err)
}

func (repo *courseCategoryGORMRepository) FindByID(
	ctx context.Context, tenantID, id string,
) (*domain.CourseCategory, error) {
	var category domain.CourseCategory
	err := repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&category).Error
	return &category, err
}

func (repo *courseCategoryGORMRepository) FindByTenant(
	ctx context.Context, tenantID string,
) ([]domain.CourseCategory, error) {
	var categories []domain.CourseCategory
	err := repo.database.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("sort_order ASC").Order("created_at ASC").Order("id ASC").
		Find(&categories).Error
	return categories, err
}

func (repo *courseCategoryGORMRepository) Update(
	ctx context.Context, tenantID string, category *domain.CourseCategory,
) error {
	result := repo.database.WithContext(ctx).Model(&domain.CourseCategory{}).
		Where("id = ? AND tenant_id = ?", category.ID, tenantID).
		Updates(map[string]interface{}{
			"name": category.Name, "normalized_name": category.NormalizedName,
			"sort_order": category.SortOrder, "status": category.Status,
		})
	if err := mapCourseCategoryConstraintError(result.Error); err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repo *courseCategoryGORMRepository) Delete(
	ctx context.Context, tenantID, id string,
) error {
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var category domain.CourseCategory
		if err := tx.Where("id = ? AND tenant_id = ?", id, tenantID).
			First(&category).Error; err != nil {
			return err
		}
		var references int64
		if err := tx.Model(&domain.Course{}).
			Where("tenant_id = ? AND category_id = ?", tenantID, id).
			Count(&references).Error; err != nil {
			return err
		}
		if references > 0 {
			return ErrCourseCategoryInUse
		}
		return affected(tx.Where("id = ? AND tenant_id = ?", id, tenantID).
			Delete(&domain.CourseCategory{}))
	})
}

func mapCourseCategoryConstraintError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "duplicate key") {
		return ErrCourseCategoryNameConflict
	}
	return err
}
