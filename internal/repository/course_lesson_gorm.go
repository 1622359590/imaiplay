package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type courseLessonGORMRepository struct{ database *gorm.DB }

func NewCourseLessonRepository(database *gorm.DB) CourseLessonRepository {
	return &courseLessonGORMRepository{database: database}
}

func (repo *courseLessonGORMRepository) Create(
	ctx context.Context, lesson *domain.CourseLesson,
) error {
	return repo.database.WithContext(ctx).Create(lesson).Error
}

func (repo *courseLessonGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.CourseLesson, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var lesson domain.CourseLesson
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&lesson).Error
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (repo *courseLessonGORMRepository) FindByChapter(
	ctx context.Context, chapterID string,
) ([]domain.CourseLesson, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var lessons []domain.CourseLesson
	err = repo.database.WithContext(ctx).
		Where("chapter_id = ? AND tenant_id = ?", chapterID, tenantID).
		Order("sort_order ASC, created_at ASC").Find(&lessons).Error
	return lessons, err
}

func (repo *courseLessonGORMRepository) Update(
	ctx context.Context, lesson *domain.CourseLesson,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repo.database.WithContext(ctx).Model(&domain.CourseLesson{}).
		Where("id = ? AND tenant_id = ?", lesson.ID, tenantID).
		Updates(map[string]interface{}{
			"title": lesson.Title, "content_type": lesson.ContentType,
			"content_url":      lesson.ContentURL,
			"duration_seconds": lesson.DurationSeconds,
			"sort_order":       lesson.SortOrder,
		})
	return affected(result)
}

func (repo *courseLessonGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.CourseLesson{})
	return affected(result)
}
