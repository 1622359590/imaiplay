package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type CourseRepository interface {
	Create(ctx context.Context, course *domain.Course) error
	FindByID(ctx context.Context, id string) (*domain.Course, error)
	FindByTenant(
		ctx context.Context,
		tenantID string,
		offset, limit int,
	) ([]domain.Course, int64, error)
	FindByTenantAndCreator(
		ctx context.Context,
		tenantID, creatorID string,
		offset, limit int,
	) ([]domain.Course, int64, error)
	FindPublishedByTenant(
		ctx context.Context,
		tenantID string,
		offset, limit int,
	) ([]domain.Course, int64, error)
	FindPublishedByID(
		ctx context.Context,
		tenantID, id string,
	) (*domain.Course, error)
	FindPublishedByLessonResource(
		ctx context.Context, tenantID, resourceID string,
	) ([]domain.Course, error)
	FindOfficial(ctx context.Context, offset, limit int) ([]domain.Course, int64, error)
	ActivateOfficial(ctx context.Context, tenantID, courseID string, enabled bool) error
	Update(ctx context.Context, course *domain.Course) error
	Delete(ctx context.Context, id string) error
}
