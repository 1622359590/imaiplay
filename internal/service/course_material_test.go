package service

import (
	"context"
	"io"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestCourseMaterialServiceOpenForLearnerPreservesInternalOpenFailure(t *testing.T) {
	service := &CourseMaterialService{
		opener: failingMaterialOpener{},
		learnerAccess: learnerMaterialAccessStub{
			material: &domain.CourseMaterial{ResourceID: "resource-1"},
			course:   &domain.Course{BaseModel: domain.BaseModel{ID: "course-1"}},
		},
	}
	if _, _, _, err := service.OpenForLearner(
		courseContext("learner-1", "tenant-1", "learner"), "material-1",
	); errorCode(err) != 50000 {
		t.Fatalf("OpenForLearner(internal opener failure) error = %#v", err)
	}
}

type failingMaterialOpener struct{}

func (failingMaterialOpener) Open(context.Context, string) (io.ReadCloser, string, string, error) {
	return nil, "", "", errorsx.Internal("database unavailable")
}

type learnerMaterialAccessStub struct {
	material *domain.CourseMaterial
	course   *domain.Course
}

func (stub learnerMaterialAccessStub) AuthorizeMaterial(
	context.Context, string,
) (*domain.CourseMaterial, *domain.Course, error) {
	return stub.material, stub.course, nil
}

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

func TestCourseMaterialServiceInstructorCanOnlyListOwnedCourseMaterials(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courses := repository.NewCourseRepository(database)
	materials := repository.NewCourseMaterialRepository(database)
	resources := repository.NewResourceRepository(database)
	materialService := NewCourseMaterialService(courses, materials, resources, nil)
	owner := courseContext("owner", "tenant-1", "instructor")
	ownedCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "owned-course", TenantID: "tenant-1"}, Title: "Owned", CreatedBy: "owner"}
	foreignCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "foreign-course", TenantID: "tenant-1"}, Title: "Foreign", CreatedBy: "other"}
	resource := &domain.Resource{BaseModel: domain.BaseModel{ID: "attachment", TenantID: "tenant-1"}, Name: "guide.pdf", ResourceType: "attachment", URL: "/uploads/guide.pdf", CreatedBy: "owner"}
	material := &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material", TenantID: "tenant-1"}, CourseID: ownedCourse.ID, ResourceID: resource.ID, DisplayName: resource.Name, CreatedBy: "owner"}
	for _, value := range []any{ownedCourse, foreignCourse, resource, material} {
		if err := database.Create(value).Error; err != nil {
			t.Fatalf("create fixture: %v", err)
		}
	}
	items, err := materialService.ListForManager(owner, ownedCourse.ID)
	if err != nil || len(items) != 1 || items[0].ID != material.ID {
		t.Fatalf("ListForManager(owned) = %#v, %v", items, err)
	}
	if _, err := materialService.ListForManager(owner, foreignCourse.ID); errorCode(err) != 40300 {
		t.Fatalf("ListForManager(foreign) error = %#v", err)
	}
	input := CourseMaterialInput{ResourceID: resource.ID, DisplayName: "changed.pdf"}
	for name, call := range map[string]func() error{
		"add":    func() error { _, err := materialService.Add(owner, ownedCourse.ID, input); return err },
		"update": func() error { _, err := materialService.Update(owner, ownedCourse.ID, material.ID, input); return err },
		"remove": func() error { return materialService.Remove(owner, ownedCourse.ID, material.ID) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); errorCode(err) != 40300 {
				t.Fatalf("error = %#v, want 40300", err)
			}
		})
	}
}
