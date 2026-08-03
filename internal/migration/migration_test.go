package migration

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrateCreatesTenantAndUserTables(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if !database.Migrator().HasTable(&domain.Tenant{}) {
		t.Fatal("AutoMigrate() did not create tenants table")
	}
	if !database.Migrator().HasTable(&domain.User{}) {
		t.Fatal("AutoMigrate() did not create users table")
	}
	if !database.Migrator().HasConstraint(&domain.User{}, "Tenant") {
		t.Fatal("AutoMigrate() did not create users tenant foreign key")
	}
	for name, model := range map[string]interface{}{
		"courses":                 &domain.Course{},
		"course_chapters":         &domain.CourseChapter{},
		"course_lessons":          &domain.CourseLesson{},
		"course_enrollments":      &domain.CourseEnrollment{},
		"lesson_progress":         &domain.LessonProgress{},
		"resources":               &domain.Resource{},
		"resource_categories":     &domain.ResourceCategory{},
		"refresh_tokens":          &domain.RefreshToken{},
		"password_resets":         &domain.PasswordReset{},
		"audit_logs":              &domain.AuditLog{},
		"plans":                   &domain.Plan{},
		"tenant_official_courses": &domain.TenantOfficialCourse{},
		"login_challenges":        &domain.LoginChallenge{},
	} {
		if !database.Migrator().HasTable(model) {
			t.Fatalf("AutoMigrate() did not create %s table", name)
		}
	}
	if !database.Migrator().HasTable("schema_migrations") || !database.Migrator().HasColumn(&domain.CourseLesson{}, "resource_id") {
		t.Fatal("versioned migrations did not create schema metadata or resource_id")
	}
	var count int64
	if err := database.Table("schema_migrations").Count(&count).Error; err != nil || count != 13 {
		t.Fatalf("schema migrations count = %d, err=%v", count, err)
	}
	for _, name := range []string{
		"idx_users_email_lookup",
		"idx_users_phone_lookup",
	} {
		if !database.Migrator().HasIndex(&domain.User{}, name) {
			t.Fatalf("AutoMigrate() did not create %s", name)
		}
	}
	if err := AutoMigrate(database); err != nil {
		t.Fatalf("repeat AutoMigrate() error = %v", err)
	}
	if err := database.Table("schema_migrations").Count(&count).Error; err != nil || count != 13 {
		t.Fatalf("repeat schema migrations count = %d, err=%v", count, err)
	}
}
