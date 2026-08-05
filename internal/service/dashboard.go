package service

import (
	"context"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type ResourceTypeCounts struct {
	Video      int64 `json:"video"`
	Image      int64 `json:"image"`
	Document   int64 `json:"document"`
	Attachment int64 `json:"attachment"`
}

type LearningRankItem struct {
	UserID          string `json:"user_id"`
	DisplayName     string `json:"display_name"`
	DurationSeconds int64  `json:"duration_seconds"`
}

type TenantDashboard struct {
	Scope                      string             `json:"scope"`
	TodayLearningUserCount     int64              `json:"today_learning_user_count"`
	YesterdayLearningUserCount int64              `json:"yesterday_learning_user_count"`
	TodayLearningUserDelta     int64              `json:"today_learning_user_delta"`
	LearnerCount               int64              `json:"learner_count"`
	TodayNewLearnerCount       int64              `json:"today_new_learner_count"`
	PublishedCourseCount       int64              `json:"published_course_count"`
	CourseCount                int64              `json:"course_count"`
	ResourceCategoryCount      int64              `json:"resource_category_count"`
	ResourceCount              int64              `json:"resource_count"`
	ManagerCount               int64              `json:"manager_count"`
	HasDemoData                bool               `json:"has_demo_data"`
	ResourceTypeCounts         ResourceTypeCounts `json:"resource_type_counts"`
	TodayLearningRanking       []LearningRankItem `json:"today_learning_ranking"`
}

type PlatformTenant struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	Status          int       `json:"status"`
	LifecycleStatus string    `json:"lifecycle_status,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type PlatformDashboard struct {
	Scope             string           `json:"scope"`
	TenantCount       int64            `json:"tenant_count"`
	ActiveTenantCount int64            `json:"active_tenant_count"`
	LearnerCount      int64            `json:"learner_count"`
	CourseCount       int64            `json:"course_count"`
	RecentTenants     []PlatformTenant `json:"recent_tenants"`
}

type InstructorCourse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    int       `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InstructorDashboard struct {
	Scope                  string             `json:"scope"`
	CourseCount            int64              `json:"course_count"`
	PublishedCourseCount   int64              `json:"published_course_count"`
	TodayLearningUserCount int64              `json:"today_learning_user_count"`
	RecentCourses          []InstructorCourse `json:"recent_courses"`
}

type DashboardService struct {
	dashboard repository.DashboardRepository
	now       func() time.Time
}

func NewDashboardService(dashboard repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboard: dashboard, now: time.Now}
}

func (service *DashboardService) Stats(ctx context.Context) (interface{}, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || userID == "" {
		return nil, errorsx.Forbidden("permission denied")
	}
	switch role {
	case "superadmin":
		if tenantID != "" {
			return nil, errorsx.Forbidden("permission denied")
		}
		return service.platformStats(ctx)
	case "tenant_admin", "instructor":
		if tenantID == "" {
			return nil, errorsx.Forbidden("permission denied")
		}
	default:
		return nil, errorsx.Forbidden("permission denied")
	}

	todayDate, yesterdayDate, todayStart, todayEnd := service.dashboardDates()
	if role == "instructor" {
		metrics, err := service.dashboard.InstructorStats(
			ctx, tenantID, userID, todayDate, todayStart, todayEnd,
		)
		if err != nil {
			return nil, errorsx.Internal("load dashboard failed")
		}
		recent := make([]InstructorCourse, 0, len(metrics.RecentCourses))
		for _, course := range metrics.RecentCourses {
			recent = append(recent, InstructorCourse{
				ID: course.ID, Title: course.Title, Status: course.Status,
				UpdatedAt: course.UpdatedAt,
			})
		}
		return InstructorDashboard{
			Scope: "instructor", CourseCount: metrics.CourseCount,
			PublishedCourseCount:   metrics.PublishedCourseCount,
			TodayLearningUserCount: metrics.TodayLearningUserCount,
			RecentCourses:          recent,
		}, nil
	}

	metrics, err := service.dashboard.TenantStats(
		ctx, tenantID, todayDate, yesterdayDate, todayStart, todayEnd,
	)
	if err != nil {
		return nil, errorsx.Internal("load dashboard failed")
	}
	ranking := make([]LearningRankItem, 0, len(metrics.TodayLearningRanking))
	for _, item := range metrics.TodayLearningRanking {
		ranking = append(ranking, LearningRankItem{
			UserID: item.UserID, DisplayName: item.DisplayName,
			DurationSeconds: item.DurationSeconds,
		})
	}
	return TenantDashboard{
		Scope:                      "tenant",
		TodayLearningUserCount:     metrics.TodayLearningUserCount,
		YesterdayLearningUserCount: metrics.YesterdayLearningUserCount,
		TodayLearningUserDelta:     metrics.TodayLearningUserDelta,
		LearnerCount:               metrics.LearnerCount,
		TodayNewLearnerCount:       metrics.TodayNewLearnerCount,
		PublishedCourseCount:       metrics.PublishedCourseCount,
		CourseCount:                metrics.CourseCount,
		ResourceCategoryCount:      metrics.ResourceCategoryCount,
		ResourceCount:              metrics.ResourceCount,
		ManagerCount:               metrics.ManagerCount,
		HasDemoData:                metrics.HasDemoData,
		ResourceTypeCounts: ResourceTypeCounts{
			Video:      metrics.ResourceTypeCounts.Video,
			Image:      metrics.ResourceTypeCounts.Image,
			Document:   metrics.ResourceTypeCounts.Document,
			Attachment: metrics.ResourceTypeCounts.Attachment,
		},
		TodayLearningRanking: ranking,
	}, nil
}

func (service *DashboardService) platformStats(ctx context.Context) (PlatformDashboard, error) {
	metrics, err := service.dashboard.PlatformStats(ctx)
	if err != nil {
		return PlatformDashboard{}, errorsx.Internal("load dashboard failed")
	}
	recent := make([]PlatformTenant, 0, len(metrics.RecentTenants))
	for _, tenant := range metrics.RecentTenants {
		recent = append(recent, PlatformTenant{
			ID: tenant.ID, Name: tenant.Name, Code: tenant.Code,
			Status: tenant.Status, LifecycleStatus: tenant.LifecycleStatus,
			CreatedAt: tenant.CreatedAt,
		})
	}
	return PlatformDashboard{
		Scope: "platform", TenantCount: metrics.TenantCount,
		ActiveTenantCount: metrics.ActiveTenantCount,
		LearnerCount:      metrics.LearnerCount, CourseCount: metrics.CourseCount,
		RecentTenants: recent,
	}, nil
}

func (service *DashboardService) dashboardDates() (
	todayDate, yesterdayDate string,
	todayStart, todayEnd time.Time,
) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	now := service.now().In(location)
	localStart := time.Date(
		now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location,
	)
	localEnd := localStart.AddDate(0, 0, 1)
	return localStart.Format("2006-01-02"),
		localStart.AddDate(0, 0, -1).Format("2006-01-02"),
		localStart.UTC(), localEnd.UTC()
}
