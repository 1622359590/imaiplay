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

func TestCourseLessonRepositoryCRUDAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseLessonRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	first := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		ChapterID: "chapter-1", Title: "Second",
		ContentType: "video", DurationSeconds: 60, SortOrder: 2,
	}
	second := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		ChapterID: "chapter-1", Title: "First",
		ContentType: "text", SortOrder: 1,
	}
	foreign := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"},
		ChapterID: "chapter-1", Title: "Foreign",
		ContentType: "document",
	}
	for _, lesson := range []*domain.CourseLesson{first, second, foreign} {
		if err := repo.Create(base, lesson); err != nil {
			t.Fatalf("Create(%s) error = %v", lesson.Title, err)
		}
	}
	items, err := repo.FindByChapter(tenantOne, "chapter-1")
	if err != nil || len(items) != 2 || items[0].ID != second.ID {
		t.Fatalf("FindByChapter() = %#v, %v", items, err)
	}
	if _, err := repo.FindByID(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	first.Title, first.SortOrder = "Updated", 3
	if err := repo.Update(tenantOne, first); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if err := repo.Delete(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(tenantOne, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
