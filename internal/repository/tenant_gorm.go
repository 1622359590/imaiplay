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

func (repository *tenantGORMRepository) FindAll(
	ctx context.Context,
) ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	if err := repository.database.WithContext(ctx).Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}
