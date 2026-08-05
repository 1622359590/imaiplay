package repository

import (
	"context"
	"errors"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestCourseMaterialRepositoryOrdersAndScopesTenantRows(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseMaterialRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	for _, resource := range []*domain.Resource{
		{BaseModel: domain.BaseModel{ID: "resource-1", TenantID: "tenant-1"}, Name: "one", ResourceType: "attachment", URL: "one.pdf", CreatedBy: "admin-1"},
		{BaseModel: domain.BaseModel{ID: "resource-2", TenantID: "tenant-1"}, Name: "two", ResourceType: "attachment", URL: "two.pdf", CreatedBy: "admin-1"},
		{BaseModel: domain.BaseModel{ID: "resource-3", TenantID: "tenant-2"}, Name: "three", ResourceType: "attachment", URL: "three.pdf", CreatedBy: "admin-2"},
		{BaseModel: domain.BaseModel{ID: "resource-4", TenantID: "tenant-1"}, Name: "four", ResourceType: "attachment", URL: "four.pdf", CreatedBy: "admin-1"},
	} {
		if err := database.Create(resource).Error; err != nil {
			t.Fatalf("create resource %s: %v", resource.ID, err)
		}
	}

	first := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: "course-1", ResourceID: "resource-1", DisplayName: "第二份.pdf", SortOrder: 2, CreatedBy: "admin-1"}
	second := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: "course-1", ResourceID: "resource-2", DisplayName: "第一份.pdf", SortOrder: 1, CreatedBy: "admin-1"}
	foreign := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: "tenant-2"}, CourseID: "course-1", ResourceID: "resource-3", DisplayName: "外部.pdf", CreatedBy: "admin-2"}
	for _, material := range []*domain.CourseMaterial{first, second, foreign} {
		if err := repo.Create(base, material); err != nil {
			t.Fatalf("Create(%s) error = %v", material.DisplayName, err)
		}
	}

	items, err := repo.FindByCourse(tenantOne, "course-1")
	if err != nil {
		t.Fatalf("FindByCourse() error = %v", err)
	}
	if len(items) != 2 || items[0].ID != second.ID || items[1].ID != first.ID {
		t.Fatalf("FindByCourse() order = %#v", items)
	}
	if _, err := repo.FindByID(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	duplicate := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: "course-1", ResourceID: "resource-1", DisplayName: "重复.pdf", CreatedBy: "admin-1"}
	if err := repo.Create(base, duplicate); err == nil {
		t.Fatal("duplicate course resource association error = nil")
	}

	first.DisplayName, first.SortOrder, first.ResourceID = "更新.pdf", 3, "resource-4"
	if err := repo.Update(tenantOne, first); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	updated, err := repo.FindByID(tenantOne, first.ID)
	if err != nil || updated.DisplayName != "更新.pdf" || updated.ResourceID != "resource-4" || updated.SortOrder != 3 {
		t.Fatalf("FindByID(updated) = %#v, %v", updated, err)
	}
	if err := repo.Delete(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(tenantOne, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestCourseMaterialRepositoryScopesOfficialRows(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	course := &domain.Course{BaseModel: domain.BaseModel{ID: "official", TenantID: ""}, Title: "官方课", Status: 1, CreatedBy: "root", IsOfficial: true}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	if err := database.Create(&domain.TenantOfficialCourse{TenantID: "tenant-1", CourseID: course.ID, Enabled: true}).Error; err != nil {
		t.Fatalf("enable official course: %v", err)
	}
	if err := database.Create(&domain.Resource{BaseModel: domain.BaseModel{ID: "platform-resource", TenantID: ""}, Name: "official", ResourceType: "attachment", URL: "official.pdf", CreatedBy: "root"}).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	material := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: ""}, CourseID: course.ID, ResourceID: "platform-resource", DisplayName: "官方手册.pdf", CreatedBy: "root"}
	if err := database.Create(material).Error; err != nil {
		t.Fatalf("create material: %v", err)
	}
	repo := NewCourseMaterialRepository(database)
	enabled := usercontext.WithUser(context.Background(), "learner-1", "tenant-1", "", "learner")
	disabled := usercontext.WithUser(context.Background(), "learner-2", "tenant-2", "", "learner")
	superadmin := usercontext.WithUser(context.Background(), "root", "", "", "superadmin")
	if items, err := repo.FindByCourse(enabled, course.ID); err != nil || len(items) != 1 {
		t.Fatalf("enabled FindByCourse() = %#v, %v", items, err)
	}
	if items, err := repo.FindByCourse(disabled, course.ID); err != nil || len(items) != 0 {
		t.Fatalf("disabled FindByCourse() = %#v, %v", items, err)
	}
	if _, err := repo.FindByID(superadmin, material.ID); err != nil {
		t.Fatalf("superadmin FindByID() error = %v", err)
	}
	material.DisplayName = "不可由学院修改.pdf"
	if err := repo.Update(enabled, material); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("tenant update official error = %v", err)
	}
}

func TestCourseMaterialRepositoryInstructorCannotMutateMaterials(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	resource := &domain.Resource{BaseModel: domain.BaseModel{ID: "resource", TenantID: "tenant-1"}, Name: "guide.pdf", ResourceType: "attachment", URL: "guide.pdf", CreatedBy: "owner"}
	material := &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material", TenantID: "tenant-1"}, CourseID: "course", ResourceID: resource.ID, DisplayName: resource.Name, CreatedBy: "owner"}
	if err := database.Create(resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if err := database.Create(material).Error; err != nil {
		t.Fatalf("create material: %v", err)
	}
	repo := NewCourseMaterialRepository(database)
	instructor := usercontext.WithUser(context.Background(), "owner", "tenant-1", "", "instructor")
	material.DisplayName = "changed.pdf"
	if err := repo.Update(instructor, material); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Update() error = %v", err)
	}
	if err := repo.Delete(instructor, material.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Delete() error = %v", err)
	}
}
