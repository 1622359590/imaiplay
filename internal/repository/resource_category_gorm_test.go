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
	if err := repo.Create(base, parent); err != nil {
		t.Fatalf("Create(parent) error = %v", err)
	}
	child := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "Onboarding", ParentID: &parent.ID,
	}
	foreign := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"}, Name: "Foreign",
	}
	for _, category := range []*domain.ResourceCategory{child, foreign} {
		if err := repo.Create(base, category); err != nil {
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
