package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type tenantGORMRepository struct {
	database *gorm.DB
}

func NewTenantRepository(database *gorm.DB) TenantRepository {
	return &tenantGORMRepository{database: database}
}

func (repository *tenantGORMRepository) Create(
	ctx context.Context,
	tenant *domain.Tenant,
) error {
	return repository.database.WithContext(ctx).Create(tenant).Error
}

func (repository *tenantGORMRepository) FindByCode(
	ctx context.Context,
	code string,
) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := repository.database.WithContext(ctx).
		Where("code = ?", code).
		First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (repository *tenantGORMRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := repository.database.WithContext(ctx).
		Where("id = ?", id).
		First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (repository *tenantGORMRepository) FindAll(
	ctx context.Context,
) ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	if err := repository.database.WithContext(ctx).Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (repository *tenantGORMRepository) Update(
	ctx context.Context,
	tenant *domain.Tenant,
) error {
	result := repository.database.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenant.ID).
		Updates(map[string]interface{}{
			"name": tenant.Name, "status": tenant.Status,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repository *tenantGORMRepository) UpdateTheme(ctx context.Context, tenant *domain.Tenant) error {
	result := repository.database.WithContext(ctx).Model(&domain.Tenant{}).
		Where("id = ?", tenant.ID).
		Updates(map[string]interface{}{
			"primary_color": tenant.PrimaryColor,
			"logo_url":      tenant.LogoURL,
			"welcome_text":  tenant.WelcomeText,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repository *tenantGORMRepository) UpdatePlan(ctx context.Context, tenant *domain.Tenant) error {
	result := repository.database.WithContext(ctx).Model(&domain.Tenant{}).Where("id = ?", tenant.ID).Update("plan_id", tenant.PlanID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repository *tenantGORMRepository) Delete(
	ctx context.Context,
	id string,
) error {
	result := repository.database.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.Tenant{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
