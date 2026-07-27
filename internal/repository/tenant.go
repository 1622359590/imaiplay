package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	FindByCode(ctx context.Context, code string) (*domain.Tenant, error)
	FindAll(ctx context.Context) ([]domain.Tenant, error)
}
