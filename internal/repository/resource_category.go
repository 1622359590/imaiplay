package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type ResourceCategoryRepository interface {
	Create(ctx context.Context, category *domain.ResourceCategory) error
	FindByID(ctx context.Context, id string) (*domain.ResourceCategory, error)
	FindByTenant(
		ctx context.Context, tenantID string,
	) ([]domain.ResourceCategory, error)
	Update(ctx context.Context, category *domain.ResourceCategory) error
	Delete(ctx context.Context, id string) error
}
