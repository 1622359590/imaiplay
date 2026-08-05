package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestDashboardServiceReturnsFixedRoleScopedDTOs(t *testing.T) {
	now := time.Date(2026, 8, 4, 16, 30, 0, 0, time.UTC)
	repo := &dashboardRepositoryStub{
		tenant: repository.TenantDashboardMetrics{
			TodayLearningUserCount: 2, YesterdayLearningUserCount: 1,
			TodayLearningUserDelta: 1, LearnerCount: 4,
			TodayNewLearnerCount: 1, PublishedCourseCount: 3, CourseCount: 5,
			ResourceCategoryCount: 2, ResourceCount: 4, ManagerCount: 2,
			HasDemoData: true,
			ResourceTypeCounts: repository.ResourceTypeCounts{
				Video: 2, Image: 1, Document: 1, Attachment: 0,
			},
			TodayLearningRanking: []repository.LearningRankItem{{
				UserID: "learner-1", DisplayName: "Learner One", DurationSeconds: 120,
			}},
		},
		instructor: repository.InstructorDashboardMetrics{
			CourseCount: 3, PublishedCourseCount: 2, TodayLearningUserCount: 1,
			RecentCourses: []repository.InstructorCourse{{
				ID: "course-1", Title: "Owned", Status: 1,
				UpdatedAt: time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC),
			}},
		},
		platform: repository.PlatformDashboardMetrics{
			TenantCount: 4, ActiveTenantCount: 3, LearnerCount: 20, CourseCount: 8,
			RecentTenants: []repository.PlatformTenant{{
				ID: "tenant-1", Name: "One", Code: "one", Status: 1,
				LifecycleStatus: "active", CreatedAt: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
			}},
		},
	}
	service := NewDashboardService(repo)
	service.now = func() time.Time { return now }

	tenantResult, err := service.Stats(usercontext.WithUser(
		context.Background(), "admin-1", "tenant-1", "", "tenant_admin",
	))
	if err != nil {
		t.Fatal(err)
	}
	tenant, ok := tenantResult.(TenantDashboard)
	if !ok || tenant.Scope != "tenant" || tenant.LearnerCount != 4 ||
		tenant.ResourceTypeCounts.Attachment != 0 || len(tenant.TodayLearningRanking) != 1 {
		t.Fatalf("tenant dashboard = %#v", tenantResult)
	}

	instructorResult, err := service.Stats(usercontext.WithUser(
		context.Background(), "instructor-1", "tenant-1", "", "instructor",
	))
	if err != nil {
		t.Fatal(err)
	}
	instructor, ok := instructorResult.(InstructorDashboard)
	if !ok || instructor.Scope != "instructor" || instructor.CourseCount != 3 ||
		len(instructor.RecentCourses) != 1 {
		t.Fatalf("instructor dashboard = %#v", instructorResult)
	}

	platformResult, err := service.Stats(usercontext.WithUser(
		context.Background(), "root", "", "", "superadmin",
	))
	if err != nil {
		t.Fatal(err)
	}
	platform, ok := platformResult.(PlatformDashboard)
	if !ok || platform.Scope != "platform" || platform.TenantCount != 4 ||
		len(platform.RecentTenants) != 1 {
		t.Fatalf("platform dashboard = %#v", platformResult)
	}

	if repo.tenantCalls != 1 || repo.instructorCalls != 1 || repo.platformCalls != 1 {
		t.Fatalf("repository calls = tenant:%d instructor:%d platform:%d", repo.tenantCalls, repo.instructorCalls, repo.platformCalls)
	}
	if repo.tenantID != "tenant-1" || repo.userID != "instructor-1" ||
		repo.todayDate != "2026-08-05" || repo.yesterdayDate != "2026-08-04" {
		t.Fatalf("repository scope/date args = %#v", repo)
	}
	wantStart := time.Date(2026, 8, 4, 16, 0, 0, 0, time.UTC)
	if !repo.todayStart.Equal(wantStart) || !repo.todayEnd.Equal(wantStart.Add(24*time.Hour)) {
		t.Fatalf("Shanghai bounds = %v..%v", repo.todayStart, repo.todayEnd)
	}
}

func TestDashboardDTOJSONContractsAreExact(t *testing.T) {
	assertJSONKeys(t, TenantDashboard{}, []string{
		"scope", "today_learning_user_count", "yesterday_learning_user_count",
		"today_learning_user_delta", "learner_count", "today_new_learner_count",
		"published_course_count", "course_count", "resource_category_count",
		"resource_count", "manager_count", "has_demo_data", "resource_type_counts",
		"today_learning_ranking",
	})
	assertJSONKeys(t, ResourceTypeCounts{}, []string{"video", "image", "document", "attachment"})
	assertJSONKeys(t, LearningRankItem{}, []string{"user_id", "display_name", "duration_seconds"})
	assertJSONKeys(t, PlatformDashboard{}, []string{
		"scope", "tenant_count", "active_tenant_count", "learner_count", "course_count", "recent_tenants",
	})
	assertJSONKeys(t, PlatformTenant{}, []string{"id", "name", "code", "status", "lifecycle_status", "created_at"})
	assertJSONKeys(t, InstructorDashboard{}, []string{
		"scope", "course_count", "published_course_count", "today_learning_user_count", "recent_courses",
	})
	assertJSONKeys(t, InstructorCourse{}, []string{"id", "title", "status", "updated_at"})
}

func TestDashboardServiceRejectsInvalidScopeAndMapsRepositoryErrors(t *testing.T) {
	for _, test := range []struct{ role, tenantID string }{
		{role: "learner", tenantID: "tenant-1"},
		{role: "instructor", tenantID: ""},
		{role: "superadmin", tenantID: "tenant-1"},
	} {
		repo := &dashboardRepositoryStub{}
		service := NewDashboardService(repo)
		_, err := service.Stats(usercontext.WithUser(
			context.Background(), "user-1", test.tenantID, "", test.role,
		))
		if errorCode(err) != 40300 {
			t.Fatalf("Stats(%s/%s) error = %#v", test.role, test.tenantID, err)
		}
		if repo.tenantCalls+repo.instructorCalls+repo.platformCalls != 0 {
			t.Fatal("repository called for invalid scope")
		}
	}

	repo := &dashboardRepositoryStub{err: errors.New("database unavailable")}
	service := NewDashboardService(repo)
	_, err := service.Stats(usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "", "tenant_admin",
	))
	if errorCode(err) != 50000 {
		t.Fatalf("Stats(database error) = %#v", err)
	}
}

func assertJSONKeys(t *testing.T, value interface{}, want []string) {
	t.Helper()
	typeOf := reflect.TypeOf(value)
	got := make([]string, 0, typeOf.NumField())
	for index := 0; index < typeOf.NumField(); index++ {
		tag := typeOf.Field(index).Tag.Get("json")
		for charIndex, char := range tag {
			if char == ',' {
				tag = tag[:charIndex]
				break
			}
		}
		got = append(got, tag)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%T JSON keys = %#v, want %#v", value, got, want)
	}
}

type dashboardRepositoryStub struct {
	tenant                                      repository.TenantDashboardMetrics
	instructor                                  repository.InstructorDashboardMetrics
	platform                                    repository.PlatformDashboardMetrics
	err                                         error
	tenantCalls, instructorCalls, platformCalls int
	tenantID, userID, todayDate, yesterdayDate  string
	todayStart, todayEnd                        time.Time
}

func (stub *dashboardRepositoryStub) TenantStats(
	_ context.Context, tenantID, todayDate, yesterdayDate string,
	todayStart, todayEnd time.Time,
) (repository.TenantDashboardMetrics, error) {
	stub.tenantCalls++
	stub.tenantID, stub.todayDate, stub.yesterdayDate = tenantID, todayDate, yesterdayDate
	stub.todayStart, stub.todayEnd = todayStart, todayEnd
	return stub.tenant, stub.err
}

func (stub *dashboardRepositoryStub) InstructorStats(
	_ context.Context, tenantID, userID, todayDate string,
	todayStart, todayEnd time.Time,
) (repository.InstructorDashboardMetrics, error) {
	stub.instructorCalls++
	stub.tenantID, stub.userID, stub.todayDate = tenantID, userID, todayDate
	stub.todayStart, stub.todayEnd = todayStart, todayEnd
	return stub.instructor, stub.err
}

func (stub *dashboardRepositoryStub) PlatformStats(
	context.Context,
) (repository.PlatformDashboardMetrics, error) {
	stub.platformCalls++
	return stub.platform, stub.err
}
