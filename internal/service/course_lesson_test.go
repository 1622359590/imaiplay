package service

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
)

func TestCourseLessonServiceCRUDValidationAndInstructorOwnership(t *testing.T) {
	fixture := newCourseFixture(t)
	author := courseContext("author-1", "tenant-1", "instructor")
	other := courseContext("author-2", "tenant-1", "instructor")
	course, err := fixture.courses.Create(author, "Course", "", "")
	if err != nil {
		t.Fatalf("Create(course) error = %v", err)
	}
	chapter, err := fixture.chapters.Create(author, course.ID, "Chapter", 1)
	if err != nil {
		t.Fatalf("Create(chapter) error = %v", err)
	}
	if _, err := fixture.lessons.Create(
		author, chapter.ID, "Invalid", "archive", "", 0, 1,
	); errorCode(err) != 40000 {
		t.Fatalf("Create(invalid type) error = %#v", err)
	}
	lesson, err := fixture.lessons.Create(
		author, chapter.ID, "Lesson", "video", "/video.mp4", 90, 1,
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	resource := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "resource.mp4", ResourceType: "video",
		URL: "/uploads/tenant-1/resource.mp4", CreatedBy: "author-1",
	}
	if err := fixture.resources.Create(context.Background(), resource); err != nil {
		t.Fatalf("Create(resource) error = %v", err)
	}
	resourceID := resource.ID
	resourceLesson, err := fixture.lessons.CreateWithResource(
		author, chapter.ID, "Resource lesson", "video", resourceID, "", 120, 2,
	)
	if err != nil || resourceLesson.ResourceID == nil || *resourceLesson.ResourceID != resourceID {
		t.Fatalf("CreateWithResource() = %#v, %v", resourceLesson, err)
	}
	items, err := fixture.lessons.List(author, chapter.ID)
	if err != nil || len(items) != 2 || items[0].ID != lesson.ID {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	if _, err := fixture.lessons.List(other, chapter.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor List() error = %#v", err)
	}
	updated, err := fixture.lessons.Update(
		author, lesson.ID, "Updated", "text", "Lesson body", 0, 2,
	)
	if err != nil || updated.Title != "Updated" || updated.ContentType != "text" {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if err := fixture.lessons.Delete(other, lesson.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor Delete() error = %#v", err)
	}
	if err := fixture.lessons.Delete(author, lesson.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestCourseLessonServiceValidatesResourceOwnership(t *testing.T) {
	fixture := newCourseFixture(t)
	tenantAdmin := courseContext("admin", "tenant-1", "tenant_admin")
	superadmin := courseContext("root", "", "superadmin")
	tenantCourse, err := fixture.courses.Create(
		tenantAdmin, "Tenant course", "", "",
	)
	if err != nil {
		t.Fatalf("Create(tenant course) error = %v", err)
	}
	tenantChapter, err := fixture.chapters.Create(
		tenantAdmin, tenantCourse.ID, "Tenant chapter", 1,
	)
	if err != nil {
		t.Fatalf("Create(tenant chapter) error = %v", err)
	}
	officialCourse, err := fixture.courses.CreateOfficial(
		superadmin, "Official course", "", "", 1,
	)
	if err != nil {
		t.Fatalf("CreateOfficial() error = %v", err)
	}
	officialChapter, err := fixture.chapters.Create(
		superadmin, officialCourse.ID, "Official chapter", 1,
	)
	if err != nil {
		t.Fatalf("Create(official chapter) error = %v", err)
	}
	tenantResource := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "tenant.mp4", ResourceType: "video",
		URL: "/uploads/tenant-1/tenant.mp4", CreatedBy: "admin",
	}
	platformResource := &domain.Resource{
		Name: "official.mp4", ResourceType: "video",
		URL: "/uploads/platform/videos/official.mp4", CreatedBy: "root",
	}
	for _, resource := range []*domain.Resource{
		tenantResource, platformResource,
	} {
		if err := fixture.resources.Create(
			context.Background(), resource,
		); err != nil {
			t.Fatalf("Create(resource) error = %v", err)
		}
	}

	if _, err := fixture.lessons.CreateWithResource(
		tenantAdmin, tenantChapter.ID, "Wrong", "video",
		platformResource.ID, "", 60, 1,
	); errorCode(err) != 40000 {
		t.Fatalf("tenant lesson with platform resource error = %#v", err)
	}
	if _, err := fixture.lessons.CreateWithResource(
		superadmin, officialChapter.ID, "Wrong", "video",
		tenantResource.ID, "", 60, 1,
	); errorCode(err) != 40000 {
		t.Fatalf("official lesson with tenant resource error = %#v", err)
	}
	if _, err := fixture.lessons.CreateWithResource(
		superadmin, officialChapter.ID, "Correct", "video",
		platformResource.ID, "", 60, 1,
	); err != nil {
		t.Fatalf("official lesson with platform resource error = %v", err)
	}
	if _, err := fixture.lessons.CreateWithResource(
		superadmin, officialChapter.ID, "Text", "text",
		platformResource.ID, "body", 0, 2,
	); errorCode(err) != 40000 {
		t.Fatalf("text lesson with resource error = %#v", err)
	}
}
