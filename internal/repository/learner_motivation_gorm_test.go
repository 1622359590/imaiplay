package repository

import (
	"context"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
)

func TestLearnerMotivationRepositoryMarksFirstLoginOnce(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	createRepositoryTenant(t, database, "tenant-first-login")
	user := newTestUser("tenant-first-login", "first-login@example.com", "First Login")
	if err := NewUserRepository(database).Create(context.Background(), user); err != nil {
		t.Fatalf("create learner: %v", err)
	}

	repository := NewLearnerMotivationRepository(database)
	ctx := usercontext.WithUser(context.Background(), user.ID, user.TenantID, user.Email, user.Role)
	firstAt := time.Date(2026, time.August, 20, 2, 30, 0, 0, time.UTC)
	changed, err := repository.MarkFirstLogin(ctx, user.ID, firstAt)
	if err != nil || !changed {
		t.Fatalf("first MarkFirstLogin() = %v, %v, want true, nil", changed, err)
	}
	changed, err = repository.MarkFirstLogin(ctx, user.ID, firstAt.Add(time.Hour))
	if err != nil || changed {
		t.Fatalf("second MarkFirstLogin() = %v, %v, want false, nil", changed, err)
	}

	var state domain.LearnerEngagementState
	if err := database.Where("tenant_id = ? AND user_id = ?", user.TenantID, user.ID).First(&state).Error; err != nil {
		t.Fatalf("load engagement state: %v", err)
	}
	if state.FirstLoginAt == nil || !state.FirstLoginAt.Equal(firstAt) {
		t.Fatalf("first_login_at = %v, want %v", state.FirstLoginAt, firstAt)
	}
	if _, err := repository.MarkFirstLogin(context.Background(), user.ID, firstAt); err == nil {
		t.Fatal("MarkFirstLogin() accepted a missing learner identity")
	}
}
