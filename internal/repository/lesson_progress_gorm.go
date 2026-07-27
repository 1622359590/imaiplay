package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type lessonProgressGORMRepository struct{ database *gorm.DB }

func NewLessonProgressRepository(database *gorm.DB) LessonProgressRepository {
	return &lessonProgressGORMRepository{database: database}
}

func (repo *lessonProgressGORMRepository) Create(
	ctx context.Context, progress *domain.LessonProgress,
) error {
	return repo.database.WithContext(ctx).Create(progress).Error
}

func (repo *lessonProgressGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.LessonProgress, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var progress domain.LessonProgress
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&progress).Error
	return &progress, err
}

func (repo *lessonProgressGORMRepository) FindByUserAndLesson(
	ctx context.Context, userID, lessonID string,
) (*domain.LessonProgress, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var progress domain.LessonProgress
	err = repo.database.WithContext(ctx).
		Where(
			"user_id = ? AND lesson_id = ? AND tenant_id = ?",
			userID, lessonID, tenantID,
		).First(&progress).Error
	return &progress, err
}

func (repo *lessonProgressGORMRepository) FindByUser(
	ctx context.Context, userID string, offset, limit int,
) ([]domain.LessonProgress, int64, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	query := repo.database.WithContext(ctx).Model(&domain.LessonProgress{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.LessonProgress
	err = query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (repo *lessonProgressGORMRepository) Upsert(
	ctx context.Context, progress *domain.LessonProgress,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	progress.TenantID = tenantID
	err = repo.database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"}, {Name: "user_id"}, {Name: "lesson_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"progress_percent":      progress.ProgressPercent,
			"status":                progress.Status,
			"last_position_seconds": progress.LastPositionSeconds,
			"completed_at":          progress.CompletedAt,
			"updated_at":            gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(progress).Error
	if err != nil {
		return err
	}
	found, err := repo.FindByUserAndLesson(ctx, progress.UserID, progress.LessonID)
	if err != nil {
		return err
	}
	*progress = *found
	return nil
}
