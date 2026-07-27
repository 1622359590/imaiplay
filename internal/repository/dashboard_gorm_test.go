package repository

import (
	"context"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestDashboardRepositoryAggregatesTenantMetrics(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	dayStart := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)
	createDashboardFixture(t, database, dayStart)

	got, err := NewDashboardRepository(database).Get(
		context.Background(), "tenant-1", dayStart, dayEnd,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	want := DashboardMetrics{
		UserCount: 2, CourseCount: 2, PublishedCourseCount: 1,
		TodayNewUserCount: 1, TodayLearningUserCount: 2,
		TotalLearningSeconds: 240, ActiveEnrollmentCount: 2,
		CompletedEnrollmentCount: 1,
	}
	if got != want {
		t.Fatalf("Get() = %#v, want %#v", got, want)
	}
}

func createDashboardFixture(
	t *testing.T, database *gorm.DB, dayStart time.Time,
) {
	t.Helper()
	for _, tenant := range []*domain.Tenant{
		{ID: "tenant-1", Code: "one", Name: "One", Status: 1},
		{ID: "tenant-2", Code: "two", Name: "Two", Status: 1},
	} {
		mustCreateDashboard(t, database, tenant)
	}
	users := []*domain.User{
		{BaseModel: domain.BaseModel{ID: "user-1", TenantID: "tenant-1"}, Email: "one@example.com", Password: "hash", Name: "One", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "user-2", TenantID: "tenant-1"}, Email: "two@example.com", Password: "hash", Name: "Two", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "disabled", TenantID: "tenant-1"}, Email: "disabled@example.com", Password: "hash", Name: "Disabled", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "foreign", TenantID: "tenant-2"}, Email: "foreign@example.com", Password: "hash", Name: "Foreign", Role: "learner", Status: 1},
	}
	for _, user := range users {
		mustCreateDashboard(t, database, user)
	}
	if err := database.Model(users[2]).Update("status", 0).Error; err != nil {
		t.Fatalf("disable user: %v", err)
	}
	setDashboardTime(t, database, users[0], "created_at", dayStart.Add(time.Hour))
	for _, user := range users[1:] {
		setDashboardTime(t, database, user, "created_at", dayStart.Add(-time.Hour))
	}

	courseOne := &domain.Course{
		BaseModel: domain.BaseModel{ID: "course-1", TenantID: "tenant-1"},
		Title:     "Published", Status: 1, CreatedBy: "admin",
	}
	courseTwo := &domain.Course{
		BaseModel: domain.BaseModel{ID: "course-2", TenantID: "tenant-1"},
		Title:     "Draft", Status: 0, CreatedBy: "admin",
	}
	foreignCourse := &domain.Course{
		BaseModel: domain.BaseModel{ID: "foreign-course", TenantID: "tenant-2"},
		Title:     "Foreign", Status: 1, CreatedBy: "foreign-admin",
	}
	for _, course := range []*domain.Course{courseOne, courseTwo, foreignCourse} {
		mustCreateDashboard(t, database, course)
	}
	chapterOne := &domain.CourseChapter{
		BaseModel: domain.BaseModel{ID: "chapter-1", TenantID: "tenant-1"},
		CourseID:  courseOne.ID, Title: "One",
	}
	chapterTwo := &domain.CourseChapter{
		BaseModel: domain.BaseModel{ID: "chapter-2", TenantID: "tenant-1"},
		CourseID:  courseTwo.ID, Title: "Two",
	}
	foreignChapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{ID: "foreign-chapter", TenantID: "tenant-2"},
		CourseID:  foreignCourse.ID, Title: "Foreign",
	}
	for _, chapter := range []*domain.CourseChapter{
		chapterOne, chapterTwo, foreignChapter,
	} {
		mustCreateDashboard(t, database, chapter)
	}
	lessons := []*domain.CourseLesson{
		dashboardLesson("lesson-1", "tenant-1", chapterOne.ID),
		dashboardLesson("lesson-2", "tenant-1", chapterOne.ID),
		dashboardLesson("lesson-3", "tenant-1", chapterTwo.ID),
		dashboardLesson("lesson-4", "tenant-1", chapterTwo.ID),
		dashboardLesson("foreign-lesson", "tenant-2", foreignChapter.ID),
	}
	for _, lesson := range lessons {
		mustCreateDashboard(t, database, lesson)
	}
	enrollments := []*domain.CourseEnrollment{
		dashboardEnrollment("enrollment-1", "tenant-1", courseOne.ID, users[0].ID),
		dashboardEnrollment("enrollment-2", "tenant-1", courseTwo.ID, users[1].ID),
		dashboardEnrollment("foreign-enrollment", "tenant-2", foreignCourse.ID, users[3].ID),
	}
	for _, enrollment := range enrollments {
		mustCreateDashboard(t, database, enrollment)
	}
	progressItems := []*domain.LessonProgress{
		dashboardProgress("progress-1", "tenant-1", users[0].ID, lessons[0].ID, 2, 100),
		dashboardProgress("progress-2", "tenant-1", users[0].ID, lessons[1].ID, 2, 80),
		dashboardProgress("progress-3", "tenant-1", users[1].ID, lessons[2].ID, 1, 60),
		dashboardProgress("progress-old", "tenant-1", users[1].ID, lessons[3].ID, 0, 0),
		dashboardProgress("foreign-progress", "tenant-2", users[3].ID, lessons[4].ID, 2, 999),
	}
	for _, progress := range progressItems {
		mustCreateDashboard(t, database, progress)
	}
	for _, progress := range progressItems[:3] {
		setDashboardTime(t, database, progress, "updated_at", dayStart.Add(time.Hour))
	}
	setDashboardTime(t, database, progressItems[3], "updated_at", dayStart.Add(-time.Hour))
	setDashboardTime(t, database, progressItems[4], "updated_at", dayStart.Add(time.Hour))
}

func dashboardLesson(id, tenantID, chapterID string) *domain.CourseLesson {
	return &domain.CourseLesson{
		BaseModel: domain.BaseModel{ID: id, TenantID: tenantID},
		ChapterID: chapterID, Title: id, ContentType: "video",
	}
}

func dashboardEnrollment(
	id, tenantID, courseID, userID string,
) *domain.CourseEnrollment {
	return &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{ID: id, TenantID: tenantID},
		CourseID:  courseID, UserID: userID, Status: 1,
	}
}

func dashboardProgress(
	id, tenantID, userID, lessonID string, status, seconds int,
) *domain.LessonProgress {
	return &domain.LessonProgress{
		BaseModel: domain.BaseModel{ID: id, TenantID: tenantID},
		UserID:    userID, LessonID: lessonID, Status: status,
		LastPositionSeconds: seconds,
	}
}

func mustCreateDashboard(t *testing.T, database *gorm.DB, value interface{}) {
	t.Helper()
	if err := database.Create(value).Error; err != nil {
		t.Fatalf("create %T: %v", value, err)
	}
}

func setDashboardTime(
	t *testing.T, database *gorm.DB, value interface{}, column string, at time.Time,
) {
	t.Helper()
	if err := database.Model(value).UpdateColumn(column, at).Error; err != nil {
		t.Fatalf("set %s on %T: %v", column, value, err)
	}
}
