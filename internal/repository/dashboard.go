package repository

import (
	"context"
	"time"
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

type DashboardRepository interface {
	Get(
		ctx context.Context,
		tenantID string,
		dayStart, dayEnd time.Time,
	) (DashboardMetrics, error)
}
