package repository

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
)

func TestDomainBindJobRepositoryPersistsAndReservesDomains(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := NewDomainBindJobRepository(database)
	job := &domain.DomainBindJob{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, Domain: "learn.example.com", State: "verified", CurrentStep: 1}
	if err := repo.Reserve(ctx, job); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateStatus(ctx, "tenant-1", "configuring", "配置中", 3, ""); err != nil {
		t.Fatal(err)
	}
	if err := repo.IncrementAttempt(ctx, "tenant-1"); err != nil {
		t.Fatal(err)
	}

	reloaded := NewDomainBindJobRepository(database)
	got, err := reloaded.FindByTenant(ctx, "tenant-1")
	if err != nil || got.Domain != job.Domain || got.State != "configuring" || got.CurrentStep != 3 || got.AttemptCount != 1 {
		t.Fatalf("FindByTenant() = %#v, %v", got, err)
	}
	if err := reloaded.Reserve(ctx, &domain.DomainBindJob{BaseModel: domain.BaseModel{TenantID: "tenant-2"}, Domain: job.Domain, State: "verified"}); err == nil {
		t.Fatal("Reserve(duplicate domain) error = nil")
	}
}
