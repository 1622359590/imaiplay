package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
)

type DashboardMetrics struct {
	UserCount                int64
	CourseCount              int64
	PublishedCourseCount     int64
	TodayNewUserCount        int64
	TodayLearningUserCount   int64
	TotalLearningSeconds     int64
	ActiveEnrollmentCount    int64
	CompletedEnrollmentCount int64
}

type PlatformDashboardMetrics struct {
	TenantCount       int64
	ActiveTenantCount int64
	LearnerCount      int64
	CourseCount       int64
	RecentTenants     []domain.Tenant
}

type DashboardRepository interface {
	Get(
		ctx context.Context,
		tenantID string,
		dayStart, dayEnd time.Time,
	) (DashboardMetrics, error)
	PlatformStats(ctx context.Context) (PlatformDashboardMetrics, error)
}
