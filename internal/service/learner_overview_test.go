package service

import (
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

func TestLearnerOverviewServiceBuildsFixedOverviewAndShanghaiToday(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	ctx := courseContext("learner-1", "tenant-1", "learner")
	alpha := &domain.CourseCategory{
		BaseModel: domain.BaseModel{ID: "category-alpha", TenantID: "tenant-1"},
		Name:      "Alpha", NormalizedName: "alpha", Status: 1,
	}
	zeta := &domain.CourseCategory{
		BaseModel: domain.BaseModel{ID: "category-zeta", TenantID: "tenant-1"},
		Name:      "Zeta", NormalizedName: "zeta", Status: 1,
	}
	for _, category := range []*domain.CourseCategory{zeta, alpha} {
		if err := database.Create(category).Error; err != nil {
			t.Fatalf("create category: %v", err)
		}
	}
	partial := serviceOverviewCourse(t, database, "course-partial", "A Partial", &zeta.ID, domain.AssignmentOptional)
	partialLessons := []*domain.CourseLesson{
		serviceOverviewLesson(t, database, partial, "partial-1", 60),
		serviceOverviewLesson(t, database, partial, "partial-2", 90),
		serviceOverviewLesson(t, database, partial, "partial-3", 120),
	}
	completed := serviceOverviewCourse(t, database, "course-completed", "B Completed", &alpha.ID, domain.AssignmentRequired)
	completedLesson := serviceOverviewLesson(t, database, completed, "completed", 30)
	zero := serviceOverviewCourse(t, database, "course-zero", "C Zero", nil, domain.AssignmentRequired)
	recent := serviceOverviewCourse(t, database, "course-recent", "D Recent", nil, domain.AssignmentOptional)
	recentLesson := serviceOverviewLesson(t, database, recent, "recent", 45)

	base := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	for _, progress := range []*domain.LessonProgress{
		{BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: base.Add(5 * time.Hour)}, UserID: "learner-1", LessonID: partialLessons[0].ID, ProgressPercent: 100, Status: 2, LastPositionSeconds: 55},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: base.Add(4 * time.Hour)}, UserID: "learner-1", LessonID: partialLessons[1].ID, ProgressPercent: 50, Status: 1, LastPositionSeconds: 20},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: base.Add(3 * time.Hour)}, UserID: "learner-1", LessonID: completedLesson.ID, ProgressPercent: 100, Status: 2, LastPositionSeconds: 30},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1", UpdatedAt: base.Add(2 * time.Hour)}, UserID: "learner-1", LessonID: recentLesson.ID, ProgressPercent: 1, Status: 1, LastPositionSeconds: 1},
	} {
		if err := database.Create(progress).Error; err != nil {
			t.Fatalf("create progress: %v", err)
		}
	}
	for _, stat := range []*domain.LearningDailyStat{
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "learner-1", StudyDate: "2026-08-05", DurationSeconds: 45},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "learner-1", StudyDate: "2026-08-04", DurationSeconds: 30},
	} {
		if err := database.Create(stat).Error; err != nil {
			t.Fatalf("create daily stat: %v", err)
		}
	}

	service := NewLearnerOverviewService(repository.NewLearnerOverviewRepository(database))
	service.now = func() time.Time {
		return time.Date(2026, 8, 4, 16, 30, 0, 0, time.UTC)
	}
	overview, err := service.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if overview.RequiredCompleted != 1 || overview.RequiredTotal != 2 ||
		overview.TodayLearningSeconds != 45 || overview.TotalLearningSeconds != 75 {
		t.Fatalf("overview totals = %#v", overview)
	}
	if len(overview.Categories) != 2 || overview.Categories[0].ID != alpha.ID || overview.Categories[1].ID != zeta.ID {
		t.Fatalf("categories = %#v", overview.Categories)
	}
	if len(overview.Courses) != 4 || overview.Courses[0].Course.ID != partial.ID || overview.Courses[2].Course.ID != zero.ID {
		t.Fatalf("courses = %#v", overview.Courses)
	}
	partialGot := overview.Courses[0]
	if partialGot.ProgressPercent != 50 || partialGot.LessonCount != 3 || partialGot.CompletedLessonCount != 1 ||
		partialGot.RecentLesson == nil || partialGot.RecentLesson.ID != partialLessons[0].ID ||
		partialGot.RecentLesson.LastPositionSeconds != 55 {
		t.Fatalf("partial course = %#v", partialGot)
	}
	zeroGot := overview.Courses[2]
	if zeroGot.ProgressPercent != 0 || zeroGot.LessonCount != 0 || zeroGot.CompletedLessonCount != 0 {
		t.Fatalf("zero lesson course = %#v", zeroGot)
	}

	items, total, err := service.GetRecent(ctx, 1, 1)
	if err != nil || total != 3 || len(items) != 1 {
		t.Fatalf("GetRecent(1,1) = %#v, %d, %v", items, total, err)
	}
	if items[0].Course.ID != completed.ID || items[0].RecentLesson.ID != completedLesson.ID ||
		items[0].ProgressPercent != 100 || items[0].LastPositionSeconds != 30 ||
		!items[0].LastLearnedAt.Equal(base.Add(3*time.Hour)) {
		t.Fatalf("deduplicated recent page = %#v", items[0])
	}
}

func TestLearnerOverviewServiceRejectsNonLearnerContext(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	service := NewLearnerOverviewService(repository.NewLearnerOverviewRepository(database))
	admin := courseContext("admin-1", "tenant-1", "tenant_admin")
	if _, err := service.Get(admin); errorCode(err) != 40300 {
		t.Fatalf("Get(admin) error = %#v", err)
	}
	if _, _, err := service.GetRecent(admin, 0, 20); errorCode(err) != 40300 {
		t.Fatalf("GetRecent(admin) error = %#v", err)
	}
}

func serviceOverviewCourse(
	t *testing.T,
	database *gorm.DB,
	id, title string,
	categoryID *string,
	assignmentType string,
) *domain.Course {
	t.Helper()
	course := &domain.Course{
		BaseModel: domain.BaseModel{ID: id, TenantID: "tenant-1"},
		Title:     title, Description: title + " description", CoverImage: title + ".png",
		Status: 1, CreatedBy: "admin", CategoryID: categoryID,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	if err := database.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  course.ID, UserID: "learner-1", Status: 1,
		AssignmentType: assignmentType,
	}).Error; err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	return course
}

func serviceOverviewLesson(t *testing.T, database *gorm.DB, course *domain.Course, id string, duration int) *domain.CourseLesson {
	t.Helper()
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{ID: "chapter-" + id, TenantID: course.TenantID},
		CourseID:  course.ID, Title: id + " chapter",
	}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{ID: "lesson-" + id, TenantID: course.TenantID},
		ChapterID: chapter.ID, Title: id, ContentType: "video", DurationSeconds: duration,
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	return lesson
}
