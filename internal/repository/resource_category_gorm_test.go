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

func TestResourceCategoryRepositoryCRUDAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewResourceCategoryRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	parent := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"}, Name: "Videos",
	}
	if err := repo.Create(tenantOne, parent); err != nil {
		t.Fatalf("Create(parent) error = %v", err)
	}
	child := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "Onboarding", ParentID: &parent.ID,
	}
	foreign := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"}, Name: "Foreign",
	}
	for index, category := range []*domain.ResourceCategory{child, foreign} {
		ctx := tenantOne
		if index == 1 {
			ctx = tenantTwo
		}
		if err := repo.Create(ctx, category); err != nil {
			t.Fatalf("Create(%s) error = %v", category.Name, err)
		}
	}
	if _, err := repo.FindByID(tenantTwo, child.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	items, err := repo.FindByTenant(tenantOne, "tenant-1")
	if err != nil || len(items) != 2 {
		t.Fatalf("FindByTenant() = %#v, %v", items, err)
	}
	child.Name = "Updated"
	if err := repo.Update(tenantOne, child); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err := repo.FindByID(tenantOne, child.ID)
	if err != nil || found.Name != "Updated" || found.ParentID == nil {
		t.Fatalf("FindByID(updated) = %#v, %v", found, err)
	}
	if err := repo.Delete(tenantTwo, child.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	resource := &domain.Resource{
		BaseModel:  domain.BaseModel{TenantID: "tenant-1"},
		CategoryID: &child.ID, Name: "video.mp4", ResourceType: "video",
		URL: "/uploads/video.mp4", CreatedBy: "admin-1",
	}
	if err := database.Create(resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if err := repo.Delete(tenantOne, child.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := database.First(resource, "id = ?", resource.ID).Error; err != nil {
		t.Fatalf("reload resource: %v", err)
	}
	if resource.CategoryID != nil {
		t.Fatalf("resource category after Delete() = %v", *resource.CategoryID)
	}
}

func TestResourceCategoryRepositoryMutationsRequireTenantAdminScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	for _, category := range []*domain.ResourceCategory{
		{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-1"}, Name: "Current update"},
		{BaseModel: domain.BaseModel{ID: "current-delete-instructor", TenantID: "tenant-1"}, Name: "Instructor delete"},
		{BaseModel: domain.BaseModel{ID: "current-delete-unauthenticated", TenantID: "tenant-1"}, Name: "Unauthenticated delete"},
		{BaseModel: domain.BaseModel{ID: "current-delete-admin", TenantID: "tenant-1"}, Name: "Admin delete"},
		{BaseModel: domain.BaseModel{ID: "foreign", TenantID: "tenant-2"}, Name: "Foreign"},
	} {
		if err := database.Create(category).Error; err != nil {
			t.Fatalf("create category %s: %v", category.ID, err)
		}
	}
	repo := NewResourceCategoryRepository(database)
	base := context.Background()
	admin := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	foreignAdmin := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	instructor := usercontext.WithUser(base, "teacher-1", "tenant-1", "", "instructor")

	created := &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "created"}, Name: "Created"}
	if err := repo.Create(admin, created); err != nil {
		t.Errorf("tenant admin Create() error = %v", err)
	}
	if created.TenantID != "tenant-1" {
		t.Errorf("created tenant = %q, want tenant-1", created.TenantID)
	}
	for _, test := range []struct {
		name     string
		ctx      context.Context
		category *domain.ResourceCategory
	}{
		{"instructor", instructor, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "created-instructor", TenantID: "tenant-1"}, Name: "Instructor"}},
		{"unauthenticated", base, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "created-unauthenticated", TenantID: "tenant-1"}, Name: "Unauthenticated"}},
		{"forged tenant", admin, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "created-forged", TenantID: "tenant-2"}, Name: "Forged"}},
		{"foreign parent", admin, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "created-foreign-parent", TenantID: "tenant-1"}, Name: "Foreign parent", ParentID: stringPointer("foreign")}},
	} {
		t.Run("create "+test.name, func(t *testing.T) {
			if err := repo.Create(test.ctx, test.category); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Create() error = %v, want record not found", err)
			}
			var count int64
			if err := database.Model(&domain.ResourceCategory{}).Where("id = ?", test.category.ID).Count(&count).Error; err != nil || count != 0 {
				t.Fatalf("category count = %d, error = %v", count, err)
			}
		})
	}

	updated := &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-1"}, Name: "Updated"}
	if err := repo.Update(admin, updated); err != nil {
		t.Errorf("tenant admin Update() error = %v", err)
	}
	for _, test := range []struct {
		name     string
		ctx      context.Context
		category *domain.ResourceCategory
	}{
		{"instructor", instructor, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-1"}, Name: "Instructor changed"}},
		{"unauthenticated", base, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-1"}, Name: "Unauthenticated changed"}},
		{"cross tenant", foreignAdmin, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-2"}, Name: "Foreign changed"}},
		{"forged tenant", admin, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-2"}, Name: "Forged changed"}},
		{"foreign parent", admin, &domain.ResourceCategory{BaseModel: domain.BaseModel{ID: "current-update", TenantID: "tenant-1"}, Name: "Foreign parent", ParentID: stringPointer("foreign")}},
	} {
		t.Run("update "+test.name, func(t *testing.T) {
			if err := repo.Update(test.ctx, test.category); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Update() error = %v, want record not found", err)
			}
		})
	}
	var stored domain.ResourceCategory
	if err := database.First(&stored, "id = ?", "current-update").Error; err != nil || stored.Name != "Updated" {
		t.Errorf("stored category = %#v, error = %v", stored, err)
	}

	for _, test := range []struct {
		name, id string
		ctx      context.Context
	}{
		{"instructor", "current-delete-instructor", instructor},
		{"unauthenticated", "current-delete-unauthenticated", base},
		{"cross tenant", "current-delete-instructor", foreignAdmin},
	} {
		t.Run("delete "+test.name, func(t *testing.T) {
			if err := repo.Delete(test.ctx, test.id); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Delete() error = %v, want record not found", err)
			}
			var count int64
			if err := database.Model(&domain.ResourceCategory{}).Where("id = ?", test.id).Count(&count).Error; err != nil || count != 1 {
				t.Fatalf("category count = %d, error = %v", count, err)
			}
		})
	}
	if err := repo.Delete(admin, "current-delete-admin"); err != nil {
		t.Errorf("tenant admin Delete() error = %v", err)
	}
}

func stringPointer(value string) *string { return &value }
