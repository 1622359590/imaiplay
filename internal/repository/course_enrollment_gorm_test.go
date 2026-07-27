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

func TestCourseEnrollmentRepositoryCRUDAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseEnrollmentRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	first := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-1", UserID: "learner-1", Status: 1,
	}
	second := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-1", UserID: "learner-2", Status: 1,
	}
	foreign := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"},
		CourseID:  "course-1", UserID: "learner-1", Status: 1,
	}
	for _, enrollment := range []*domain.CourseEnrollment{first, second, foreign} {
		if err := repo.Create(base, enrollment); err != nil {
			t.Fatalf("Create(%s) error = %v", enrollment.UserID, err)
		}
	}
	if _, err := repo.FindByID(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	items, err := repo.FindByCourse(tenantOne, "course-1")
	if err != nil || len(items) != 2 {
		t.Fatalf("FindByCourse() = %#v, %v", items, err)
	}
	items, err = repo.FindByUser(tenantOne, "learner-1")
	if err != nil || len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("FindByUser() = %#v, %v", items, err)
	}
	found, err := repo.FindByCourseAndUser(tenantOne, "course-1", "learner-1")
	if err != nil || found.ID != first.ID {
		t.Fatalf("FindByCourseAndUser() = %#v, %v", found, err)
	}
	duplicate := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-1", UserID: "learner-1", Status: 1,
	}
	if err := repo.Create(base, duplicate); err == nil {
		t.Fatal("Create(duplicate) error = nil")
	}
	if err := repo.Delete(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(tenantOne, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
