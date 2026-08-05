package repository

import (
	"context"
	"errors"
	"fmt"
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
	for _, course := range []*domain.Course{
		{BaseModel: domain.BaseModel{ID: "course-1", TenantID: "tenant-1"}, Title: "One", CreatedBy: "admin-1"},
		{BaseModel: domain.BaseModel{ID: "course-2", TenantID: "tenant-2"}, Title: "Two", CreatedBy: "admin-2"},
	} {
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", course.ID, err)
		}
	}
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
	foreign := &domain.CourseMaterial{BaseModel: domain.BaseModel{TenantID: "tenant-2"}, CourseID: "course-2", ResourceID: "resource-3", DisplayName: "外部.pdf", CreatedBy: "admin-2"}
	for index, material := range []*domain.CourseMaterial{first, second, foreign} {
		ctx := tenantOne
		if index == 2 {
			ctx = tenantTwo
		}
		if err := repo.Create(ctx, material); err != nil {
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
	if err := repo.Create(tenantOne, duplicate); err == nil {
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

func TestCourseMaterialRepositoryCreateEnforcesManagerAndParentScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	for _, course := range []*domain.Course{
		{BaseModel: domain.BaseModel{ID: "tenant-course", TenantID: "tenant-1"}, Title: "Tenant", CreatedBy: "admin-1"},
		{BaseModel: domain.BaseModel{ID: "foreign-course", TenantID: "tenant-2"}, Title: "Foreign", CreatedBy: "admin-2"},
		{BaseModel: domain.BaseModel{ID: "tenant-official", TenantID: "tenant-1"}, Title: "Invalid official", CreatedBy: "root", IsOfficial: true},
		{BaseModel: domain.BaseModel{ID: "official-course"}, Title: "Official", CreatedBy: "root", IsOfficial: true},
		{BaseModel: domain.BaseModel{ID: "platform-nonofficial"}, Title: "Invalid platform", CreatedBy: "root"},
	} {
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", course.ID, err)
		}
	}
	for index := 0; index < 8; index++ {
		resource := &domain.Resource{
			BaseModel: domain.BaseModel{ID: fmt.Sprintf("tenant-resource-%d", index), TenantID: "tenant-1"},
			Name:      fmt.Sprintf("tenant-%d.pdf", index), ResourceType: "attachment",
			URL: fmt.Sprintf("tenant-%d.pdf", index), CreatedBy: "admin-1",
		}
		if err := database.Create(resource).Error; err != nil {
			t.Fatalf("create tenant resource %d: %v", index, err)
		}
	}
	if err := database.Create(&domain.Resource{
		BaseModel: domain.BaseModel{ID: "foreign-resource", TenantID: "tenant-2"},
		Name:      "foreign.pdf", ResourceType: "attachment", URL: "foreign.pdf", CreatedBy: "admin-2",
	}).Error; err != nil {
		t.Fatalf("create foreign resource: %v", err)
	}
	if err := database.Create(&domain.Resource{
		BaseModel: domain.BaseModel{ID: "platform-resource"},
		Name:      "platform.pdf", ResourceType: "attachment", URL: "platform.pdf", CreatedBy: "root",
	}).Error; err != nil {
		t.Fatalf("create platform resource: %v", err)
	}

	repo := NewCourseMaterialRepository(database)
	base := context.Background()
	admin := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	instructor := usercontext.WithUser(base, "teacher-1", "tenant-1", "", "instructor")
	root := usercontext.WithUser(base, "root", "", "", "superadmin")
	tests := []struct {
		name     string
		ctx      context.Context
		material *domain.CourseMaterial
		allowed  bool
	}{
		{"tenant admin current nonofficial", admin, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-admin", TenantID: "tenant-1"}, CourseID: "tenant-course", ResourceID: "tenant-resource-0", DisplayName: "admin.pdf", CreatedBy: "admin-1"}, true},
		{"superadmin official empty scope", root, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-root"}, CourseID: "official-course", ResourceID: "platform-resource", DisplayName: "root.pdf", CreatedBy: "root"}, true},
		{"instructor", instructor, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-instructor", TenantID: "tenant-1"}, CourseID: "tenant-course", ResourceID: "tenant-resource-1", DisplayName: "instructor.pdf", CreatedBy: "teacher-1"}, false},
		{"unauthenticated", base, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-unauthenticated", TenantID: "tenant-1"}, CourseID: "tenant-course", ResourceID: "tenant-resource-2", DisplayName: "unauthenticated.pdf", CreatedBy: "admin-1"}, false},
		{"forged tenant", admin, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-forged", TenantID: "tenant-2"}, CourseID: "tenant-course", ResourceID: "tenant-resource-3", DisplayName: "forged.pdf", CreatedBy: "admin-1"}, false},
		{"foreign parent course", admin, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-foreign-course", TenantID: "tenant-1"}, CourseID: "foreign-course", ResourceID: "tenant-resource-4", DisplayName: "foreign-course.pdf", CreatedBy: "admin-1"}, false},
		{"tenant scoped official parent", admin, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-tenant-official", TenantID: "tenant-1"}, CourseID: "tenant-official", ResourceID: "tenant-resource-5", DisplayName: "tenant-official.pdf", CreatedBy: "admin-1"}, false},
		{"foreign resource", admin, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-foreign-resource", TenantID: "tenant-1"}, CourseID: "tenant-course", ResourceID: "foreign-resource", DisplayName: "foreign-resource.pdf", CreatedBy: "admin-1"}, false},
		{"superadmin tenant parent", root, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-root-tenant"}, CourseID: "tenant-course", ResourceID: "platform-resource", DisplayName: "root-tenant.pdf", CreatedBy: "root"}, false},
		{"superadmin nonofficial platform parent", root, &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-root-nonofficial"}, CourseID: "platform-nonofficial", ResourceID: "platform-resource", DisplayName: "root-nonofficial.pdf", CreatedBy: "root"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := repo.Create(test.ctx, test.material)
			if test.allowed {
				if err != nil {
					t.Fatalf("Create() error = %v", err)
				}
				return
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Create() error = %v, want record not found", err)
			}
			var count int64
			if countErr := database.Model(&domain.CourseMaterial{}).Where("id = ?", test.material.ID).Count(&count).Error; countErr != nil || count != 0 {
				t.Fatalf("material count = %d, error = %v", count, countErr)
			}
		})
	}
}
