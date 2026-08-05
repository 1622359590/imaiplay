package repository

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestDashboardRepositoryTenantMetricsUseStrictTenantActiveAndVisibleScopes(t *testing.T) {
	database := dashboardDatabase(t)
	start := time.Date(2026, 8, 4, 16, 0, 0, 0, time.UTC)
	for _, tenant := range []*domain.Tenant{
		{ID: "tenant-1", Code: "one", Name: "One", Status: 1, LifecycleStatus: "active"},
		{ID: "tenant-2", Code: "two", Name: "Two", Status: 1, LifecycleStatus: "active"},
	} {
		dashboardCreate(t, database, tenant)
	}
	for index := 1; index <= 12; index++ {
		id := fmt.Sprintf("learner-%02d", index)
		name := fmt.Sprintf("Learner %02d", index)
		if index == 10 {
			name = "Alpha Tie"
		}
		if index == 11 {
			name = "Zulu Tie"
		}
		if index == 8 || index == 9 {
			name = "Same Tie"
		}
		user := &domain.User{
			BaseModel: domain.BaseModel{ID: id, TenantID: "tenant-1"},
			Email:     id + "@example.com", Password: "hash", Name: name,
			Role: "learner", Status: 1,
		}
		dashboardCreate(t, database, user)
		createdAt := start.Add(-time.Hour)
		if index == 1 {
			createdAt = start.Add(time.Hour)
		}
		dashboardSetTime(t, database, user, "created_at", createdAt)
		duration := int64(210 - index*10)
		if index == 9 {
			duration = 130
		}
		if index == 11 {
			duration = 110
		}
		dashboardCreate(t, database, &domain.LearningDailyStat{
			BaseModel: domain.BaseModel{ID: "today-" + id, TenantID: "tenant-1"},
			UserID:    id, StudyDate: "2026-08-05", DurationSeconds: duration,
		})
	}
	dashboardCreate(t, database, &domain.LearningDailyStat{
		BaseModel: domain.BaseModel{ID: "yesterday", TenantID: "tenant-1"},
		UserID:    "learner-01", StudyDate: "2026-08-04", DurationSeconds: 30,
	})
	for _, user := range []*domain.User{
		{BaseModel: domain.BaseModel{ID: "inactive", TenantID: "tenant-1"}, Email: "inactive@example.com", Password: "hash", Name: "Inactive", Role: "learner", Status: 0},
		{BaseModel: domain.BaseModel{ID: "admin", TenantID: "tenant-1"}, Email: "admin@example.com", Password: "hash", Name: "Admin", Role: "tenant_admin", Status: 1},
		{BaseModel: domain.BaseModel{ID: "instructor", TenantID: "tenant-1"}, Email: "instructor@example.com", Password: "hash", Name: "Instructor", Role: "instructor", Status: 1},
		{BaseModel: domain.BaseModel{ID: "disabled-admin", TenantID: "tenant-1"}, Email: "disabled-admin@example.com", Password: "hash", Name: "Disabled", Role: "tenant_admin", Status: 0},
		{BaseModel: domain.BaseModel{ID: "foreign", TenantID: "tenant-2"}, Email: "foreign@example.com", Password: "hash", Name: "Foreign", Role: "learner", Status: 1},
	} {
		dashboardCreate(t, database, user)
	}
	if err := database.Model(&domain.User{}).
		Where("id IN ?", []string{"inactive", "disabled-admin"}).
		Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	dashboardCreate(t, database, &domain.LearningDailyStat{BaseModel: domain.BaseModel{ID: "inactive-stat", TenantID: "tenant-1"}, UserID: "inactive", StudyDate: "2026-08-05", DurationSeconds: 999})
	dashboardCreate(t, database, &domain.LearningDailyStat{BaseModel: domain.BaseModel{ID: "foreign-stat", TenantID: "tenant-2"}, UserID: "foreign", StudyDate: "2026-08-05", DurationSeconds: 999})

	ownedPublished := &domain.Course{BaseModel: domain.BaseModel{ID: "owned-published", TenantID: "tenant-1"}, Title: "Owned published", Status: 1, CreatedBy: "admin"}
	ownedDraft := &domain.Course{BaseModel: domain.BaseModel{ID: "owned-draft", TenantID: "tenant-1"}, Title: "Owned draft", Status: 0, CreatedBy: "admin"}
	officialPublished := &domain.Course{BaseModel: domain.BaseModel{ID: "official-published", TenantID: ""}, Title: "Official published", Status: 1, CreatedBy: "root", IsOfficial: true}
	officialDraft := &domain.Course{BaseModel: domain.BaseModel{ID: "official-draft", TenantID: ""}, Title: "Official draft", Status: 0, CreatedBy: "root", IsOfficial: true}
	officialDisabled := &domain.Course{BaseModel: domain.BaseModel{ID: "official-disabled", TenantID: ""}, Title: "Official disabled", Status: 1, CreatedBy: "root", IsOfficial: true}
	for _, course := range []*domain.Course{ownedPublished, ownedDraft, officialPublished, officialDraft, officialDisabled} {
		dashboardCreate(t, database, course)
	}
	for _, mapping := range []*domain.TenantOfficialCourse{
		{TenantID: "tenant-1", CourseID: officialPublished.ID, Enabled: true},
		{TenantID: "tenant-1", CourseID: officialDraft.ID, Enabled: true},
		{TenantID: "tenant-1", CourseID: officialDisabled.ID, Enabled: false},
		{TenantID: "tenant-2", CourseID: officialDisabled.ID, Enabled: true},
	} {
		dashboardCreate(t, database, mapping)
	}
	for index := 1; index <= 2; index++ {
		dashboardCreate(t, database, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: fmt.Sprintf("category-%d", index), TenantID: "tenant-1"}, Name: "Category"})
	}
	for _, resource := range []*domain.Resource{
		{BaseModel: domain.BaseModel{ID: "video-1", TenantID: "tenant-1"}, Name: "Video 1", ResourceType: "video", URL: "/video-1", CreatedBy: "admin"},
		{BaseModel: domain.BaseModel{ID: "video-2", TenantID: "tenant-1"}, Name: "Video 2", ResourceType: "video", URL: "/video-2", CreatedBy: "admin"},
		{BaseModel: domain.BaseModel{ID: "image-1", TenantID: "tenant-1"}, Name: "Image", ResourceType: "image", URL: "/image", CreatedBy: "admin"},
		{BaseModel: domain.BaseModel{ID: "foreign-resource", TenantID: "tenant-2"}, Name: "Foreign", ResourceType: "attachment", URL: "/foreign", CreatedBy: "foreign"},
	} {
		dashboardCreate(t, database, resource)
	}
	dashboardCreate(t, database, &domain.TenantDemoRecord{
		BaseModel: domain.BaseModel{ID: "demo", TenantID: "tenant-1"},
		BatchID:   "batch", RecordType: DemoRecordResource, RecordID: "video-1",
	})

	got, err := NewDashboardRepository(database).TenantStats(
		context.Background(), "tenant-1", "2026-08-05", "2026-08-04", start, start.Add(24*time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.LearnerCount != 12 || got.TodayNewLearnerCount != 1 ||
		got.TodayLearningUserCount != 12 || got.YesterdayLearningUserCount != 1 ||
		got.TodayLearningUserDelta != 11 || got.ManagerCount != 2 ||
		got.CourseCount != 4 || got.PublishedCourseCount != 2 ||
		got.ResourceCategoryCount != 2 || got.ResourceCount != 3 || !got.HasDemoData {
		t.Fatalf("TenantStats() = %#v", got)
	}
	if got.ResourceTypeCounts != (ResourceTypeCounts{Video: 2, Image: 1}) {
		t.Fatalf("resource counts = %#v", got.ResourceTypeCounts)
	}
	if len(got.TodayLearningRanking) != 10 ||
		got.TodayLearningRanking[7].UserID != "learner-08" ||
		got.TodayLearningRanking[8].UserID != "learner-09" ||
		got.TodayLearningRanking[9].UserID != "learner-10" {
		t.Fatalf("ranking = %#v", got.TodayLearningRanking)
	}
}

func TestDashboardRepositoryTenantMetricsReturnStableZeros(t *testing.T) {
	database := dashboardDatabase(t)
	dashboardCreate(t, database, &domain.Tenant{ID: "empty", Code: "empty", Name: "Empty", Status: 1})
	dashboardCreate(t, database, &domain.TenantDemoRecord{
		BaseModel: domain.BaseModel{ID: "foreign-demo", TenantID: "other-tenant"},
		BatchID:   "foreign", RecordType: DemoRecordResource, RecordID: "foreign-resource",
	})
	start := time.Date(2026, 8, 4, 16, 0, 0, 0, time.UTC)
	got, err := NewDashboardRepository(database).TenantStats(
		context.Background(), "empty", "2026-08-05", "2026-08-04", start, start.Add(24*time.Hour),
	)
	want := TenantDashboardMetrics{TodayLearningRanking: []LearningRankItem{}}
	if err != nil || !reflect.DeepEqual(got, want) || got.TodayLearningRanking == nil {
		t.Fatalf("empty TenantStats() = %#v, %v", got, err)
	}
}

func TestDashboardRepositoryInstructorMetricsRequireOwnedCourseLearning(t *testing.T) {
	database := dashboardDatabase(t)
	dashboardCreate(t, database, &domain.Tenant{ID: "tenant-1", Code: "one", Name: "One", Status: 1})
	start := time.Date(2026, 8, 4, 16, 0, 0, 0, time.UTC)
	for _, user := range []*domain.User{
		{BaseModel: domain.BaseModel{ID: "instructor-1", TenantID: "tenant-1"}, Email: "i@example.com", Password: "hash", Name: "I", Role: "instructor", Status: 1},
		{BaseModel: domain.BaseModel{ID: "owned-learner", TenantID: "tenant-1"}, Email: "owned@example.com", Password: "hash", Name: "Owned", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "other-learner", TenantID: "tenant-1"}, Email: "other@example.com", Password: "hash", Name: "Other", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "inactive-learner", TenantID: "tenant-1"}, Email: "inactive@example.com", Password: "hash", Name: "Inactive", Role: "learner", Status: 0},
	} {
		dashboardCreate(t, database, user)
	}
	if err := database.Model(&domain.User{}).Where("id = ?", "inactive-learner").
		Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"owned-learner", "other-learner", "inactive-learner"} {
		dashboardCreate(t, database, &domain.LearningDailyStat{BaseModel: domain.BaseModel{ID: "stat-" + id, TenantID: "tenant-1"}, UserID: id, StudyDate: "2026-08-05", DurationSeconds: 30})
	}
	var ownedCourses []*domain.Course
	for index := 1; index <= 7; index++ {
		course := &domain.Course{BaseModel: domain.BaseModel{ID: fmt.Sprintf("owned-%d", index), TenantID: "tenant-1"}, Title: fmt.Sprintf("Owned %d", index), Status: index % 2, CreatedBy: "instructor-1"}
		dashboardCreate(t, database, course)
		updatedAt := start.Add(time.Duration(8-index) * time.Hour)
		if index == 2 {
			updatedAt = start.Add(7 * time.Hour)
		}
		dashboardSetTime(t, database, course, "updated_at", updatedAt)
		ownedCourses = append(ownedCourses, course)
	}
	otherCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "other-course", TenantID: "tenant-1"}, Title: "Other", Status: 1, CreatedBy: "instructor-2"}
	official := &domain.Course{BaseModel: domain.BaseModel{ID: "official", TenantID: ""}, Title: "Official", Status: 1, CreatedBy: "instructor-1", IsOfficial: true}
	dashboardCreate(t, database, otherCourse)
	dashboardCreate(t, database, official)
	createLearning := func(prefix, learnerID string, course *domain.Course) {
		chapter := &domain.CourseChapter{BaseModel: domain.BaseModel{ID: prefix + "-chapter", TenantID: "tenant-1"}, CourseID: course.ID, Title: prefix}
		lesson := &domain.CourseLesson{BaseModel: domain.BaseModel{ID: prefix + "-lesson", TenantID: "tenant-1"}, ChapterID: chapter.ID, Title: prefix, ContentType: "video"}
		dashboardCreate(t, database, chapter)
		dashboardCreate(t, database, lesson)
		dashboardCreate(t, database, &domain.CourseEnrollment{BaseModel: domain.BaseModel{ID: prefix + "-enrollment", TenantID: "tenant-1"}, CourseID: course.ID, UserID: learnerID, Status: 1})
		report := &domain.LearningTimeReport{BaseModel: domain.BaseModel{ID: prefix + "-report", TenantID: "tenant-1"}, UserID: learnerID, LessonID: lesson.ID, ReportID: prefix, WatchedSecondsDelta: 15}
		dashboardCreate(t, database, report)
		dashboardSetTime(t, database, report, "created_at", start.Add(time.Hour))
	}
	createLearning("owned", "owned-learner", ownedCourses[0])
	createLearning("other", "other-learner", otherCourse)
	createLearning("inactive", "inactive-learner", ownedCourses[0])

	got, err := NewDashboardRepository(database).InstructorStats(
		context.Background(), "tenant-1", "instructor-1", "2026-08-05", start, start.Add(24*time.Hour),
	)
	if err != nil || got.CourseCount != 7 || got.PublishedCourseCount != 4 ||
		got.TodayLearningUserCount != 1 || len(got.RecentCourses) != 5 {
		t.Fatalf("InstructorStats() = %#v, %v", got, err)
	}
	if got.RecentCourses[0].ID != "owned-1" || got.RecentCourses[1].ID != "owned-2" {
		t.Fatalf("stable recent order = %#v", got.RecentCourses)
	}
	for _, course := range got.RecentCourses {
		if course.ID == otherCourse.ID || course.ID == official.ID {
			t.Fatalf("recent course escaped ownership: %#v", course)
		}
	}
}

func TestDashboardRepositoryPlatformMetricsAreSafeStableAndLimited(t *testing.T) {
	database := dashboardDatabase(t)
	created := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	for index := 1; index <= 6; index++ {
		tenant := &domain.Tenant{ID: fmt.Sprintf("tenant-%d", index), Code: fmt.Sprintf("code-%d", index), Name: fmt.Sprintf("Tenant %d", index), Status: 1, LifecycleStatus: "active"}
		if index == 6 {
			tenant.Status, tenant.LifecycleStatus = 0, "suspended"
		}
		dashboardCreate(t, database, tenant)
		if index == 6 {
			if err := database.Model(&domain.Tenant{}).Where("id = ?", tenant.ID).
				Update("status", 0).Error; err != nil {
				t.Fatal(err)
			}
		}
		createdAt := created.Add(time.Duration(index) * time.Hour)
		if index == 5 {
			createdAt = created.Add(6 * time.Hour)
		}
		dashboardSetTime(t, database, tenant, "created_at", createdAt)
	}
	dashboardCreate(t, database, &domain.User{BaseModel: domain.BaseModel{ID: "platform-learner", TenantID: "tenant-1"}, Email: "learner@example.com", Password: "hash", Name: "Learner", Role: "learner", Status: 1})
	dashboardCreate(t, database, &domain.User{BaseModel: domain.BaseModel{ID: "platform-disabled", TenantID: "tenant-1"}, Email: "disabled@example.com", Password: "hash", Name: "Disabled", Role: "learner", Status: 0})
	if err := database.Model(&domain.User{}).Where("id = ?", "platform-disabled").
		Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	dashboardCreate(t, database, &domain.Course{BaseModel: domain.BaseModel{ID: "platform-course", TenantID: "tenant-1"}, Title: "Course", CreatedBy: "admin"})

	got, err := NewDashboardRepository(database).PlatformStats(context.Background())
	if err != nil || got.TenantCount != 6 || got.ActiveTenantCount != 5 ||
		got.LearnerCount != 1 || got.CourseCount != 1 || len(got.RecentTenants) != 5 {
		t.Fatalf("PlatformStats() = %#v, %v", got, err)
	}
	if got.RecentTenants[0].ID != "tenant-5" || got.RecentTenants[1].ID != "tenant-6" ||
		got.RecentTenants[1].LifecycleStatus != "suspended" {
		t.Fatalf("recent tenants = %#v", got.RecentTenants)
	}
}

func dashboardDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return database
}

func dashboardCreate(t *testing.T, database *gorm.DB, value interface{}) {
	t.Helper()
	if err := database.Create(value).Error; err != nil {
		t.Fatalf("create %T: %v", value, err)
	}
}

func dashboardSetTime(t *testing.T, database *gorm.DB, value interface{}, column string, at time.Time) {
	t.Helper()
	if err := database.Model(value).UpdateColumn(column, at).Error; err != nil {
		t.Fatalf("set %s on %T: %v", column, value, err)
	}
}
