package service

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestCourseMaterialServiceEnforcesOwnershipAndManagesAssociations(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courses := repository.NewCourseRepository(database)
	materials := repository.NewCourseMaterialRepository(database)
	resources := repository.NewResourceRepository(database)
	service := NewCourseMaterialService(courses, materials, resources, nil)

	tenantCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "tenant-course", TenantID: "tenant-1"}, Title: "Tenant", Status: 1, CreatedBy: "admin"}
	foreignCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "foreign-course", TenantID: "tenant-2"}, Title: "Foreign", Status: 1, CreatedBy: "other"}
	officialCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "official-course", TenantID: ""}, Title: "Official", Status: 1, CreatedBy: "root", IsOfficial: true}
	for _, course := range []*domain.Course{tenantCourse, foreignCourse, officialCourse} {
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", course.ID, err)
		}
	}
	tenantResource := &domain.Resource{BaseModel: domain.BaseModel{ID: "tenant-resource", TenantID: "tenant-1"}, Name: "guide.pdf", ResourceType: "attachment", URL: "/uploads/tenant-1/guide.pdf", SizeBytes: 100, CreatedBy: "admin"}
	replacement := &domain.Resource{BaseModel: domain.BaseModel{ID: "replacement", TenantID: "tenant-1"}, Name: "new.zip", ResourceType: "attachment", URL: "/uploads/tenant-1/new.zip", SizeBytes: 200, CreatedBy: "admin"}
	platformResource := &domain.Resource{BaseModel: domain.BaseModel{ID: "platform-resource", TenantID: ""}, Name: "official.pdf", ResourceType: "attachment", URL: "/uploads/platform/attachments/official.pdf", SizeBytes: 300, CreatedBy: "root"}
	for _, resource := range []*domain.Resource{tenantResource, replacement, platformResource} {
		if err := database.Create(resource).Error; err != nil {
			t.Fatalf("create resource %s: %v", resource.ID, err)
		}
	}

	tenantAdmin := courseContext("admin", "tenant-1", "tenant_admin")
	created, err := service.Add(tenantAdmin, tenantCourse.ID, CourseMaterialInput{ResourceID: tenantResource.ID, DisplayName: "  入门手册.pdf  ", SortOrder: 2})
	if err != nil || created.DisplayName != "入门手册.pdf" || created.TenantID != "tenant-1" {
		t.Fatalf("Add() = %#v, %v", created, err)
	}
	if _, err := service.Add(tenantAdmin, tenantCourse.ID, CourseMaterialInput{ResourceID: tenantResource.ID, DisplayName: "重复.pdf"}); errorCode(err) != 40900 {
		t.Fatalf("Add(duplicate) error = %#v", err)
	}
	updated, err := service.Update(tenantAdmin, tenantCourse.ID, created.ID, CourseMaterialInput{ResourceID: replacement.ID, DisplayName: "  新资料.zip ", SortOrder: 1})
	if err != nil || updated.DisplayName != "新资料.zip" || updated.ResourceID != replacement.ID || updated.SortOrder != 1 {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if _, err := service.Add(tenantAdmin, foreignCourse.ID, CourseMaterialInput{ResourceID: tenantResource.ID, DisplayName: "跨租户.pdf"}); errorCode(err) != 40400 {
		t.Fatalf("Add(foreign) error = %#v", err)
	}
	if _, err := service.Add(courseContext("learner", "tenant-1", "learner"), tenantCourse.ID, CourseMaterialInput{ResourceID: tenantResource.ID, DisplayName: "无权限.pdf"}); errorCode(err) != 40300 {
		t.Fatalf("Add(learner) error = %#v", err)
	}
	if _, err := service.Add(tenantAdmin, officialCourse.ID, CourseMaterialInput{ResourceID: tenantResource.ID, DisplayName: "官方.pdf"}); errorCode(err) != 40400 && errorCode(err) != 40300 {
		t.Fatalf("Add(official as tenant) error = %#v", err)
	}

	root := courseContext("root", "", "superadmin")
	official, err := service.Add(root, officialCourse.ID, CourseMaterialInput{ResourceID: platformResource.ID, DisplayName: "官方手册.pdf"})
	if err != nil || official.TenantID != "" || official.CourseID != officialCourse.ID {
		t.Fatalf("Add(official) = %#v, %v", official, err)
	}
	if err := service.Remove(tenantAdmin, tenantCourse.ID, created.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := resources.FindByID(tenantAdmin, replacement.ID); err != nil {
		t.Fatalf("Remove() deleted resource: %v", err)
	}
}

func TestCourseMaterialServiceRejectsInvalidResourceAndDisplayName(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courses := repository.NewCourseRepository(database)
	materials := repository.NewCourseMaterialRepository(database)
	resources := repository.NewResourceRepository(database)
	service := NewCourseMaterialService(courses, materials, resources, nil)
	ctx := courseContext("admin", "tenant-1", "tenant_admin")
	if err := database.Create(&domain.Course{BaseModel: domain.BaseModel{ID: "course", TenantID: "tenant-1"}, Title: "Course", CreatedBy: "admin"}).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	video := &domain.Resource{BaseModel: domain.BaseModel{ID: "video", TenantID: "tenant-1"}, Name: "video.mp4", ResourceType: "video", URL: "/uploads/video.mp4", CreatedBy: "admin"}
	if err := database.Create(video).Error; err != nil {
		t.Fatalf("create video: %v", err)
	}
	if _, err := service.Add(ctx, "course", CourseMaterialInput{ResourceID: video.ID, DisplayName: "video.mp4"}); errorCode(err) != 40000 {
		t.Fatalf("Add(video) error = %#v", err)
	}
	if _, err := service.Add(ctx, "course", CourseMaterialInput{ResourceID: "missing", DisplayName: "missing.pdf"}); errorCode(err) != 40400 {
		t.Fatalf("Add(missing resource) error = %#v", err)
	}
	if _, err := service.Add(ctx, "course", CourseMaterialInput{ResourceID: video.ID, DisplayName: "  "}); errorCode(err) != 40000 {
		t.Fatalf("Add(blank name) error = %#v", err)
	}
}
