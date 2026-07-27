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

func TestCourseRepositoryCRUDScopeAndPublishedList(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseRepository(database)
	base := context.Background()
	admin := usercontext.WithUser(base, "admin", "tenant-1", "", "tenant_admin")
	instructor := usercontext.WithUser(base, "author-1", "tenant-1", "", "instructor")
	otherTenant := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")

	draft := newCourse("tenant-1", "author-1", "Draft", 0)
	published := newCourse("tenant-1", "author-2", "Published", 1)
	foreign := newCourse("tenant-2", "author-1", "Foreign", 1)
	for _, course := range []*domain.Course{draft, published, foreign} {
		if err := repo.Create(base, course); err != nil {
			t.Fatalf("Create(%s) error = %v", course.Title, err)
		}
	}

	if _, err := repo.FindByID(otherTenant, draft.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	items, total, err := repo.FindByTenant(admin, "tenant-1", 0, 10)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("admin FindByTenant() = %#v, %d, %v", items, total, err)
	}
	items, total, err = repo.FindByTenant(instructor, "tenant-1", 0, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != draft.ID {
		t.Fatalf("instructor FindByTenant() = %#v, %d, %v", items, total, err)
	}
	items, total, err = repo.FindPublishedByTenant(base, "tenant-1", 0, 10)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != published.ID {
		t.Fatalf("FindPublishedByTenant() = %#v, %d, %v", items, total, err)
	}

	draft.Title, draft.Status = "Updated", 1
	if err := repo.Update(admin, draft); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err := repo.FindByID(admin, draft.ID)
	if err != nil || found.Title != "Updated" || found.Status != 1 {
		t.Fatalf("FindByID(updated) = %#v, %v", found, err)
	}
	if err := repo.Delete(otherTenant, draft.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(admin, draft.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestCourseRepositoryDeleteCascadesContentWithinTenant(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	ctx := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "", "tenant_admin",
	)
	course := newCourse("tenant-1", "author", "Course", 0)
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  course.ID, Title: "Chapter",
	}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "video",
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	if err := NewCourseRepository(database).Delete(ctx, course.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	for name, model := range map[string]interface{}{
		"course": course, "chapter": chapter, "lesson": lesson,
	} {
		var count int64
		if err := database.Model(model).Where("id = ?", modelID(model)).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("%s count=%d error=%v", name, count, err)
		}
	}
}

func modelID(model interface{}) string {
	switch value := model.(type) {
	case *domain.Course:
		return value.ID
	case *domain.CourseChapter:
		return value.ID
	case *domain.CourseLesson:
		return value.ID
	default:
		return ""
	}
}

func newCourse(tenantID, creator, title string, status int) *domain.Course {
	return &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Title:     title, Status: status, CreatedBy: creator,
	}
}
