package repository

import (
	"context"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
)

func TestPasswordResetRepositoryCreateFindAndUpdate(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewPasswordResetRepository(database)
	ctx := context.Background()
	now := time.Now().UTC()
	older := &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, Phone: "13800000000", Purpose: "password_reset", CodeHash: "old", ExpiresAt: now.Add(-time.Hour)}
	latest := &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, Phone: "13800000000", Purpose: "password_reset", CodeHash: "latest", ExpiresAt: now.Add(time.Hour)}
	otherPurpose := &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, Phone: "13800000000", Purpose: "login_code", CodeHash: "login", ExpiresAt: now.Add(time.Hour)}
	for _, reset := range []*domain.PasswordReset{older, latest, otherPurpose} {
		if err := repo.Create(ctx, reset); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}
	if err := database.Model(older).Update("created_at", now.Add(-time.Hour)).Error; err != nil {
		t.Fatalf("set older created_at: %v", err)
	}
	if err := database.Model(latest).Update("created_at", now).Error; err != nil {
		t.Fatalf("set latest created_at: %v", err)
	}

	found, err := repo.FindLatest(ctx, "tenant-1", "13800000000")
	if err != nil || found.ID != latest.ID {
		t.Fatalf("FindLatest() = %#v, %v", found, err)
	}
	found, err = repo.FindLatestForPurpose(ctx, "tenant-1", "13800000000", "login_code")
	if err != nil || found.ID != otherPurpose.ID {
		t.Fatalf("FindLatestForPurpose() = %#v, %v", found, err)
	}

	if err := repo.IncrementAttempts(ctx, latest.ID); err != nil {
		t.Fatalf("IncrementAttempts() error = %v", err)
	}
	var updated domain.PasswordReset
	if err := database.First(&updated, "id = ?", latest.ID).Error; err != nil {
		t.Fatalf("reload reset: %v", err)
	}
	if updated.Attempts != 1 {
		t.Fatalf("Attempts = %d, want 1", updated.Attempts)
	}
	if err := repo.MarkUsed(ctx, latest.ID); err != nil {
		t.Fatalf("MarkUsed() error = %v", err)
	}
	if err := database.First(&updated, "id = ?", latest.ID).Error; err != nil {
		t.Fatalf("reload used reset: %v", err)
	}
	if !updated.Used {
		t.Fatal("MarkUsed() did not set Used")
	}
}
