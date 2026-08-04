package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type CourseMaterialRepository interface {
	Create(ctx context.Context, material *domain.CourseMaterial) error
	FindByID(ctx context.Context, id string) (*domain.CourseMaterial, error)
	FindByCourse(ctx context.Context, courseID string) ([]domain.CourseMaterial, error)
	Update(ctx context.Context, material *domain.CourseMaterial) error
	Delete(ctx context.Context, id string) error
	DeleteByCourse(ctx context.Context, courseID string) error
}
