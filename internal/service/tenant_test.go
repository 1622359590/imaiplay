package service

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
)

func TestTenantServiceCRUDAndSuperadminAuthorization(t *testing.T) {
	_, tenantRepo, _ := serviceRepositories(t)
	service := NewTenantService(tenantRepo)
	superadmin := usercontext.WithUser(
		context.Background(), "root", "", "root@example.com", "superadmin",
	)
	tenantAdmin := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "a@example.com", "tenant_admin",
	)

	if _, err := service.Create(tenantAdmin, "acme", "Acme"); errorCode(err) != 40300 {
		t.Fatalf("Create(tenant_admin) error = %#v", err)
	}
	created, err := service.Create(superadmin, "acme", "Acme")
	if err != nil || created.Status != 1 {
		t.Fatalf("Create() = %#v, %v", created, err)
	}
	if _, err := service.Create(superadmin, "acme", "Duplicate"); errorCode(err) != 40900 {
		t.Fatalf("Create(duplicate) error = %#v", err)
	}
	items, total, err := service.List(superadmin, 0, 20)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("List() = %#v, %d, %v", items, total, err)
	}
	got, err := service.Get(superadmin, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("Get() = %#v, %v", got, err)
	}
	updated, err := service.Update(superadmin, created.ID, "Acme Academy", 0)
	if err != nil || updated.Code != "acme" ||
		updated.Name != "Acme Academy" || updated.Status != 0 {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if err := service.Delete(superadmin, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(superadmin, created.ID); errorCode(err) != 40400 {
		t.Fatalf("Get(deleted) error = %#v", err)
	}
}

func TestTenantServiceDeleteMapsForeignKeyToConflict(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	if err := database.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	tenant := &domain.Tenant{Code: "protected", Name: "Protected", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if err := userRepo.Create(context.Background(), &domain.User{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Email: "user@example.com", Password: "hash", Name: "User", Role: "learner", Status: 1}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	superadmin := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	if errorCode(NewTenantService(tenantRepo).Delete(superadmin, tenant.ID)) != 40900 {
		t.Fatalf("delete error = %#v", NewTenantService(tenantRepo).Delete(superadmin, tenant.ID))
	}
}
