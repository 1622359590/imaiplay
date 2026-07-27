package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type dashboardGORMRepository struct {
	database *gorm.DB
}

func NewDashboardRepository(database *gorm.DB) DashboardRepository {
	return &dashboardGORMRepository{database: database}
}

func (repo *dashboardGORMRepository) Get(
	ctx context.Context,
	tenantID string,
	dayStart, dayEnd time.Time,
) (DashboardMetrics, error) {
	var metrics DashboardMetrics
	err := repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		queries := []struct {
			query *gorm.DB
			value *int64
		}{
			{
				tx.Model(&domain.User{}).
					Where("tenant_id = ? AND status = ?", tenantID, 1),
				&metrics.UserCount,
			},
			{
				tx.Model(&domain.Course{}).Where("tenant_id = ?", tenantID),
				&metrics.CourseCount,
			},
			{
				tx.Model(&domain.Course{}).
					Where("tenant_id = ? AND status = ?", tenantID, 1),
				&metrics.PublishedCourseCount,
			},
			{
				tx.Model(&domain.User{}).Where(
					"tenant_id = ? AND created_at >= ? AND created_at < ?",
					tenantID, dayStart, dayEnd,
				),
				&metrics.TodayNewUserCount,
			},
		}
		for _, query := range queries {
			if err := query.query.Count(query.value).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&domain.LessonProgress{}).
			Where(
				"tenant_id = ? AND updated_at >= ? AND updated_at < ?",
				tenantID, dayStart, dayEnd,
			).
			Distinct("user_id").
			Count(&metrics.TodayLearningUserCount).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.LessonProgress{}).
			Select("COALESCE(SUM(last_position_seconds), 0)").
			Where("tenant_id = ?", tenantID).
			Scan(&metrics.TotalLearningSeconds).Error; err != nil {
			return err
		}
		if err := enrollmentCounts(tx, tenantID, &metrics); err != nil {
			return err
		}
		return nil
	})
	return metrics, err
}

func enrollmentCounts(
	tx *gorm.DB, tenantID string, metrics *DashboardMetrics,
) error {
	active := tx.Table("course_enrollments AS ce").
		Where("ce.tenant_id = ? AND ce.status = ?", tenantID, 1)
	if err := active.Distinct("ce.user_id").
		Count(&metrics.ActiveEnrollmentCount).Error; err != nil {
		return err
	}
	return active.Where(`
		NOT EXISTS (
			SELECT 1
			FROM course_lessons cl
			JOIN course_chapters cc ON cc.id = cl.chapter_id
			WHERE cc.course_id = ce.course_id
			  AND cc.tenant_id = ce.tenant_id
			  AND cl.tenant_id = ce.tenant_id
			  AND NOT EXISTS (
				SELECT 1 FROM lesson_progress lp
				WHERE lp.user_id = ce.user_id
				  AND lp.lesson_id = cl.id
				  AND lp.tenant_id = ce.tenant_id
				  AND lp.status = 2
			  )
		)
	`).Distinct("ce.user_id").
		Count(&metrics.CompletedEnrollmentCount).Error
}
