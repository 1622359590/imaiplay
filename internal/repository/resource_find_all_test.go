package repository

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
)

func TestResourceRepositoryFindAllAcrossTenants(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewResourceRepository(database)
	for _, resource := range []*domain.Resource{
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-1"},
			Name:      "one.pdf", ResourceType: "document", URL: "/uploads/one.pdf",
			CreatedBy: "admin-1",
		},
		{
			BaseModel: domain.BaseModel{TenantID: "tenant-2"},
			Name:      "two.mp4", ResourceType: "video", URL: "/uploads/two.mp4",
			CreatedBy: "admin-2",
		},
	} {
		if err := repo.Create(context.Background(), resource); err != nil {
			t.Fatalf("Create(%s) error = %v", resource.Name, err)
		}
	}
	items, total, err := repo.FindAll(context.Background(), 0, 20)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("FindAll() = %#v, %d, %v", items, total, err)
	}
}
