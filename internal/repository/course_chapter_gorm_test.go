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

func TestCourseChapterRepositoryCRUDAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseChapterRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	first := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-1", Title: "Second", SortOrder: 2,
	}
	second := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-1", Title: "First", SortOrder: 1,
	}
	foreign := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"},
		CourseID:  "course-1", Title: "Foreign", SortOrder: 0,
	}
	for _, chapter := range []*domain.CourseChapter{first, second, foreign} {
		if err := repo.Create(base, chapter); err != nil {
			t.Fatalf("Create(%s) error = %v", chapter.Title, err)
		}
	}
	items, err := repo.FindByCourse(tenantOne, "course-1")
	if err != nil || len(items) != 2 || items[0].ID != second.ID {
		t.Fatalf("FindByCourse() = %#v, %v", items, err)
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

func TestCourseChapterRepositoryDeleteCascadesLessons(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	ctx := usercontext.WithUser(
		context.Background(), "admin", "tenant-1", "", "tenant_admin",
	)
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-1", Title: "Chapter",
	}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		ChapterID: chapter.ID, Title: "Lesson", ContentType: "text",
	}
	if err := database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	progress := &domain.LessonProgress{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner", LessonID: lesson.ID, ProgressPercent: 20,
	}
	if err := database.Create(progress).Error; err != nil {
		t.Fatalf("create progress: %v", err)
	}
	if err := NewCourseChapterRepository(database).Delete(ctx, chapter.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	var count int64
	if err := database.Model(&domain.CourseLesson{}).
		Where("id = ?", lesson.ID).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("lesson count=%d error=%v", count, err)
	}
	if err := database.Model(&domain.LessonProgress{}).
		Where("id = ?", progress.ID).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("progress count=%d error=%v", count, err)
	}
}

func TestCourseChapterRepositoryInstructorCannotMutateForeignCourse(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	course := newCourse("tenant-1", "owner", "Foreign", 0)
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter := &domain.CourseChapter{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: course.ID, Title: "Chapter"}
	if err := database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	repo := NewCourseChapterRepository(database)
	instructor := usercontext.WithUser(context.Background(), "other", "tenant-1", "", "instructor")
	chapter.Title = "Changed"
	if err := repo.Update(instructor, chapter); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Update(foreign) error = %v", err)
	}
	if err := repo.Delete(instructor, chapter.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Delete(foreign) error = %v", err)
	}
}
