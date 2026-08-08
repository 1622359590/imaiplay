package repository

import (
	"context"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

func TestLearnerOverviewRepositoryAggregatesOnlyActiveVisibleAssignments(t *testing.T) {
	database := learningTimeDatabase(t)
	category := &domain.CourseCategory{
		BaseModel: domain.BaseModel{ID: "category-tenant", TenantID: "tenant-1"},
		Name:      "Tenant Category", NormalizedName: "tenant category", Status: 1,
	}
	platformCategory := &domain.CourseCategory{
		BaseModel: domain.BaseModel{ID: "category-platform"},
		Name:      "Platform Category", NormalizedName: "platform category", Status: 1,
	}
	for _, item := range []*domain.CourseCategory{category, platformCategory} {
		if err := database.Create(item).Error; err != nil {
			t.Fatalf("create category: %v", err)
		}
	}

	assigned := overviewCourse(t, database, "assigned", "tenant-1", 1, false, &category.ID)
	if err := database.Model(assigned).Update("course_type", domain.CourseTypeOptional).Error; err != nil {
		t.Fatalf("set assigned course type: %v", err)
	}
	assigned.CourseType = domain.CourseTypeOptional
	assignedLessons := []*domain.CourseLesson{
		overviewLesson(t, database, assigned, "assigned-a", 80),
		overviewLesson(t, database, assigned, "assigned-b", 120),
	}
	completed := overviewCourse(t, database, "completed", "tenant-1", 1, false, nil)
	completedLesson := overviewLesson(t, database, completed, "completed-lesson", 60)
	zeroLesson := overviewCourse(t, database, "zero", "tenant-1", 1, false, nil)
	draft := overviewCourse(t, database, "draft", "tenant-1", 0, false, nil)
	inactive := overviewCourse(t, database, "inactive", "tenant-1", 1, false, nil)
	foreign := overviewCourse(t, database, "foreign", "tenant-2", 1, false, nil)
	official := overviewCourse(t, database, "official", "", 1, true, &platformCategory.ID)
	officialLesson := overviewLesson(t, database, official, "official-lesson", 30)
	disabledOfficial := overviewCourse(t, database, "official-disabled", "", 1, true, nil)
	unactivatedOfficial := overviewCourse(t, database, "official-unactivated", "", 1, true, nil)
	draftOfficial := overviewCourse(t, database, "official-draft", "", 0, true, nil)

	activeEnrollment := func(course *domain.Course, assignment string) *domain.CourseEnrollment {
		return &domain.CourseEnrollment{
			BaseModel: domain.BaseModel{TenantID: "tenant-1"},
			CourseID:  course.ID, UserID: "learner-1", Status: 1,
			AssignmentType: assignment,
		}
	}
	inactiveEnrollment := activeEnrollment(inactive, domain.AssignmentRequired)
	enrollments := []*domain.CourseEnrollment{
		activeEnrollment(assigned, domain.AssignmentRequired),
		activeEnrollment(completed, domain.AssignmentRequired),
		activeEnrollment(zeroLesson, domain.AssignmentOptional),
		activeEnrollment(draft, domain.AssignmentRequired),
		inactiveEnrollment,
		activeEnrollment(foreign, domain.AssignmentRequired),
		activeEnrollment(official, domain.AssignmentOptional),
		activeEnrollment(disabledOfficial, domain.AssignmentRequired),
		activeEnrollment(unactivatedOfficial, domain.AssignmentRequired),
		activeEnrollment(draftOfficial, domain.AssignmentRequired),
	}
	for _, enrollment := range enrollments {
		if err := database.Create(enrollment).Error; err != nil {
			t.Fatalf("create enrollment: %v", err)
		}
	}
	if err := database.Model(inactiveEnrollment).Update("status", 0).Error; err != nil {
		t.Fatalf("deactivate enrollment: %v", err)
	}
	for _, activation := range []*domain.TenantOfficialCourse{
		{TenantID: "tenant-1", CourseID: official.ID, Enabled: true},
		{TenantID: "tenant-1", CourseID: disabledOfficial.ID, Enabled: true},
		{TenantID: "tenant-1", CourseID: draftOfficial.ID, Enabled: true},
	} {
		if err := database.Create(activation).Error; err != nil {
			t.Fatalf("create official activation: %v", err)
		}
	}
	if err := database.Model(&domain.TenantOfficialCourse{}).
		Where("tenant_id = ? AND course_id = ?", "tenant-1", disabledOfficial.ID).
		Update("enabled", false).Error; err != nil {
		t.Fatalf("disable official activation: %v", err)
	}

	older := time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	for _, progress := range []*domain.LessonProgress{
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: older},
			UserID:    "learner-1", LessonID: assignedLessons[0].ID,
			ProgressPercent: 50, Status: 1, LastPositionSeconds: 30,
		},
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: newer},
			UserID:    "learner-1", LessonID: assignedLessons[1].ID,
			ProgressPercent: 100, Status: 2, LastPositionSeconds: 90,
		},
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: older.Add(-time.Hour)},
			UserID:    "learner-1", LessonID: completedLesson.ID,
			ProgressPercent: 100, Status: 2, LastPositionSeconds: 60,
		},
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: newer.Add(time.Hour)},
			UserID:    "other-learner", LessonID: officialLesson.ID,
			ProgressPercent: 100, Status: 2, LastPositionSeconds: 30,
		},
	} {
		if err := database.Create(progress).Error; err != nil {
			t.Fatalf("create progress: %v", err)
		}
	}
	for _, stat := range []*domain.LearningDailyStat{
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "learner-1", StudyDate: "2026-08-05", DurationSeconds: 15},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "learner-1", StudyDate: "2026-08-04", DurationSeconds: 20},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "other-learner", StudyDate: "2026-08-05", DurationSeconds: 800},
		{BaseModel: domain.BaseModel{TenantID: "tenant-2"}, UserID: "learner-1", StudyDate: "2026-08-05", DurationSeconds: 900},
	} {
		if err := database.Create(stat).Error; err != nil {
			t.Fatalf("create daily stat: %v", err)
		}
	}

	repo := NewLearnerOverviewRepository(database)
	got, err := repo.Get(learnerRepositoryContext("learner-1", "tenant-1"), "2026-08-05")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.TodayLearningSeconds != 15 || got.TotalLearningSeconds != 35 {
		t.Fatalf("learning seconds = today %d total %d", got.TodayLearningSeconds, got.TotalLearningSeconds)
	}
	if len(got.Courses) != 4 {
		t.Fatalf("visible courses = %#v", got.Courses)
	}
	byID := make(map[string]LearnerOverviewCourse, len(got.Courses))
	for _, course := range got.Courses {
		byID[course.Course.ID] = course
	}
	assignedGot := byID[assigned.ID]
	if assignedGot.AssignmentType != domain.CourseTypeOptional ||
		assignedGot.LessonCount != 2 || assignedGot.CompletedLessonCount != 1 ||
		assignedGot.ProgressPercent != 75 || assignedGot.RecentLesson == nil ||
		assignedGot.RecentLesson.ID != assignedLessons[1].ID ||
		assignedGot.LastPositionSeconds != 90 ||
		assignedGot.LastLearnedAt == nil || !assignedGot.LastLearnedAt.Equal(newer) ||
		assignedGot.Category == nil || assignedGot.Category.ID != category.ID {
		t.Fatalf("assigned aggregate = %#v", assignedGot)
	}
	completedGot := byID[completed.ID]
	if completedGot.LessonCount != 1 || completedGot.CompletedLessonCount != 1 || completedGot.ProgressPercent != 100 {
		t.Fatalf("completed aggregate = %#v", completedGot)
	}
	zeroGot := byID[zeroLesson.ID]
	if zeroGot.LessonCount != 0 || zeroGot.CompletedLessonCount != 0 || zeroGot.ProgressPercent != 0 {
		t.Fatalf("zero lesson aggregate = %#v", zeroGot)
	}
	officialGot := byID[official.ID]
	if officialGot.Course.ID == "" || officialGot.LessonCount != 1 || officialGot.ProgressPercent != 0 ||
		officialGot.Category == nil || officialGot.Category.ID != platformCategory.ID {
		t.Fatalf("official aggregate = %#v", officialGot)
	}
	for _, hidden := range []*domain.Course{draft, inactive, foreign, disabledOfficial, unactivatedOfficial, draftOfficial} {
		if _, ok := byID[hidden.ID]; ok {
			t.Errorf("hidden course %q leaked", hidden.Title)
		}
	}
}

func TestLearnerOverviewRepositoryRejectsNonLearnerContext(t *testing.T) {
	repo := NewLearnerOverviewRepository(learningTimeDatabase(t))
	ctx := usercontext.WithUser(context.Background(), "admin-1", "tenant-1", "", "tenant_admin")
	if _, err := repo.Get(ctx, "2026-08-05"); err == nil {
		t.Fatal("Get(admin) error = nil")
	}
}

func overviewCourse(
	t *testing.T,
	database *gorm.DB,
	title, tenantID string,
	status int,
	official bool,
	categoryID *string,
) *domain.Course {
	t.Helper()
	course := &domain.Course{
		BaseModel: domain.BaseModel{ID: "course-" + title, TenantID: tenantID},
		Title:     title, Description: title + " description", CoverImage: title + ".png",
		Status: status, CreatedBy: "admin", IsOfficial: official, CategoryID: categoryID,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course %s: %v", title, err)
	}
	return course
}

func overviewLesson(t *testing.T, database *gorm.DB, course *domain.Course, title string, duration int) *domain.CourseLesson {
	t.Helper()
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{ID: "chapter-" + title, TenantID: course.TenantID},
		CourseID:  course.ID, Title: title + " chapter",
	}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{ID: "lesson-" + title, TenantID: course.TenantID},
		ChapterID: chapter.ID, Title: title, ContentType: "video",
		DurationSeconds: duration,
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	return lesson
}
