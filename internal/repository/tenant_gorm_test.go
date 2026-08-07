package repository

import (
	"context"
	"fmt"
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
	found, err = repository.FindByID(ctx, first.ID)
	if err != nil || found.Code != "acme" {
		t.Fatalf("FindByID() = %#v, %v", found, err)
	}
	first.Name = "Acme Academy"
	first.Status = 1
	first.Code = "renamed"
	if err := repository.Update(ctx, first); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err = repository.FindByID(ctx, first.ID)
	if err != nil {
		t.Fatalf("FindByID(after update) error = %v", err)
	}
	if found.Code != "acme" {
		t.Fatalf("Update() changed code to %q, want acme", found.Code)
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
	if err := repository.Delete(ctx, second.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.FindByID(ctx, second.ID); err == nil {
		t.Fatal("FindByID(deleted) error = nil")
	}
}

func TestTenantRepositoryUpdateThemePersistsBrandName(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repository := NewTenantRepository(database)
	tenant := &domain.Tenant{Code: "brand", Name: "Brand Tenant", BrandName: "Old Brand"}
	if err := repository.Create(context.Background(), tenant); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	tenant.BrandName = "New Brand"
	if err := repository.UpdateTheme(context.Background(), tenant); err != nil {
		t.Fatalf("UpdateTheme() error = %v", err)
	}
	stored, err := repository.FindByID(context.Background(), tenant.ID)
	if err != nil || stored.BrandName != "New Brand" {
		t.Fatalf("stored brand = %q, err=%v", stored.BrandName, err)
	}
}

func TestTenantRepositoryDoesNotDeleteTenantWithUsers(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	tenantRepo := NewTenantRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()
	tenant := &domain.Tenant{Code: "protected", Name: "Protected", Status: 1}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("Create(tenant) error = %v", err)
	}
	user := newTestUser(tenant.ID, "admin@example.com", "Admin")
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create(user) error = %v", err)
	}
	if err := tenantRepo.Delete(ctx, tenant.ID); err == nil {
		t.Fatal("Delete(tenant with users) error = nil")
	}
}

func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf(
		"file:%s?mode=memory&cache=shared&_foreign_keys=1",
		uuid.NewString(),
	)
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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
