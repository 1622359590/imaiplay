package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type CourseChapterRepository interface {
	Create(ctx context.Context, chapter *domain.CourseChapter) error
	FindByID(ctx context.Context, id string) (*domain.CourseChapter, error)
	FindByCourse(ctx context.Context, courseID string) ([]domain.CourseChapter, error)
	Update(ctx context.Context, chapter *domain.CourseChapter) error
	Delete(ctx context.Context, id string) error
}
