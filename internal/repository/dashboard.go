package repository

import (
	"context"
	"time"
)

type ResourceTypeCounts struct {
	Video      int64
	Image      int64
	Document   int64
	Attachment int64
}

type LearningRankItem struct {
	UserID          string
	DisplayName     string
	DurationSeconds int64
}

type TenantDashboardMetrics struct {
	TodayLearningUserCount     int64
	YesterdayLearningUserCount int64
	TodayLearningUserDelta     int64
	LearnerCount               int64
	TodayNewLearnerCount       int64
	PublishedCourseCount       int64
	CourseCount                int64
	ResourceCategoryCount      int64
	ResourceCount              int64
	ManagerCount               int64
	HasDemoData                bool
	ResourceTypeCounts         ResourceTypeCounts
	TodayLearningRanking       []LearningRankItem
}

type PlatformTenant struct {
	ID              string
	Name            string
	Code            string
	Status          int
	LifecycleStatus string
	CreatedAt       time.Time
}

type PlatformDashboardMetrics struct {
	TenantCount       int64
	ActiveTenantCount int64
	LearnerCount      int64
	CourseCount       int64
	RecentTenants     []PlatformTenant
}

type InstructorCourse struct {
	ID        string
	Title     string
	Status    int
	UpdatedAt time.Time
}

type InstructorDashboardMetrics struct {
	CourseCount            int64
	PublishedCourseCount   int64
	TodayLearningUserCount int64
	RecentCourses          []InstructorCourse
}

type DashboardRepository interface {
	TenantStats(
		ctx context.Context,
		tenantID, todayDate, yesterdayDate string,
		todayStart, todayEnd time.Time,
	) (TenantDashboardMetrics, error)
	InstructorStats(
		ctx context.Context,
		tenantID, userID, todayDate string,
		todayStart, todayEnd time.Time,
	) (InstructorDashboardMetrics, error)
	PlatformStats(ctx context.Context) (PlatformDashboardMetrics, error)
}
