package repository

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type courseGORMRepository struct{ database *gorm.DB }

func NewCourseRepository(database *gorm.DB) CourseRepository {
	return &courseGORMRepository{database: database}
}

func (repo *courseGORMRepository) Create(
	ctx context.Context, course *domain.Course,
) error {
	return repo.database.WithContext(ctx).Create(course).Error
}

func (repo *courseGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.Course, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var course domain.Course
	err = repo.courseScope(ctx, tenantID).
		Where("id = ?", id).First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (repo *courseGORMRepository) FindByTenant(
	ctx context.Context, tenantID string, offset, limit int,
) ([]domain.Course, int64, error) {
	return repo.find(ctx, repo.courseScope(ctx, tenantID), offset, limit)
}

func (repo *courseGORMRepository) FindPublishedByTenant(
	ctx context.Context, tenantID string, offset, limit int,
) ([]domain.Course, int64, error) {
	query := repo.database.WithContext(ctx).Model(&domain.Course{}).
		Where("(tenant_id = ? AND status = ?) OR (is_official = ? AND status = ? AND id IN (SELECT course_id FROM tenant_official_courses WHERE tenant_id = ? AND enabled = ?))", tenantID, 1, true, 1, tenantID, true)
	return repo.find(ctx, query, offset, limit)
}

func (repo *courseGORMRepository) FindPublishedByID(
	ctx context.Context, tenantID, id string,
) (*domain.Course, error) {
	var course domain.Course
	err := repo.database.WithContext(ctx).
		Where("id = ? AND ((tenant_id = ? AND status = ?) OR (is_official = ? AND status = ? AND id IN (SELECT course_id FROM tenant_official_courses WHERE tenant_id = ? AND enabled = ?)))", id, tenantID, 1, true, 1, tenantID, true).
		First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (repo *courseGORMRepository) FindOfficial(ctx context.Context, offset, limit int) ([]domain.Course, int64, error) {
	return repo.find(ctx, repo.database.WithContext(ctx).Model(&domain.Course{}).Where("is_official = ? AND tenant_id = ?", true, ""), offset, limit)
}

func (repo *courseGORMRepository) ActivateOfficial(ctx context.Context, tenantID, courseID string, enabled bool) error {
	var course domain.Course
	if err := repo.database.WithContext(ctx).Where("id = ? AND is_official = ? AND tenant_id = ?", courseID, true, "").First(&course).Error; err != nil {
		return err
	}
	return repo.database.WithContext(ctx).Exec("INSERT INTO tenant_official_courses (tenant_id, course_id, enabled) VALUES (?, ?, ?) ON CONFLICT (tenant_id, course_id) DO UPDATE SET enabled = excluded.enabled", tenantID, courseID, enabled).Error
}

func (repo *courseGORMRepository) find(
	ctx context.Context, query *gorm.DB, offset, limit int,
) ([]domain.Course, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var courses []domain.Course
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).
		Find(&courses).Error
	return courses, total, err
}

func (repo *courseGORMRepository) Update(
	ctx context.Context, course *domain.Course,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repo.courseScope(ctx, tenantID).
		Where("id = ?", course.ID).
		Updates(map[string]interface{}{
			"title": course.Title, "description": course.Description,
			"cover_image": course.CoverImage, "status": course.Status,
		})
	return affected(result)
}

func (repo *courseGORMRepository) Delete(ctx context.Context, id string) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scoped := (&courseGORMRepository{database: tx}).courseScope(ctx, tenantID)
		var course domain.Course
		if err := scoped.Where("id = ?", id).First(&course).Error; err != nil {
			return err
		}
		if course.IsOfficial {
			chapterIDs := tx.Model(&domain.CourseChapter{}).Select("id").
				Where("course_id = ? AND tenant_id = ?", id, "")
			lessonIDs := tx.Model(&domain.CourseLesson{}).Select("id").
				Where("tenant_id = ? AND chapter_id IN (?)", "", chapterIDs)
			if err := tx.Where(
				"lesson_id IN (?)", lessonIDs,
			).Delete(&domain.LessonProgress{}).Error; err != nil {
				return err
			}
			if err := tx.Where(
				"course_id = ?", id,
			).Delete(&domain.CourseEnrollment{}).Error; err != nil {
				return err
			}
			if err := tx.Where(
				"course_id = ?", id,
			).Delete(&domain.TenantOfficialCourse{}).Error; err != nil {
				return err
			}
			if err := tx.Where(
				"tenant_id = ? AND chapter_id IN (?)", "", chapterIDs,
			).Delete(&domain.CourseLesson{}).Error; err != nil {
				return err
			}
			if err := tx.Where(
				"course_id = ? AND tenant_id = ?", id, "",
			).Delete(&domain.CourseChapter{}).Error; err != nil {
				return err
			}
			return tx.Where("id = ? AND tenant_id = ?", id, "").
				Delete(&domain.Course{}).Error
		}
		chapterIDs := tx.Model(&domain.CourseChapter{}).Select("id").
			Where("course_id = ? AND tenant_id = ?", id, tenantID)
		lessonIDs := tx.Model(&domain.CourseLesson{}).Select("id").
			Where("tenant_id = ? AND chapter_id IN (?)", tenantID, chapterIDs)
		if err := tx.Where(
			"tenant_id = ? AND lesson_id IN (?)", tenantID, lessonIDs,
		).Delete(&domain.LessonProgress{}).Error; err != nil {
			return err
		}
		if err := tx.Where(
			"tenant_id = ? AND chapter_id IN (?)", tenantID, chapterIDs,
		).Delete(&domain.CourseLesson{}).Error; err != nil {
			return err
		}
		if err := tx.Where(
			"course_id = ? AND tenant_id = ?", id, tenantID,
		).Delete(&domain.CourseChapter{}).Error; err != nil {
			return err
		}
		if err := tx.Where(
			"course_id = ? AND tenant_id = ?", id, tenantID,
		).Delete(&domain.CourseEnrollment{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND tenant_id = ?", id, tenantID).
			Delete(&domain.Course{}).Error
	})
}

func (repo *courseGORMRepository) courseScope(
	ctx context.Context, tenantID string,
) *gorm.DB {
	query := repo.database.WithContext(ctx).Model(&domain.Course{}).
		Where("tenant_id = ?", tenantID)
	userID, _, _, role, ok := usercontext.UserFromContext(ctx)
	if ok && role == "instructor" {
		query = query.Where("created_by = ?", userID)
	}
	return query
}
