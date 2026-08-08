package service

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestCourseServiceListsEnabledOfficialWithoutEnrollment(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courseRepo := repository.NewCourseRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	service := NewCourseService(
		courseRepo,
		repository.NewCourseChapterRepository(database),
		repository.NewCourseLessonRepository(database),
		enrollmentRepo,
		repository.NewCourseMaterialRepository(database),
	)
	official := &domain.Course{
		BaseModel: domain.BaseModel{ID: "automatic-official"},
		Title:     "Automatic Official", Status: 1, CreatedBy: "root", IsOfficial: true,
	}
	if err := database.Create(official).Error; err != nil {
		t.Fatalf("create official course: %v", err)
	}
	if err := database.Create(&domain.TenantOfficialCourse{
		TenantID: "tenant-1", CourseID: official.ID, Enabled: true,
	}).Error; err != nil {
		t.Fatalf("enable official course: %v", err)
	}
	learner := courseContext("learner-1", "tenant-1", "learner")

	assertSingleOfficial := func(stage string) {
		t.Helper()
		items, total, err := service.ListPublished(learner, 0, 20)
		if err != nil || total != 1 || len(items) != 1 || items[0].ID != official.ID {
			t.Fatalf("ListPublished(%s) = %#v, %d, %v", stage, items, total, err)
		}
	}
	assertSingleOfficial("without enrollment")

	if err := database.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  official.ID, UserID: "learner-1", Status: 1,
		AssignmentType: domain.AssignmentRequired,
	}).Error; err != nil {
		t.Fatalf("create explicit enrollment: %v", err)
	}
	assertSingleOfficial("with enrollment")
}
