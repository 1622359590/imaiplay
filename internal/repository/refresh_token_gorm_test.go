package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestRefreshTokenRepositoryLifecycle(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewRefreshTokenRepository(database)
	ctx := context.Background()
	first := &domain.RefreshToken{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "user-1", TokenHash: "hash-1", ExpiresAt: time.Now().Add(time.Hour)}
	second := &domain.RefreshToken{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "user-1", TokenHash: "hash-2", ExpiresAt: time.Now().Add(time.Hour)}
	otherUser := &domain.RefreshToken{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, UserID: "user-2", TokenHash: "hash-3", ExpiresAt: time.Now().Add(time.Hour)}
	for _, token := range []*domain.RefreshToken{first, second, otherUser} {
		if err := repo.Create(ctx, token); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	found, err := repo.FindValidByHash(ctx, "hash-1")
	if err != nil || found.ID != first.ID {
		t.Fatalf("FindValidByHash() = %#v, %v", found, err)
	}
	if err := repo.Revoke(ctx, "hash-1"); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if _, err := repo.FindValidByHash(ctx, "hash-1"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("FindValidByHash(revoked) error = %v, want record not found", err)
	}
	if err := repo.Revoke(ctx, "missing"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Revoke(missing) error = %v, want record not found", err)
	}
	if err := repo.RevokeAllForUser(ctx, "user-1"); err != nil {
		t.Fatalf("RevokeAllForUser() error = %v", err)
	}
	if _, err := repo.FindValidByHash(ctx, "hash-2"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("FindValidByHash(revoked user) error = %v, want record not found", err)
	}
	if found, err := repo.FindValidByHash(ctx, "hash-3"); err != nil || found.ID != otherUser.ID {
		t.Fatalf("FindValidByHash(other user) = %#v, %v", found, err)
	}
}
