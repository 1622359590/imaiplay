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

func TestUserRepositoryCRUDAndTenantIsolation(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repository := NewUserRepository(database)
	base := context.Background()
	createRepositoryTenant(t, database, "tenant-1")
	createRepositoryTenant(t, database, "tenant-2")
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")

	first := newTestUser("tenant-1", "same@example.com", "First")
	second := newTestUser("tenant-2", "same@example.com", "Second")
	if err := repository.Create(base, first); err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	if err := repository.Create(base, second); err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	found, err := repository.FindByEmailAndTenant(
		base, "same@example.com", "tenant-1",
	)
	if err != nil || found.ID != first.ID {
		t.Fatalf("FindByEmailAndTenant() = %#v, %v", found, err)
	}
	if _, err := repository.FindByID(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v, want record not found", err)
	}

	users, total, err := repository.FindByTenant(base, "tenant-1", 0, 10)
	if err != nil || total != 1 || len(users) != 1 || users[0].ID != first.ID {
		t.Fatalf("FindByTenant() = %#v, %d, %v", users, total, err)
	}

	first.Name = "Updated"
	first.Status = 1
	if err := repository.Update(tenantOne, first); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err = repository.FindByID(tenantOne, first.ID)
	if err != nil || found.Name != "Updated" {
		t.Fatalf("FindByID() = %#v, %v", found, err)
	}
	if err := repository.Delete(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v, want record not found", err)
	}
	if err := repository.Delete(tenantOne, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.FindByID(tenantOne, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("FindByID(deleted) error = %v, want record not found", err)
	}
}

func TestUserRepositoryEnforcesTenantEmailUniqueness(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repository := NewUserRepository(database)
	createRepositoryTenant(t, database, "tenant-1")
	first := newTestUser("tenant-1", "duplicate@example.com", "First")
	second := newTestUser("tenant-1", "duplicate@example.com", "Second")
	if err := repository.Create(context.Background(), first); err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	if err := repository.Create(context.Background(), second); err == nil {
		t.Fatal("Create(duplicate) error = nil")
	}
}

func createRepositoryTenant(t *testing.T, database *gorm.DB, id string) {
	t.Helper()
	tenant := &domain.Tenant{ID: id, Code: id, Name: id, Status: 1}
	if err := database.Create(tenant).Error; err != nil {
		t.Fatalf("create test tenant %q: %v", id, err)
	}
}

func newTestUser(tenantID, email, name string) *domain.User {
	return &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Email:     email, Password: "hash", Name: name, Role: "learner",
	}
}
