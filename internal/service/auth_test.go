package service

import (
	"context"
	"errors"
	"testing"
	"time"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type captureSMSSender struct{ params map[string]string }

func (sender *captureSMSSender) Send(_ context.Context, _ string, _ string, params map[string]string) error {
	sender.params = params
	return nil
}

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

func TestAuthServiceRefreshRotationAndLogout(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	if err := database.AutoMigrate(&domain.RefreshToken{}); err != nil {
		t.Fatalf("migrate refresh tokens: %v", err)
	}
	refreshRepo := repository.NewRefreshTokenRepository(database)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewAuthServiceWithRefreshTokens(userRepo, tenantRepo, refreshRepo, "secret")
	ctx := tenantcontext.WithTenant(context.Background(), "acme", "header_code")
	if _, err := service.Register(ctx, "admin@example.com", "password123", "Admin", "tenant_admin"); err != nil {
		t.Fatalf("register: %v", err)
	}
	first, err := service.LoginWithRefresh(ctx, "admin@example.com", "password123")
	if err != nil || first.RefreshToken == "" {
		t.Fatalf("login tokens = %#v, %v", first, err)
	}
	stored, findErr := refreshRepo.FindValidByHash(ctx, security.HashRefreshToken(first.RefreshToken))
	if findErr != nil || stored.UserID == "" || stored.TenantID != tenant.ID {
		t.Fatalf("stored refresh token = %#v, %v", stored, findErr)
	}
	second, err := service.Refresh(ctx, first.RefreshToken)
	if err != nil || second.RefreshToken == "" || second.RefreshToken == first.RefreshToken {
		t.Fatalf("refresh = %#v, %v", second, err)
	}
	if _, err := service.Refresh(ctx, first.RefreshToken); errorCode(err) != 40100 {
		t.Fatalf("reused refresh error = %#v", err)
	}
	user, err := userRepo.FindByEmailAndTenant(ctx, "admin@example.com", tenant.ID)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	authCtx := tenantcontext.WithUser(ctx, user.ID, tenant.ID, user.Email, user.Role)
	if err := service.Logout(authCtx, second.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.Refresh(ctx, second.RefreshToken); errorCode(err) != 40100 {
		t.Fatalf("revoked refresh error = %#v", err)
	}
}

func TestAuthServicePhoneLoginAndPasswordReset(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	refreshRepo := repository.NewRefreshTokenRepository(database)
	resetRepo := repository.NewPasswordResetRepository(database)
	tenant := &domain.Tenant{Code: "phone", Name: "Phone", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	service := NewAuthServiceWithRefreshTokens(userRepo, tenantRepo, refreshRepo, "secret")
	service.SetPasswordResetRepository(resetRepo)
	capture := &captureSMSSender{}
	service.SetSMSSender(capture)
	ctx := tenantcontext.WithTenant(context.Background(), "phone", "test")
	if _, err := service.RegisterWithPhone(ctx, "phone@example.com", "13800138000", "oldpass123", "Phone", "learner"); err != nil {
		t.Fatal(err)
	}
	first, err := service.LoginWithRefresh(ctx, "13800138000", "oldpass123")
	if err != nil || first.RefreshToken == "" {
		t.Fatalf("phone login = %#v, %v", first, err)
	}
	if err := service.ForgotPassword(ctx, "13800138000"); err != nil {
		t.Fatalf("forgot = %v", err)
	}
	code := capture.params["code"]
	if len(code) != 6 {
		t.Fatalf("code = %q", code)
	}
	if err := service.ResetPassword(ctx, "13800138000", code, "newpass123"); err != nil {
		t.Fatalf("reset = %v", err)
	}
	if _, err := service.LoginWithRefresh(ctx, "13800138000", "oldpass123"); errorCode(err) != 40100 {
		t.Fatalf("old password error = %#v", err)
	}
	second, err := service.LoginWithRefresh(ctx, "13800138000", "newpass123")
	if err != nil || second.RefreshToken == "" {
		t.Fatalf("new password login = %#v, %v", second, err)
	}
	if _, err := service.Refresh(ctx, first.RefreshToken); errorCode(err) != 40100 {
		t.Fatalf("old refresh error = %#v", err)
	}
	if err := service.ForgotPassword(ctx, "13900139000"); err != nil {
		t.Fatalf("unknown phone forgot = %v", err)
	}
}

func TestAuthServicePasswordResetGuards(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	resetRepo := repository.NewPasswordResetRepository(database)
	tenant := &domain.Tenant{Code: "guards", Name: "Guards", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	service.SetPasswordResetRepository(resetRepo)
	capture := &captureSMSSender{}
	service.SetSMSSender(capture)
	ctx := tenantcontext.WithTenant(context.Background(), "guards", "test")
	if _, err := service.RegisterWithPhone(ctx, "guards@example.com", "13800138001", "oldpass123", "Guards", "learner"); err != nil {
		t.Fatal(err)
	}
	if err := service.ForgotPassword(ctx, "13800138001"); err != nil {
		t.Fatal(err)
	}
	if err := service.ForgotPassword(ctx, "13800138001"); errorCode(err) != 40900 {
		t.Fatalf("repeat forgot error = %#v", err)
	}
	for attempt := 0; attempt < 5; attempt++ {
		if errorCode(service.ResetPassword(ctx, "13800138001", "000000", "newpass123")) != 40000 {
			t.Fatalf("wrong attempt %d was accepted", attempt)
		}
	}
	if errorCode(service.ResetPassword(ctx, "13800138001", "000000", "newpass123")) != 40000 {
		t.Fatal("sixth attempt did not fail")
	}
	if err := resetRepo.Create(ctx, &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Phone: "13800138001", CodeHash: hashVerificationCode("123456"), ExpiresAt: time.Now().Add(-time.Minute)}); err != nil {
		t.Fatal(err)
	}
	if errorCode(service.ResetPassword(ctx, "13800138001", "123456", "newpass123")) != 40000 {
		t.Fatal("expired code was accepted")
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
