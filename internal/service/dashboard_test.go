package service

import (
	"context"
	"errors"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestDashboardServiceReturnsMetricsForManagerRoles(t *testing.T) {
	for _, role := range []string{"tenant_admin", "instructor"} {
		t.Run(role, func(t *testing.T) {
			repo := &dashboardRepositoryStub{
				metrics: repository.DashboardMetrics{
					UserCount: 5, CourseCount: 3, PublishedCourseCount: 2,
					TodayNewUserCount: 1, TodayLearningUserCount: 4,
					TotalLearningSeconds: 7200, ActiveEnrollmentCount: 4,
					CompletedEnrollmentCount: 3,
				},
			}
			service := NewDashboardService(repo)
			service.now = func() time.Time {
				return time.Date(
					2026, 7, 27, 1, 30, 0, 0,
					time.FixedZone("UTC+8", 8*60*60),
				)
			}
			tenantID := "tenant-1"
			if role == "superadmin" {
				tenantID = ""
			}
			ctx := usercontext.WithUser(context.Background(), "manager", tenantID, "", role)
			got, err := service.Stats(ctx)
			if err != nil {
				t.Fatalf("Stats() error = %v", err)
			}
			if got.UserCount != 5 || got.CourseCount != 3 ||
				got.CourseCompletionRate != 0.75 {
				t.Fatalf("Stats() = %#v", got)
			}
			wantStart := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
			wantTenantID := "tenant-1"
			if role == "superadmin" {
				wantTenantID = ""
			}
			if repo.tenantID != wantTenantID ||
				!repo.dayStart.Equal(wantStart) ||
				!repo.dayEnd.Equal(wantStart.AddDate(0, 0, 1)) {
				t.Fatalf(
					"repository args = %q, %v, %v",
					repo.tenantID, repo.dayStart, repo.dayEnd,
				)
			}
		})
	}
}

func TestDashboardServiceReturnsPlatformMetricsForSuperadmin(t *testing.T) {
	repo := &dashboardRepositoryStub{platform: repository.PlatformDashboardMetrics{
		TenantCount: 4, ActiveTenantCount: 3, LearnerCount: 20, CourseCount: 8,
	}}
	service := NewDashboardService(repo)
	ctx := usercontext.WithUser(context.Background(), "root", "", "", "superadmin")
	stats, err := service.Stats(ctx)
	if err != nil || stats.Platform == nil || stats.Platform.TenantCount != 4 || stats.Platform.CourseCount != 8 {
		t.Fatalf("Stats() = %#v, %v", stats, err)
	}
	if !repo.platformCalled || repo.called {
		t.Fatal("wrong dashboard repository method called for superadmin")
	}
}

func TestDashboardServiceReturnsZeroRateWithoutEnrollments(t *testing.T) {
	service := NewDashboardService(&dashboardRepositoryStub{})
	ctx := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "", "tenant_admin",
	)
	stats, err := service.Stats(ctx)
	if err != nil || stats.CourseCompletionRate != 0 {
		t.Fatalf("Stats() = %#v, %v", stats, err)
	}
}

func TestDashboardServiceRejectsUnauthorizedRolesAndMapsErrors(t *testing.T) {
	for _, test := range []struct {
		name     string
		role     string
		tenantID string
	}{
		{"learner", "learner", "tenant-1"},
		{"missing tenant", "instructor", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &dashboardRepositoryStub{}
			service := NewDashboardService(repo)
			ctx := usercontext.WithUser(
				context.Background(), "user", test.tenantID, "", test.role,
			)
			if _, err := service.Stats(ctx); errorCode(err) != 40300 {
				t.Fatalf("Stats() error = %#v", err)
			}
			if repo.called {
				t.Fatal("repository called for unauthorized request")
			}
		})
	}

	repo := &dashboardRepositoryStub{err: errors.New("database unavailable")}
	service := NewDashboardService(repo)
	ctx := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "", "tenant_admin",
	)
	if _, err := service.Stats(ctx); errorCode(err) != 50000 {
		t.Fatalf("Stats(database error) = %#v", err)
	}
}

type dashboardRepositoryStub struct {
	metrics          repository.DashboardMetrics
	platform         repository.PlatformDashboardMetrics
	err              error
	called           bool
	platformCalled   bool
	tenantID         string
	dayStart, dayEnd time.Time
}

func (stub *dashboardRepositoryStub) Get(
	_ context.Context, tenantID string, dayStart, dayEnd time.Time,
) (repository.DashboardMetrics, error) {
	stub.called = true
	stub.tenantID, stub.dayStart, stub.dayEnd = tenantID, dayStart, dayEnd
	return stub.metrics, stub.err
}

func (stub *dashboardRepositoryStub) PlatformStats(
	_ context.Context,
) (repository.PlatformDashboardMetrics, error) {
	stub.platformCalled = true
	return stub.platform, stub.err
}
