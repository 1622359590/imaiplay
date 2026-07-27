package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type planGORMRepository struct{ database *gorm.DB }

func NewPlanRepository(database *gorm.DB) PlanRepository {
	return &planGORMRepository{database: database}
}

func (repo *planGORMRepository) Create(ctx context.Context, plan *domain.Plan) error {
	return repo.database.WithContext(ctx).Create(plan).Error
}
func (repo *planGORMRepository) FindByID(ctx context.Context, id string) (*domain.Plan, error) {
	var plan domain.Plan
	err := repo.database.WithContext(ctx).First(&plan, "id = ?", id).Error
	return &plan, err
}
func (repo *planGORMRepository) FindDefault(ctx context.Context) (*domain.Plan, error) {
	var plan domain.Plan
	err := repo.database.WithContext(ctx).Where("is_default = ? AND status = ?", true, 1).First(&plan).Error
	return &plan, err
}
func (repo *planGORMRepository) List(ctx context.Context, offset, limit int) ([]domain.Plan, int64, error) {
	query := repo.database.WithContext(ctx).Model(&domain.Plan{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var plans []domain.Plan
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&plans).Error
	return plans, total, err
}
func (repo *planGORMRepository) Update(ctx context.Context, plan *domain.Plan) error {
	result := repo.database.WithContext(ctx).Model(&domain.Plan{}).Where("id = ?", plan.ID).Updates(map[string]interface{}{
		"name": plan.Name, "storage_quota_bytes": plan.StorageQuotaBytes, "max_users": plan.MaxUsers,
		"max_courses": plan.MaxCourses, "features": plan.Features, "is_default": plan.IsDefault, "status": plan.Status,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (repo *planGORMRepository) Delete(ctx context.Context, id string) error {
	result := repo.database.WithContext(ctx).Where("id = ?", id).Delete(&domain.Plan{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
