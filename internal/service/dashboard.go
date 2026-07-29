package service

import (
	"context"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type DashboardStats struct {
	UserCount              int64          `json:"user_count"`
	CourseCount            int64          `json:"course_count"`
	PublishedCourseCount   int64          `json:"published_course_count"`
	TodayNewUserCount      int64          `json:"today_new_user_count"`
	TodayLearningUserCount int64          `json:"today_learning_user_count"`
	TotalLearningSeconds   int64          `json:"total_learning_seconds"`
	CourseCompletionRate   float64        `json:"course_completion_rate"`
	Platform               *PlatformStats `json:"platform,omitempty"`
}

type PlatformStats struct {
	TenantCount       int64            `json:"tenant_count"`
	ActiveTenantCount int64            `json:"active_tenant_count"`
	LearnerCount      int64            `json:"learner_count"`
	CourseCount       int64            `json:"course_count"`
	RecentTenants     []PlatformTenant `json:"recent_tenants"`
}

type PlatformTenant struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	Status          int       `json:"status"`
	LifecycleStatus string    `json:"lifecycle_status,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type DashboardService struct {
	dashboard repository.DashboardRepository
	now       func() time.Time
}

func NewDashboardService(
	dashboard repository.DashboardRepository,
) *DashboardService {
	return &DashboardService{dashboard: dashboard, now: time.Now}
}

func (service *DashboardService) Stats(
	ctx context.Context,
) (DashboardStats, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (role != "superadmin" && tenantID == "") ||
		(role != "tenant_admin" && role != "instructor" && role != "superadmin") {
		return DashboardStats{}, errorsx.Forbidden("permission denied")
	}
	if role == "superadmin" {
		metrics, err := service.dashboard.PlatformStats(ctx)
		if err != nil {
			return DashboardStats{}, errorsx.Internal("load dashboard failed")
		}
		recentTenants := make([]PlatformTenant, 0, len(metrics.RecentTenants))
		for _, tenant := range metrics.RecentTenants {
			recentTenants = append(recentTenants, PlatformTenant{
				ID: tenant.ID, Name: tenant.Name, Code: tenant.Code,
				Status: tenant.Status, LifecycleStatus: tenant.LifecycleStatus,
				CreatedAt: tenant.CreatedAt,
			})
		}
		return DashboardStats{Platform: &PlatformStats{
			TenantCount: metrics.TenantCount, ActiveTenantCount: metrics.ActiveTenantCount,
			LearnerCount: metrics.LearnerCount, CourseCount: metrics.CourseCount,
			RecentTenants: recentTenants,
		}}, nil
	}
	now := service.now().UTC()
	dayStart := time.Date(
		now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC,
	)
	metrics, err := service.dashboard.Get(
		ctx, tenantID, dayStart, dayStart.AddDate(0, 0, 1),
	)
	if err != nil {
		return DashboardStats{}, errorsx.Internal("load dashboard failed")
	}
	rate := 0.0
	if metrics.ActiveEnrollmentCount > 0 {
		rate = float64(metrics.CompletedEnrollmentCount) /
			float64(metrics.ActiveEnrollmentCount)
	}
	return DashboardStats{
		UserCount:              metrics.UserCount,
		CourseCount:            metrics.CourseCount,
		PublishedCourseCount:   metrics.PublishedCourseCount,
		TodayNewUserCount:      metrics.TodayNewUserCount,
		TodayLearningUserCount: metrics.TodayLearningUserCount,
		TotalLearningSeconds:   metrics.TotalLearningSeconds,
		CourseCompletionRate:   rate,
	}, nil
}
