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
				tenantScoped(tx.Model(&domain.User{}), "tenant_id", tenantID).
					Where("status = ?", 1),
				&metrics.UserCount,
			},
			{
				tenantScoped(tx.Model(&domain.Course{}), "tenant_id", tenantID),
				&metrics.CourseCount,
			},
			{
				tenantScoped(tx.Model(&domain.Course{}), "tenant_id", tenantID).
					Where("status = ?", 1),
				&metrics.PublishedCourseCount,
			},
			{
				tenantScoped(tx.Model(&domain.User{}), "tenant_id", tenantID).
					Where("created_at >= ? AND created_at < ?", dayStart, dayEnd),
				&metrics.TodayNewUserCount,
			},
		}
		for _, query := range queries {
			if err := query.query.Count(query.value).Error; err != nil {
				return err
			}
		}
		if err := tenantScoped(tx.Model(&domain.LessonProgress{}), "tenant_id", tenantID).
			Where("updated_at >= ? AND updated_at < ?", dayStart, dayEnd).
			Distinct("user_id").
			Count(&metrics.TodayLearningUserCount).Error; err != nil {
			return err
		}
		if err := tenantScoped(tx.Model(&domain.LessonProgress{}), "tenant_id", tenantID).
			Select("COALESCE(SUM(last_position_seconds), 0)").
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

func (repo *dashboardGORMRepository) PlatformStats(
	ctx context.Context,
) (PlatformDashboardMetrics, error) {
	var metrics PlatformDashboardMetrics
	database := repo.database.WithContext(ctx)
	if err := database.Model(&domain.Tenant{}).Count(&metrics.TenantCount).Error; err != nil {
		return metrics, err
	}
	if err := database.Model(&domain.Tenant{}).Where("status = ?", 1).Count(&metrics.ActiveTenantCount).Error; err != nil {
		return metrics, err
	}
	if err := database.Model(&domain.User{}).
		Where("role = ? AND status = ?", "learner", 1).
		Count(&metrics.LearnerCount).Error; err != nil {
		return metrics, err
	}
	if err := database.Model(&domain.Course{}).Count(&metrics.CourseCount).Error; err != nil {
		return metrics, err
	}
	if err := database.Model(&domain.Tenant{}).
		Order("created_at DESC").Limit(5).Find(&metrics.RecentTenants).Error; err != nil {
		return metrics, err
	}
	return metrics, nil
}

func enrollmentCounts(
	tx *gorm.DB, tenantID string, metrics *DashboardMetrics,
) error {
	active := tx.Table("course_enrollments AS ce").Where("ce.status = ?", 1)
	if tenantID != "" {
		active = active.Where("ce.tenant_id = ?", tenantID)
	}
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

func tenantScoped(query *gorm.DB, column, tenantID string) *gorm.DB {
	if tenantID == "" {
		return query
	}
	return query.Where(column+" = ?", tenantID)
}
