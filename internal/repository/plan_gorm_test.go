package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestPlanRepositoryCRUDAndQueries(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewPlanRepository(database)
	ctx := context.Background()
	if err := database.Where("is_default = ?", true).Delete(&domain.Plan{}).Error; err != nil {
		t.Fatalf("clear seeded default plan: %v", err)
	}
	defaultPlan := &domain.Plan{Name: "Free", StorageQuotaBytes: 100, MaxUsers: 10, MaxCourses: 2, Features: "{}", IsDefault: true, Status: 1}
	otherPlan := &domain.Plan{Name: "Pro", StorageQuotaBytes: 1000, MaxUsers: 100, MaxCourses: 20, Features: "{\"reports\":true}", Status: 1}
	for _, plan := range []*domain.Plan{defaultPlan, otherPlan} {
		if err := repo.Create(ctx, plan); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}
	if defaultPlan.ID == "" || otherPlan.ID == "" {
		t.Fatal("Create() did not generate plan IDs")
	}

	found, err := repo.FindByID(ctx, defaultPlan.ID)
	if err != nil || found.Name != "Free" {
		t.Fatalf("FindByID() = %#v, %v", found, err)
	}
	found, err = repo.FindDefault(ctx)
	if err != nil || found.ID != defaultPlan.ID {
		t.Fatalf("FindDefault() = %#v, %v", found, err)
	}
	plans, total, err := repo.List(ctx, 0, 10)
	if err != nil || total != 2 || len(plans) != 2 {
		t.Fatalf("List() = %#v, %d, %v", plans, total, err)
	}

	otherPlan.Name = "Professional"
	otherPlan.StorageQuotaBytes = 2000
	otherPlan.MaxUsers = 200
	otherPlan.MaxCourses = 30
	otherPlan.Features = "{\"reports\":true,\"sso\":true}"
	otherPlan.Status = 0
	if err := repo.Update(ctx, otherPlan); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err = repo.FindByID(ctx, otherPlan.ID)
	if err != nil || found.Name != "Professional" || found.Status != 0 || found.StorageQuotaBytes != 2000 {
		t.Fatalf("FindByID(updated) = %#v, %v", found, err)
	}
	if err := repo.Update(ctx, &domain.Plan{ID: "missing"}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Update(missing) error = %v, want record not found", err)
	}
	if err := repo.Delete(ctx, otherPlan.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := repo.Delete(ctx, otherPlan.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Delete(deleted) error = %v, want record not found", err)
	}
}
