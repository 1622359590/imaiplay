package repository

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantRepositoryCRUD(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repository := NewTenantRepository(database)
	ctx := context.Background()

	first := &domain.Tenant{Code: "acme", Name: "Acme Training"}
	if err := repository.Create(ctx, first); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := uuid.Parse(first.ID); err != nil {
		t.Fatalf("Create() ID = %q, want UUID: %v", first.ID, err)
	}
	if first.Status != 1 {
		t.Fatalf("Create() status = %d, want 1", first.Status)
	}

	found, err := repository.FindByCode(ctx, "acme")
	if err != nil {
		t.Fatalf("FindByCode() error = %v", err)
	}
	if found.ID != first.ID || found.Code != "acme" || found.Name != "Acme Training" {
		t.Fatalf("FindByCode() = %#v", found)
	}

	second := &domain.Tenant{Code: "globex", Name: "Globex Academy"}
	if err := repository.Create(ctx, second); err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	all, err := repository.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("FindAll() returned %d tenants, want 2", len(all))
	}
	codes := map[string]bool{all[0].Code: true, all[1].Code: true}
	if !codes["acme"] || !codes["globex"] {
		t.Fatalf("FindAll() codes = %#v", codes)
	}
}

func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close sqlite: %v", err)
		}
	})
	return database
}
