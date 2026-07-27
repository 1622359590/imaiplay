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

func TestResourceRepositoryCRUDAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewResourceRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	first := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "guide.pdf", ResourceType: "document", URL: "/uploads/guide.pdf",
		SizeBytes: 100, CreatedBy: "admin-1",
	}
	second := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "video.mp4", ResourceType: "video", URL: "/uploads/video.mp4",
		SizeBytes: 200, CreatedBy: "admin-1",
	}
	foreign := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"},
		Name:      "foreign.png", ResourceType: "image", URL: "/uploads/foreign.png",
		SizeBytes: 50, CreatedBy: "admin-2",
	}
	for _, resource := range []*domain.Resource{first, second, foreign} {
		if err := repo.Create(base, resource); err != nil {
			t.Fatalf("Create(%s) error = %v", resource.Name, err)
		}
	}
	if _, err := repo.FindByID(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	items, total, err := repo.FindByTenant(tenantOne, "tenant-1", 0, 1)
	if err != nil || total != 2 || len(items) != 1 {
		t.Fatalf("FindByTenant() = %#v, %d, %v", items, total, err)
	}
	first.Name = "updated.pdf"
	if err := repo.Update(tenantOne, first); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err := repo.FindByID(tenantOne, first.ID)
	if err != nil || found.Name != "updated.pdf" {
		t.Fatalf("FindByID(updated) = %#v, %v", found, err)
	}
	if err := repo.Delete(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(tenantOne, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
