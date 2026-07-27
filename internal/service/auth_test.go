package service

import (
	"context"
	"errors"
	"testing"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	_ = database
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	ctx := tenantcontext.WithTenant(context.Background(), "acme", "header_code")

	user, err := service.Register(
		ctx, "admin@example.com", "password123", "Admin", "tenant_admin",
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.TenantID != tenant.ID || user.Password == "password123" ||
		!security.CheckPassword("password123", user.Password) {
		t.Fatalf("registered user = %#v", user)
	}
	token, err := service.Login(ctx, "admin@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	claims, err := security.ValidateToken(token, "secret")
	if err != nil || claims.UserID != user.ID || claims.TenantID != tenant.ID {
		t.Fatalf("ValidateToken() = %#v, %v", claims, err)
	}
}

func TestAuthServiceRejectsSuperadminPublicRegistration(t *testing.T) {
	_, tenantRepo, userRepo := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	ctx := tenantcontext.WithTenant(context.Background(), "acme", "header_code")

	_, err := service.Register(
		ctx, "root@example.com", "password123", "Root", "superadmin",
	)
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40000 ||
		appErr.Message != "superadmin 不可通過公開註冊創建" {
		t.Fatalf("Register(superadmin) error = %#v", err)
	}
}

func TestAuthServiceRejectsDuplicateAndInvalidLogin(t *testing.T) {
	_, tenantRepo, userRepo := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	ctx := tenantcontext.WithTenant(context.Background(), "acme", "header_code")
	if _, err := service.Register(
		ctx, "admin@example.com", "password123", "Admin", "tenant_admin",
	); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := service.Register(
		ctx, "admin@example.com", "password123", "Other", "learner",
	); errorCode(err) != 40900 {
		t.Fatalf("duplicate Register() error = %#v", err)
	}
	if _, err := service.Login(ctx, "admin@example.com", "wrong"); errorCode(err) != 40100 {
		t.Fatalf("Login(wrong) error = %#v", err)
	}
}

func TestAuthServiceReportsUserDatabaseFailure(t *testing.T) {
	_, tenantRepo, _ := serviceRepositories(t)
	userDatabase, _, userRepo := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	sqlDB, err := userDatabase.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close user database: %v", err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	ctx := tenantcontext.WithTenant(context.Background(), "acme", "header_code")

	if _, err := service.Login(ctx, "admin@example.com", "password123"); errorCode(err) != 50000 {
		t.Fatalf("Login(database failure) error = %#v", err)
	}
}

func serviceRepositories(
	t *testing.T,
) (*gorm.DB, repository.TenantRepository, repository.UserRepository) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return database,
		repository.NewTenantRepository(database),
		repository.NewUserRepository(database)
}

func errorCode(err error) int {
	var appErr *errorsx.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return 0
}
