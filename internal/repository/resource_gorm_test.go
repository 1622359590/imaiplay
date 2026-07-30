package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestResourceRepositoryPlatformIsolationAndReferences(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewResourceRepository(database)
	platform := &domain.Resource{
		Name: "cover.png", ResourceType: "image",
		URL: "/uploads/platform/images/cover.png", CreatedBy: "root",
	}
	tenantResource := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: "tenant-a"},
		Name:      "private.mp4", ResourceType: "video",
		URL: "/uploads/tenant-a/private.mp4", CreatedBy: "admin-a",
	}
	for _, resource := range []*domain.Resource{platform, tenantResource} {
		if err := repo.Create(context.Background(), resource); err != nil {
			t.Fatalf("Create(%s) error = %v", resource.Name, err)
		}
	}

	items, total, err := repo.FindPlatform(context.Background(), 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != platform.ID {
		t.Fatalf("FindPlatform() = %#v, %d, %v", items, total, err)
	}
	if _, err := repo.FindPlatformByID(
		context.Background(), tenantResource.ID,
	); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("FindPlatformByID(tenant resource) error = %v", err)
	}

	course := &domain.Course{
		Title: "Official", CoverImage: "/api/v1/platform-covers/" + platform.ID,
		Status: 1, CreatedBy: "root", IsOfficial: true,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	referenced, err := repo.IsPlatformReferenced(
		context.Background(), platform.ID,
		[]string{course.CoverImage, platform.URL},
	)
	if err != nil || !referenced {
		t.Fatalf("IsPlatformReferenced(cover) = %v, %v", referenced, err)
	}
	if err := repo.DeletePlatform(context.Background(), tenantResource.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("DeletePlatform(tenant resource) error = %v", err)
	}
}

func TestResourceRepositoryOfficialCourseAccess(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewResourceRepository(database)
	resource := &domain.Resource{
		Name: "lesson.mp4", ResourceType: "video",
		URL: "/uploads/platform/videos/lesson.mp4", CreatedBy: "root",
	}
	course := &domain.Course{
		Title: "Official", Status: 1, CreatedBy: "root", IsOfficial: true,
	}
	if err := database.Create(resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{CourseID: course.ID, Title: "Chapter"}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "video",
		ResourceID: &resource.ID,
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	if err := database.Create(&domain.TenantOfficialCourse{
		TenantID: "tenant-enabled", CourseID: course.ID, Enabled: true,
	}).Error; err != nil {
		t.Fatalf("enable official course: %v", err)
	}
	if err := database.Create(&domain.TenantOfficialCourse{
		TenantID: "tenant-disabled", CourseID: course.ID, Enabled: false,
	}).Error; err != nil {
		t.Fatalf("disable official course: %v", err)
	}
	enrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-enabled"},
		CourseID:  course.ID, UserID: "learner-enrolled", Status: 1,
	}
	if err := database.Create(enrollment).Error; err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	tests := []struct {
		name, tenantID, userID, role string
		want                         bool
	}{
		{"superadmin", "", "root", "superadmin", true},
		{"tenant admin enabled", "tenant-enabled", "admin", "tenant_admin", true},
		{"instructor enabled", "tenant-enabled", "teacher", "instructor", true},
		{"learner enrolled", "tenant-enabled", "learner-enrolled", "learner", true},
		{"learner not enrolled", "tenant-enabled", "learner-other", "learner", false},
		{"tenant disabled", "tenant-disabled", "admin", "tenant_admin", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allowed, err := repo.CanAccessPlatformResource(
				context.Background(), resource.ID,
				test.tenantID, test.userID, test.role,
			)
			if err != nil || allowed != test.want {
				t.Fatalf("CanAccessPlatformResource() = %v, %v, want %v", allowed, err, test.want)
			}
		})
	}
	allowed, err := repo.CanAccessPlatformResource(
		context.Background(), "missing-resource",
		"tenant-enabled", "admin", "tenant_admin",
	)
	if err != nil || allowed {
		t.Fatalf("CanAccessPlatformResource(missing) = %v, %v", allowed, err)
	}
}
