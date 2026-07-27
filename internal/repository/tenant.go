package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	FindByID(ctx context.Context, id string) (*domain.Tenant, error)
	FindByCode(ctx context.Context, code string) (*domain.Tenant, error)
	FindByCustomDomain(ctx context.Context, customDomain string) (*domain.Tenant, error)
	FindAll(ctx context.Context) ([]domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	UpdateTheme(ctx context.Context, tenant *domain.Tenant) error
	UpdatePlan(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id string) error
}
