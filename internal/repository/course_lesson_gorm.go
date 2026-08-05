package repository

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
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
		Where(
			"id = ? AND (tenant_id = ? OR (tenant_id = ? AND chapter_id IN (SELECT chapters.id FROM course_chapters AS chapters JOIN courses ON courses.id = chapters.course_id JOIN tenant_official_courses AS enabled_courses ON enabled_courses.course_id = courses.id AND enabled_courses.tenant_id = ? AND enabled_courses.enabled = ? WHERE chapters.tenant_id = ? AND courses.is_official = ? AND courses.status = ? AND courses.tenant_id = ?)))",
			id, tenantID, "", tenantID, true, "", true, 1, "",
		).
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
		Where("chapter_id = ? AND (tenant_id = ? OR (tenant_id = '' AND EXISTS (SELECT 1 FROM courses JOIN course_chapters ON course_chapters.course_id = courses.id WHERE course_chapters.id = chapter_id AND courses.is_official = true)))", chapterID, tenantID).
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
	result := repo.mutationScope(ctx, repo.database, tenantID).
		Where("id = ? AND tenant_id = ?", lesson.ID, tenantID).
		Updates(map[string]interface{}{
			"title": lesson.Title, "content_type": lesson.ContentType,
			"resource_id":      lesson.ResourceID,
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
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lesson domain.CourseLesson
		if err := repo.mutationScope(ctx, tx, tenantID).Where(
			"id = ? AND tenant_id = ?", id, tenantID,
		).First(&lesson).Error; err != nil {
			return err
		}
		if err := tx.Where(
			"lesson_id = ? AND tenant_id = ?", id, tenantID,
		).Delete(&domain.LessonProgress{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND tenant_id = ?", id, tenantID).
			Delete(&domain.CourseLesson{}).Error
	})
}

func (repo *courseLessonGORMRepository) mutationScope(
	ctx context.Context, database *gorm.DB, tenantID string,
) *gorm.DB {
	query := database.WithContext(ctx).Model(&domain.CourseLesson{})
	userID, _, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok {
		return query.Where("1 = 0")
	}
	switch role {
	case "superadmin":
		return query.Where("chapter_id IN (SELECT chapters.id FROM course_chapters AS chapters JOIN courses ON courses.id = chapters.course_id WHERE chapters.tenant_id = ? AND courses.tenant_id = ? AND courses.is_official = ?)", "", "", true)
	case "tenant_admin":
		return query.Where("tenant_id = ?", tenantID)
	case "instructor":
		return query.Where("tenant_id = ? AND chapter_id IN (SELECT chapters.id FROM course_chapters AS chapters JOIN courses ON courses.id = chapters.course_id WHERE chapters.tenant_id = ? AND courses.tenant_id = ? AND courses.is_official = ? AND courses.created_by = ?)", tenantID, tenantID, tenantID, false, userID)
	default:
		return query.Where("1 = 0")
	}
}
