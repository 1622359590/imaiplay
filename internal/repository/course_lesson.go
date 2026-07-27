package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type CourseLessonRepository interface {
	Create(ctx context.Context, lesson *domain.CourseLesson) error
	FindByID(ctx context.Context, id string) (*domain.CourseLesson, error)
	FindByChapter(ctx context.Context, chapterID string) ([]domain.CourseLesson, error)
	Update(ctx context.Context, lesson *domain.CourseLesson) error
	Delete(ctx context.Context, id string) error
}
