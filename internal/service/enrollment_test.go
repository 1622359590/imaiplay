package service

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

func TestEnrollmentServiceTenantAdminFlow(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")

	if _, err := fixture.enrollments.Enroll(
		learner, fixture.course.ID, fixture.learner.ID,
	); errorCode(err) != 40300 {
		t.Fatalf("Enroll(learner role) error = %#v", err)
	}
	enrollment, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID,
	)
	if err != nil || enrollment.TenantID != fixture.tenant.ID ||
		enrollment.Status != 1 {
		t.Fatalf("Enroll() = %#v, %v", enrollment, err)
	}
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID,
	); errorCode(err) != 40900 {
		t.Fatalf("Enroll(duplicate) error = %#v", err)
	}
	items, err := fixture.enrollments.ListByCourse(admin, fixture.course.ID)
	if err != nil || len(items) != 1 || items[0].ID != enrollment.ID {
		t.Fatalf("ListByCourse() = %#v, %v", items, err)
	}
	if err := fixture.enrollments.Remove(admin, enrollment.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
}

func TestEnrollmentServiceOnlyAcceptsLearnerInSameTenant(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	instructor := &domain.User{
		BaseModel: domain.BaseModel{TenantID: fixture.tenant.ID},
		Email:     "teacher@example.com", Password: "hash", Name: "Teacher",
		Role: "instructor", Status: 1,
	}
	if err := fixture.users.Create(context.Background(), instructor); err != nil {
		t.Fatalf("create instructor: %v", err)
	}
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, instructor.ID,
	); errorCode(err) != 40000 {
		t.Fatalf("Enroll(instructor) error = %#v", err)
	}
	foreignAdmin := courseContext("foreign-admin", "tenant-2", "tenant_admin")
	if _, err := fixture.enrollments.Enroll(
		foreignAdmin, fixture.course.ID, fixture.learner.ID,
	); errorCode(err) != 40400 {
		t.Fatalf("Enroll(cross tenant) error = %#v", err)
	}
}

type learningFixture struct {
	database       *gorm.DB
	tenant         *domain.Tenant
	admin          *domain.User
	learner        *domain.User
	course         *domain.Course
	chapter        *domain.CourseChapter
	lesson         *domain.CourseLesson
	users          repository.UserRepository
	enrollmentRepo repository.CourseEnrollmentRepository
	enrollments    *EnrollmentService
	progress       *ProgressService
}

func newLearningFixture(t *testing.T) learningFixture {
	t.Helper()
	database, tenantRepo, userRepo := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "learning", Name: "Learning", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	admin := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     "admin@example.com", Password: "hash", Name: "Admin",
		Role: "tenant_admin", Status: 1,
	}
	learner := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     "learner@example.com", Password: "hash", Name: "Learner",
		Role: "learner", Status: 1,
	}
	for _, user := range []*domain.User{admin, learner} {
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	progressRepo := repository.NewLessonProgressRepository(database)
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Title:     "Course", Status: 1, CreatedBy: admin.ID,
	}
	if err := courseRepo.Create(context.Background(), course); err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		CourseID:  course.ID, Title: "Chapter",
	}
	if err := chapterRepo.Create(context.Background(), chapter); err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "video",
		ContentURL: "/uploads/video.mp4", DurationSeconds: 100,
	}
	if err := lessonRepo.Create(context.Background(), lesson); err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	return learningFixture{
		database: database,
		tenant:   tenant, admin: admin, learner: learner,
		course: course, chapter: chapter, lesson: lesson, users: userRepo,
		enrollmentRepo: enrollmentRepo,
		enrollments:    NewEnrollmentService(enrollmentRepo, courseRepo, userRepo),
		progress: NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
		),
	}
}
