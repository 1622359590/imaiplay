package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type LessonProgressRepository interface {
	Create(ctx context.Context, progress *domain.LessonProgress) error
	FindByID(ctx context.Context, id string) (*domain.LessonProgress, error)
	FindByUserAndLesson(
		ctx context.Context, userID, lessonID string,
	) (*domain.LessonProgress, error)
	FindByUser(
		ctx context.Context, userID string, offset, limit int,
	) ([]domain.LessonProgress, int64, error)
	Upsert(ctx context.Context, progress *domain.LessonProgress) error
}
