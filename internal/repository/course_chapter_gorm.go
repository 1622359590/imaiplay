package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type courseChapterGORMRepository struct{ database *gorm.DB }

func NewCourseChapterRepository(database *gorm.DB) CourseChapterRepository {
	return &courseChapterGORMRepository{database: database}
}

func (repo *courseChapterGORMRepository) Create(
	ctx context.Context, chapter *domain.CourseChapter,
) error {
	return repo.database.WithContext(ctx).Create(chapter).Error
}

func (repo *courseChapterGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.CourseChapter, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var chapter domain.CourseChapter
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&chapter).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

func (repo *courseChapterGORMRepository) FindByCourse(
	ctx context.Context, courseID string,
) ([]domain.CourseChapter, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var chapters []domain.CourseChapter
	err = repo.database.WithContext(ctx).
		Where("course_id = ? AND tenant_id = ?", courseID, tenantID).
		Order("sort_order ASC, created_at ASC").Find(&chapters).Error
	return chapters, err
}

func (repo *courseChapterGORMRepository) Update(
	ctx context.Context, chapter *domain.CourseChapter,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repo.database.WithContext(ctx).Model(&domain.CourseChapter{}).
		Where("id = ? AND tenant_id = ?", chapter.ID, tenantID).
		Updates(map[string]interface{}{
			"title": chapter.Title, "sort_order": chapter.SortOrder,
		})
	return affected(result)
}

func (repo *courseChapterGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var chapter domain.CourseChapter
		if err := tx.Where(
			"id = ? AND tenant_id = ?", id, tenantID,
		).First(&chapter).Error; err != nil {
			return err
		}
		lessonIDs := tx.Model(&domain.CourseLesson{}).Select("id").
			Where("chapter_id = ? AND tenant_id = ?", id, tenantID)
		if err := tx.Where(
			"tenant_id = ? AND lesson_id IN (?)", tenantID, lessonIDs,
		).Delete(&domain.LessonProgress{}).Error; err != nil {
			return err
		}
		if err := tx.Where(
			"chapter_id = ? AND tenant_id = ?", id, tenantID,
		).Delete(&domain.CourseLesson{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND tenant_id = ?", id, tenantID).
			Delete(&domain.CourseChapter{}).Error
	})
}
