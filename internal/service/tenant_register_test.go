package service

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantRegistrationCreatesTenantAdminAndDemoData(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	service := NewTenantRegistrationService(database, "secret")

	result, err := service.Register(context.Background(), "Acme 公司", "ADMIN@ACME.COM", "管理员", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if result.Tenant.Code != "acme" || result.User.Role != "tenant_admin" || result.Token == "" {
		t.Fatalf("registration result = %#v", result)
	}
	var users []domain.User
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&users).Error; err != nil {
		t.Fatalf("find users: %v", err)
	}
	if len(users) != 4 {
		t.Fatalf("user count = %d, want 4", len(users))
	}
	var courses []domain.Course
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&courses).Error; err != nil {
		t.Fatalf("find courses: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != demoCourseTitle {
		t.Fatalf("courses = %#v", courses)
	}
	var resources []domain.Resource
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&resources).Error; err != nil {
		t.Fatalf("find resources: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("resource count = %d, want 2", len(resources))
	}
}

func TestTenantRegistrationUsesUniqueSlugAndClearsDemoData(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	service := NewTenantRegistrationService(database, "secret")
	first, err := service.Register(context.Background(), "Acme Inc", "one@example.com", "One", "password123")
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	second, err := service.Register(context.Background(), "Acme Inc", "two@example.com", "Two", "password123")
	if err != nil {
		t.Fatalf("second register: %v", err)
	}
	if first.Tenant.Code != "acme-inc" || second.Tenant.Code == first.Tenant.Code {
		t.Fatalf("tenant codes = %q, %q", first.Tenant.Code, second.Tenant.Code)
	}
	ctx := withUserContext(context.Background(), first.Tenant.ID, first.User.ID, "tenant_admin")
	if err := service.ClearDemoData(ctx); err != nil {
		t.Fatalf("clear demo data: %v", err)
	}
	var count int64
	if err := database.Model(&domain.Course{}).Where("tenant_id = ?", first.Tenant.ID).Count(&count).Error; err != nil {
		t.Fatalf("count courses: %v", err)
	}
	if count != 0 {
		t.Fatalf("course count after clear = %d", count)
	}
}

func TestTenantCodeSlug(t *testing.T) {
	cases := map[string]string{"Acme Inc": "acme-inc", "  ACME!! ": "acme", "测试": "t-"}
	for input, prefix := range cases {
		if got := tenantCodeSlug(input); len(got) == 0 || (prefix != "t-" && got != prefix) || (prefix == "t-" && len(got) < 3) {
			t.Fatalf("tenantCodeSlug(%q) = %q", input, got)
		}
	}
}

func withUserContext(ctx context.Context, tenantID, userID, role string) context.Context {
	return usercontext.WithUser(ctx, userID, tenantID, userID+"@example.com", role)
}
