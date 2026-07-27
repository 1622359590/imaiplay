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

func TestLessonProgressRepositoryUpsertAndScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewLessonProgressRepository(database)
	base := context.Background()
	learner := usercontext.WithUser(base, "learner-1", "tenant-1", "", "learner")
	otherTenant := usercontext.WithUser(base, "learner-1", "tenant-2", "", "learner")
	progress := &domain.LessonProgress{
		BaseModel:           domain.BaseModel{TenantID: "tenant-1"},
		UserID:              "learner-1",
		LessonID:            "lesson-1",
		ProgressPercent:     20,
		Status:              1,
		LastPositionSeconds: 12,
	}
	if err := repo.Create(base, progress); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	originalID := progress.ID
	if _, err := repo.FindByID(otherTenant, originalID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	progress.ProgressPercent = 100
	progress.Status = 2
	progress.LastPositionSeconds = 60
	if err := repo.Upsert(learner, progress); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	found, err := repo.FindByUserAndLesson(learner, "learner-1", "lesson-1")
	if err != nil {
		t.Fatalf("FindByUserAndLesson() error = %v", err)
	}
	if found.ID != originalID || found.ProgressPercent != 100 ||
		found.LastPositionSeconds != 60 || found.Status != 2 {
		t.Fatalf("FindByUserAndLesson() = %#v", found)
	}
	second := &domain.LessonProgress{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner-1", LessonID: "lesson-2", Status: 1,
	}
	if err := repo.Upsert(learner, second); err != nil {
		t.Fatalf("Upsert(new) error = %v", err)
	}
	items, total, err := repo.FindByUser(learner, "learner-1", 0, 1)
	if err != nil || total != 2 || len(items) != 1 {
		t.Fatalf("FindByUser() = %#v, %d, %v", items, total, err)
	}
	if !items[0].UpdatedAt.After(items[len(items)-1].UpdatedAt) &&
		items[0].ID != second.ID {
		t.Fatalf("FindByUser() is not ordered by recent update: %#v", items)
	}
}

func TestLessonProgressRepositoryUniquePerTenantUserLesson(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewLessonProgressRepository(database)
	base := context.Background()
	first := &domain.LessonProgress{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner-1", LessonID: "lesson-1",
	}
	if err := repo.Create(base, first); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	duplicate := &domain.LessonProgress{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner-1", LessonID: "lesson-1",
	}
	if err := repo.Create(base, duplicate); err == nil {
		t.Fatal("Create(duplicate) error = nil")
	}
}
