package repository

import (
	"context"
	"errors"

	"github.com/1622359590/imaiplay/internal/domain"
)

var (
	ErrCourseCategoryNameConflict = errors.New("course category name already exists")
	ErrCourseCategoryInUse        = errors.New("course category is referenced")
)

type CourseCategoryRepository interface {
	Create(ctx context.Context, category *domain.CourseCategory) error
	FindByID(ctx context.Context, tenantID, id string) (*domain.CourseCategory, error)
	FindByTenant(ctx context.Context, tenantID string) ([]domain.CourseCategory, error)
	Update(ctx context.Context, tenantID string, category *domain.CourseCategory) error
	Delete(ctx context.Context, tenantID, id string) error
}
