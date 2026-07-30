package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type ResourceRepository interface {
	Create(ctx context.Context, resource *domain.Resource) error
	FindByID(ctx context.Context, id string) (*domain.Resource, error)
	FindByTenant(
		ctx context.Context, tenantID string, offset, limit int,
	) ([]domain.Resource, int64, error)
	FindAll(
		ctx context.Context, offset, limit int,
	) ([]domain.Resource, int64, error)
	Update(ctx context.Context, resource *domain.Resource) error
	Delete(ctx context.Context, id string) error
	TotalSizeByTenant(ctx context.Context, tenantID string) (int64, error)
}
