package service

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPlanStorageQuota(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	tenantRepo := repository.NewTenantRepository(database)
	planRepo := repository.NewPlanRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	plan := &domain.Plan{Name: "Small", StorageQuotaBytes: 100, Features: "{}", Status: 1}
	if err := planRepo.Create(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{ID: "tenant-plan", Code: "plan", Name: "Plan", PlanID: &plan.ID}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	resource := &domain.Resource{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Name: "used", ResourceType: "image", URL: "/uploads/used", SizeBytes: 80, CreatedBy: "admin"}
	if err := resourceRepo.Create(context.Background(), resource); err != nil {
		t.Fatal(err)
	}
	service := NewPlanService(planRepo, tenantRepo, resourceRepo)
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	if err := service.CheckStorage(ctx, tenant.ID, 21); err == nil {
		t.Fatal("expected quota error")
	}
	if err := service.CheckStorage(ctx, tenant.ID, 20); err != nil {
		t.Fatalf("exact quota should pass: %v", err)
	}
	current, err := service.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if current["used_bytes"] != int64(80) || current["quota_bytes"] != int64(100) {
		t.Fatalf("unexpected current usage: %#v", current)
	}
}

func TestPlanUpdatePreservesDefaultFlagWhenPayloadOmitsIt(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	planRepo := repository.NewPlanRepository(database)
	defaultPlan, err := planRepo.FindDefault(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	service := NewPlanService(planRepo, repository.NewTenantRepository(database), repository.NewResourceRepository(database))
	root := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")

	updated, err := service.Update(root, &domain.Plan{
		ID: defaultPlan.ID, Name: "免费版（更新）", StorageQuotaBytes: 2048,
		Status: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.IsDefault {
		t.Fatal("editing the default plan cleared its default flag")
	}
	if updated.Features != defaultPlan.Features {
		t.Fatalf("features = %q, want %q", updated.Features, defaultPlan.Features)
	}
}

func TestPlanDeleteRejectsDefaultPlan(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	planRepo := repository.NewPlanRepository(database)
	defaultPlan, err := planRepo.FindDefault(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	service := NewPlanService(planRepo, repository.NewTenantRepository(database), repository.NewResourceRepository(database))
	root := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")

	err = service.Delete(root, defaultPlan.ID)
	if errorCode(err) != 40000 {
		t.Fatalf("Delete(default) error = %v, want bad request", err)
	}
}
