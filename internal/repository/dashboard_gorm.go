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

func (repo *dashboardGORMRepository) TenantStats(
	ctx context.Context,
	tenantID, todayDate, yesterdayDate string,
	todayStart, todayEnd time.Time,
) (TenantDashboardMetrics, error) {
	metrics := TenantDashboardMetrics{
		TodayLearningRanking: make([]LearningRankItem, 0),
	}
	err := repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		counts := []struct {
			query *gorm.DB
			value *int64
		}{
			{
				tx.Model(&domain.User{}).Where(
					"tenant_id = ? AND role = ? AND status = ?", tenantID, "learner", 1,
				),
				&metrics.LearnerCount,
			},
			{
				tx.Model(&domain.User{}).Where(
					"tenant_id = ? AND role = ? AND status = ? AND created_at >= ? AND created_at < ?",
					tenantID, "learner", 1, todayStart, todayEnd,
				),
				&metrics.TodayNewLearnerCount,
			},
			{
				tenantVisibleCourses(tx, tenantID),
				&metrics.CourseCount,
			},
			{
				tenantVisibleCourses(tx, tenantID).Where("courses.status = ?", 1),
				&metrics.PublishedCourseCount,
			},
			{
				tx.Model(&domain.ResourceCategory{}).Where("tenant_id = ?", tenantID),
				&metrics.ResourceCategoryCount,
			},
			{
				tx.Model(&domain.Resource{}).Where("tenant_id = ?", tenantID),
				&metrics.ResourceCount,
			},
			{
				tx.Model(&domain.User{}).Where(
					"tenant_id = ? AND role IN ? AND status = ?",
					tenantID, []string{"tenant_admin", "instructor"}, 1,
				),
				&metrics.ManagerCount,
			},
		}
		for _, count := range counts {
			if err := count.query.Count(count.value).Error; err != nil {
				return err
			}
		}

		var err error
		metrics.TodayLearningUserCount, err = activeLearningUserCount(
			tx, tenantID, todayDate,
		)
		if err != nil {
			return err
		}
		metrics.YesterdayLearningUserCount, err = activeLearningUserCount(
			tx, tenantID, yesterdayDate,
		)
		if err != nil {
			return err
		}
		metrics.TodayLearningUserDelta = metrics.TodayLearningUserCount -
			metrics.YesterdayLearningUserCount

		var demoCount int64
		if err := tx.Model(&domain.TenantDemoRecord{}).
			Where("tenant_id = ?", tenantID).Limit(1).Count(&demoCount).Error; err != nil {
			return err
		}
		metrics.HasDemoData = demoCount > 0

		type resourceTypeCount struct {
			ResourceType string
			Count        int64
		}
		var resourceCounts []resourceTypeCount
		if err := tx.Model(&domain.Resource{}).
			Select("resource_type, COUNT(*) AS count").
			Where("tenant_id = ?", tenantID).
			Group("resource_type").Scan(&resourceCounts).Error; err != nil {
			return err
		}
		for _, count := range resourceCounts {
			switch count.ResourceType {
			case "video":
				metrics.ResourceTypeCounts.Video = count.Count
			case "image":
				metrics.ResourceTypeCounts.Image = count.Count
			case "document":
				metrics.ResourceTypeCounts.Document = count.Count
			case "attachment":
				metrics.ResourceTypeCounts.Attachment = count.Count
			}
		}

		return tx.Table("learning_daily_stats AS stats").
			Select(`
				stats.user_id AS user_id,
				users.name AS display_name,
				stats.duration_seconds AS duration_seconds
			`).
			Joins("JOIN users ON users.id = stats.user_id AND users.tenant_id = stats.tenant_id").
			Where(
				"stats.tenant_id = ? AND stats.study_date = ? AND stats.duration_seconds > ? AND users.role = ? AND users.status = ?",
				tenantID, todayDate, 0, "learner", 1,
			).
			Order("stats.duration_seconds DESC").
			Order("users.name ASC").
			Order("stats.user_id ASC").
			Limit(10).
			Scan(&metrics.TodayLearningRanking).Error
	})
	return metrics, err
}

func activeLearningUserCount(
	tx *gorm.DB, tenantID, studyDate string,
) (int64, error) {
	var count int64
	err := tx.Table("learning_daily_stats AS stats").
		Joins("JOIN users ON users.id = stats.user_id AND users.tenant_id = stats.tenant_id").
		Where(
			"stats.tenant_id = ? AND stats.study_date = ? AND stats.duration_seconds > ? AND users.role = ? AND users.status = ?",
			tenantID, studyDate, 0, "learner", 1,
		).
		Distinct("stats.user_id").Count(&count).Error
	return count, err
}

func tenantVisibleCourses(tx *gorm.DB, tenantID string) *gorm.DB {
	return tx.Model(&domain.Course{}).
		Where(`
			(courses.tenant_id = ? AND courses.is_official = ?)
			OR
			(courses.tenant_id = ? AND courses.is_official = ? AND EXISTS (
				SELECT 1 FROM tenant_official_courses AS enabled_courses
				WHERE enabled_courses.tenant_id = ?
				  AND enabled_courses.course_id = courses.id
				  AND enabled_courses.enabled = ?
			))
		`, tenantID, false, "", true, tenantID, true)
}

func (repo *dashboardGORMRepository) InstructorStats(
	ctx context.Context,
	tenantID, userID, todayDate string,
	todayStart, todayEnd time.Time,
) (InstructorDashboardMetrics, error) {
	metrics := InstructorDashboardMetrics{
		RecentCourses: make([]InstructorCourse, 0),
	}
	database := repo.database.WithContext(ctx)
	owned := database.Model(&domain.Course{}).Where(
		"tenant_id = ? AND is_official = ? AND created_by = ?",
		tenantID, false, userID,
	)
	if err := owned.Count(&metrics.CourseCount).Error; err != nil {
		return metrics, err
	}
	if err := owned.Where("status = ?", 1).
		Count(&metrics.PublishedCourseCount).Error; err != nil {
		return metrics, err
	}
	if err := database.Model(&domain.Course{}).
		Select("id, title, status, updated_at").
		Where(
			"tenant_id = ? AND is_official = ? AND created_by = ?",
			tenantID, false, userID,
		).
		Order("updated_at DESC").Order("id ASC").Limit(5).
		Scan(&metrics.RecentCourses).Error; err != nil {
		return metrics, err
	}
	err := database.Table("learning_daily_stats AS stats").
		Joins("JOIN users ON users.id = stats.user_id AND users.tenant_id = stats.tenant_id").
		Joins("JOIN learning_time_reports AS reports ON reports.user_id = stats.user_id AND reports.tenant_id = stats.tenant_id").
		Joins("JOIN course_lessons AS lessons ON lessons.id = reports.lesson_id").
		Joins("JOIN course_chapters AS chapters ON chapters.id = lessons.chapter_id").
		Joins("JOIN courses ON courses.id = chapters.course_id").
		Joins(`JOIN course_enrollments AS enrollments
			ON enrollments.course_id = courses.id
			AND enrollments.user_id = stats.user_id
			AND enrollments.tenant_id = stats.tenant_id
			AND enrollments.status = 1`).
		Where(
			`stats.tenant_id = ? AND stats.study_date = ? AND stats.duration_seconds > 0
			AND users.role = ? AND users.status = ?
			AND reports.created_at >= ? AND reports.created_at < ?
			AND courses.tenant_id = ? AND courses.is_official = ? AND courses.created_by = ?
			AND chapters.tenant_id = ? AND lessons.tenant_id = ?`,
			tenantID, todayDate, "learner", 1, todayStart, todayEnd,
			tenantID, false, userID, tenantID, tenantID,
		).
		Distinct("stats.user_id").Count(&metrics.TodayLearningUserCount).Error
	return metrics, err
}

func (repo *dashboardGORMRepository) PlatformStats(
	ctx context.Context,
) (PlatformDashboardMetrics, error) {
	metrics := PlatformDashboardMetrics{
		RecentTenants: make([]PlatformTenant, 0),
	}
	database := repo.database.WithContext(ctx)
	counts := []struct {
		query *gorm.DB
		value *int64
	}{
		{database.Model(&domain.Tenant{}), &metrics.TenantCount},
		{database.Model(&domain.Tenant{}).Where("status = ?", 1), &metrics.ActiveTenantCount},
		{database.Model(&domain.User{}).Where("role = ? AND status = ?", "learner", 1), &metrics.LearnerCount},
		{database.Model(&domain.Course{}), &metrics.CourseCount},
	}
	for _, count := range counts {
		if err := count.query.Count(count.value).Error; err != nil {
			return metrics, err
		}
	}
	if err := database.Model(&domain.Tenant{}).
		Select("id, name, code, status, lifecycle_status, created_at").
		Order("created_at DESC").Order("id ASC").Limit(5).
		Scan(&metrics.RecentTenants).Error; err != nil {
		return metrics, err
	}
	return metrics, nil
}
