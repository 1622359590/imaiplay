package service

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
)

func TestUserServiceCRUDAndTenantAdminAuthorization(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	service := NewUserService(userRepo)
	admin := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "admin@example.com", "tenant_admin",
	)
	learner := usercontext.WithUser(
		context.Background(), "learner", "tenant-1", "learner@example.com", "learner",
	)

	if _, err := service.Create(
		learner, "new@example.com", "password123", "New", "learner",
	); errorCode(err) != 40300 {
		t.Fatalf("Create(learner) error = %#v", err)
	}
	created, err := service.Create(
		admin, "new@example.com", "password123", "New", "learner",
	)
	if err != nil || created.TenantID != "tenant-1" ||
		!security.CheckPassword("password123", created.Password) {
		t.Fatalf("Create() = %#v, %v", created, err)
	}
	items, total, err := service.List(admin, 0, 20)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("List() = %#v, %d, %v", items, total, err)
	}
	got, err := service.Get(admin, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("Get() = %#v, %v", got, err)
	}
	updated, err := service.Update(admin, created.ID, "Updated", 0, "newpass123")
	if err != nil || updated.Name != "Updated" || updated.Status != 0 ||
		!security.CheckPassword("newpass123", updated.Password) {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if err := service.Delete(admin, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(admin, created.ID); errorCode(err) != 40400 {
		t.Fatalf("Get(deleted) error = %#v", err)
	}
}

func TestUserServiceSuperadminListsUsersAcrossTenants(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	service := NewUserService(userRepo)
	for _, code := range []string{"one", "two"} {
		ctx := usercontext.WithUser(context.Background(), "admin-"+code, code, code+"@admin.example.com", "tenant_admin")
		if _, err := service.Create(ctx, code+"@example.com", "password123", code, "learner"); err != nil {
			t.Fatal(err)
		}
	}
	superadmin := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	items, total, err := service.List(superadmin, 0, 20)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("superadmin List() = %#v, %d, %v", items, total, err)
	}
}

func TestUserServiceRejectsSuperadminRole(t *testing.T) {
	_, _, userRepo := serviceRepositories(t)
	service := NewUserService(userRepo)
	admin := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "admin@example.com", "tenant_admin",
	)
	if _, err := service.Create(
		admin, "root@example.com", "password123", "Root", "superadmin",
	); errorCode(err) != 40000 {
		t.Fatalf("Create(superadmin) error = %#v", err)
	}
}

func TestUserServiceCreateRejectsWhenTenantEmployeeLimitReached(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	planRepo := repository.NewPlanRepository(database)
	ctx := context.Background()
	plan := &domain.Plan{Name: "Single employee", MaxUsers: 1, Status: 1}
	if err := planRepo.Create(ctx, plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "limited", Name: "Limited", Status: 1, PlanID: &plan.ID}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatal(err)
	}
	service := NewUserService(userRepo, UserLimitRepositories{Tenants: tenantRepo, Plans: planRepo})
	admin := usercontext.WithUser(ctx, "admin", tenant.ID, "admin@example.com", "tenant_admin")

	if _, err := service.Create(admin, "one@example.com", "password123", "One", "learner"); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if _, err := service.Create(admin, "two@example.com", "password123", "Two", "learner"); errorCode(err) != 40300 || err.Error() != "员工数已达套餐上限，请升级套餐" {
		t.Fatalf("second Create() error = %#v", err)
	}
}

func TestUserServiceCreateAllowsUnlimitedEmployeesWhenMaxUsersIsZero(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	planRepo := repository.NewPlanRepository(database)
	ctx := context.Background()
	plan := &domain.Plan{Name: "Unlimited", MaxUsers: 0, Status: 1}
	if err := planRepo.Create(ctx, plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "unlimited", Name: "Unlimited", Status: 1, PlanID: &plan.ID}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatal(err)
	}
	service := NewUserService(userRepo, UserLimitRepositories{Tenants: tenantRepo, Plans: planRepo})
	admin := usercontext.WithUser(ctx, "admin", tenant.ID, "admin@example.com", "tenant_admin")

	for _, email := range []string{"one@example.com", "two@example.com"} {
		if _, err := service.Create(admin, email, "password123", email, "learner"); err != nil {
			t.Fatalf("Create(%q) error = %v", email, err)
		}
	}
}
