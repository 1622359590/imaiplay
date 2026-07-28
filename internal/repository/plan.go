package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type PlanRepository interface {
	Create(context.Context, *domain.Plan) error
	FindByID(context.Context, string) (*domain.Plan, error)
	FindDefault(context.Context) (*domain.Plan, error)
	List(context.Context, int, int) ([]domain.Plan, int64, error)
	Update(context.Context, *domain.Plan) error
	Delete(context.Context, string) error
}
