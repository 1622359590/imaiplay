package service

import (
	"context"
	"errors"
	"reflect"
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

func TestAuthServiceRegisterRejectsWhenTenantEmployeeLimitReached(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	planRepo := repository.NewPlanRepository(database)
	ctx := context.Background()
	plan := &domain.Plan{Name: "Single employee", MaxUsers: 1, Status: 1}
	if err := planRepo.Create(ctx, plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "registration-limit", Name: "Registration Limit", Status: 1, PlanID: &plan.ID}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Create(ctx, &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     "existing@example.com", Password: "hash", Name: "Existing", Role: "learner", Status: 1,
	}); err != nil {
		t.Fatal(err)
	}
	auth := NewAuthService(userRepo, tenantRepo, "secret")
	auth.SetEmployeeCapacityChecker(NewUserService(userRepo, UserLimitRepositories{
		Tenants: tenantRepo,
		Plans:   planRepo,
	}))
	tenantCtx := tenantcontext.WithTenant(ctx, tenant.Code, tenantcontext.SourceHeaderCode)

	_, err := auth.Register(tenantCtx, "new@example.com", "password123", "New", "learner")
	if errorCode(err) != 40300 || err.Error() != "员工数已达套餐上限，请升级套餐" {
		t.Fatalf("Register() error = %#v", err)
	}
}

func TestAuthServiceBootstrapSuperadminAndLoginWithoutTenant(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	_ = database
	service := NewAuthService(userRepo, tenantRepo, "secret")

	user, pair, err := service.BootstrapSuperadmin(context.Background(), "root@example.com", "Root", "password123")
	if err != nil {
		t.Fatalf("BootstrapSuperadmin() error = %v", err)
	}
	if user.Role != "superadmin" || user.TenantID != "" || pair.AccessToken == "" {
		t.Fatalf("bootstrapped superadmin = %#v, tokens = %#v", user, pair)
	}
	var tenantID *string
	if err := database.Table("users").Select("tenant_id").Where("id = ?", user.ID).Scan(&tenantID).Error; err != nil {
		t.Fatalf("read superadmin tenant id: %v", err)
	}
	if tenantID != nil {
		t.Fatalf("superadmin tenant id = %q, want NULL", *tenantID)
	}

	ctx := tenantcontext.WithTenant(context.Background(), tenantcontext.UnknownTenant, tenantcontext.SourceUnknown)
	loggedIn, err := service.LoginWithRefresh(ctx, "root@example.com", "password123")
	if err != nil {
		t.Fatalf("superadmin LoginWithRefresh() error = %v", err)
	}
	if loggedIn.AccessToken == "" {
		t.Fatal("superadmin login returned an empty access token")
	}
}

func TestAuthServicePlatformLoginFindsUniqueTenantAdminByEmail(t *testing.T) {
	_, tenantRepo, userRepo := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	tenantCtx := tenantcontext.WithTenant(context.Background(), tenant.Code, tenantcontext.SourceHeaderCode)
	user, err := service.Register(
		tenantCtx, "admin@example.com", "password123", "Admin", "tenant_admin",
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	platformCtx := tenantcontext.WithTenant(context.Background(), tenantcontext.UnknownTenant, tenantcontext.SourceUnknown)
	pair, err := service.LoginWithRefresh(platformCtx, "ADMIN@example.com", "password123")
	if err != nil {
		t.Fatalf("LoginWithRefresh() error = %v", err)
	}
	claims, err := security.ValidateToken(pair.AccessToken, "secret")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != user.ID || claims.TenantID != tenant.ID || claims.Role != "tenant_admin" {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestAuthServicePlatformLoginFindsUniqueTenantAdminByPhone(t *testing.T) {
	_, tenantRepo, userRepo := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "phone-admin", Name: "Phone Admin", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewAuthService(userRepo, tenantRepo, "secret")
	tenantCtx := tenantcontext.WithTenant(context.Background(), tenant.Code, tenantcontext.SourceHeaderCode)
	user, err := service.RegisterWithPhone(
		tenantCtx, "phone-admin@example.com", "13800138000", "password123", "Phone Admin", "tenant_admin",
	)
	if err != nil {
		t.Fatalf("RegisterWithPhone() error = %v", err)
	}

	platformCtx := tenantcontext.WithTenant(context.Background(), tenantcontext.UnknownTenant, tenantcontext.SourceUnknown)
	pair, err := service.LoginWithRefresh(platformCtx, "13800138000", "password123")
	if err != nil {
		t.Fatalf("LoginWithRefresh() error = %v", err)
	}
	claims, err := security.ValidateToken(pair.AccessToken, "secret")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != user.ID || claims.TenantID != tenant.ID || claims.Role != "tenant_admin" {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestAuthServicePlatformLoginRequiresTenantCodeForDuplicateCredential(t *testing.T) {
	_, tenantRepo, userRepo := serviceRepositories(t)
	service := NewAuthService(userRepo, tenantRepo, "secret")
	for _, code := range []string{"tenant-one", "tenant-two"} {
		tenant := &domain.Tenant{Code: code, Name: code, Status: 1}
		if err := tenantRepo.Create(context.Background(), tenant); err != nil {
			t.Fatalf("create tenant %q: %v", code, err)
		}
		tenantCtx := tenantcontext.WithTenant(context.Background(), code, tenantcontext.SourceHeaderCode)
		if _, err := service.Register(
			tenantCtx, "shared-admin@example.com", "password123", code, "tenant_admin",
		); err != nil {
			t.Fatalf("register tenant admin %q: %v", code, err)
		}
	}

	platformCtx := tenantcontext.WithTenant(context.Background(), tenantcontext.UnknownTenant, tenantcontext.SourceUnknown)
	_, err := service.LoginWithRefresh(platformCtx, "shared-admin@example.com", "password123")
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40900 ||
		appErr.Message != "account_exists_multiple_tenants" {
		t.Fatalf("LoginWithRefresh() error = %#v", err)
	}
}

func TestAuthServiceBeginLoginFindsSingleLearnerAcrossTenants(t *testing.T) {
	service, _, tenants, _ := discoveryAuthService(t)
	tenant, user := registerDiscoveryUser(
		t,
		service,
		tenants,
		"acme",
		"learner@example.com",
		"password123",
		"learner",
	)

	outcome, err := service.BeginLogin(
		platformContext(),
		"LEARNER@example.com",
		"password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.RequiresTenantSelection || outcome.Pair == nil ||
		outcome.User.ID != user.ID || outcome.Tenant.Code != tenant.Code {
		t.Fatalf("outcome=%#v", outcome)
	}
	claims, err := security.ValidateToken(outcome.Pair.AccessToken, "secret")
	if err != nil || claims.TenantID != tenant.ID || claims.Role != "learner" {
		t.Fatalf("claims=%#v err=%v", claims, err)
	}
}

func TestAuthServiceBeginLoginUsesPasswordBeforeSelectingOrganization(t *testing.T) {
	service, database, tenants, _ := discoveryAuthService(t)
	acme, _ := registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "acme-password", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "bravo-password", "instructor",
	)

	outcome, err := service.BeginLogin(
		platformContext(),
		"shared@example.com",
		"acme-password",
	)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.RequiresTenantSelection || outcome.Tenant.Code != acme.Code {
		t.Fatalf("outcome=%#v", outcome)
	}
	var challengeCount int64
	if err := database.Model(&domain.LoginChallenge{}).
		Count(&challengeCount).Error; err != nil {
		t.Fatal(err)
	}
	if challengeCount != 0 {
		t.Fatalf("challenge count=%d want=0", challengeCount)
	}
}

func TestAuthServiceBeginLoginSelectsAndRefreshesOrganization(t *testing.T) {
	service, _, tenants, _ := discoveryAuthService(t)
	acme, _ := registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "same-password", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "same-password", "instructor",
	)

	outcome, err := service.BeginLogin(
		platformContext(),
		"shared@example.com",
		"same-password",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.RequiresTenantSelection || outcome.SelectionToken == "" ||
		len(outcome.Organizations) != 2 || outcome.Pair != nil {
		t.Fatalf("outcome=%#v", outcome)
	}
	selected, err := service.SelectTenant(
		platformContext(),
		outcome.SelectionToken,
		acme.Code,
	)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := security.ValidateToken(selected.Pair.AccessToken, "secret")
	if err != nil || claims.TenantID != acme.ID || claims.Role != "learner" {
		t.Fatalf("claims=%#v err=%v", claims, err)
	}
	refreshed, err := service.Refresh(
		context.Background(),
		selected.Pair.RefreshToken,
	)
	if err != nil {
		t.Fatal(err)
	}
	refreshClaims, err := security.ValidateToken(refreshed.AccessToken, "secret")
	if err != nil || refreshClaims.TenantID != acme.ID {
		t.Fatalf("refresh claims=%#v err=%v", refreshClaims, err)
	}
	if _, err := service.SelectTenant(
		platformContext(),
		outcome.SelectionToken,
		acme.Code,
	); errorCode(err) != 40100 {
		t.Fatalf("replay error=%#v", err)
	}
}

func TestAuthServiceSelectTenantRejectsUserDisabledAfterChallenge(t *testing.T) {
	service, _, tenants, users := discoveryAuthService(t)
	acme, acmeUser := registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "password123", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "password123", "instructor",
	)

	outcome, err := service.BeginLogin(platformContext(), "shared@example.com", "password123")
	if err != nil || !outcome.RequiresTenantSelection {
		t.Fatalf("BeginLogin() = %#v, %v", outcome, err)
	}
	acmeUser.Status = 0
	userCtx := tenantcontext.WithUser(
		context.Background(), acmeUser.ID, acme.ID, acmeUser.Email, acmeUser.Role,
	)
	if err := users.Update(userCtx, acmeUser); err != nil {
		t.Fatalf("disable user: %v", err)
	}

	if _, err := service.SelectTenant(platformContext(), outcome.SelectionToken, acme.Code); errorCode(err) != 40100 {
		t.Fatalf("SelectTenant(disabled user) error = %#v", err)
	}
}

func TestAuthServiceSelectTenantRejectsTenantSuspendedAfterChallenge(t *testing.T) {
	service, _, tenants, _ := discoveryAuthService(t)
	acme, _ := registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "password123", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "password123", "instructor",
	)

	outcome, err := service.BeginLogin(platformContext(), "shared@example.com", "password123")
	if err != nil || !outcome.RequiresTenantSelection {
		t.Fatalf("BeginLogin() = %#v, %v", outcome, err)
	}
	acme.LifecycleStatus = "suspended"
	if err := tenants.Update(context.Background(), acme); err != nil {
		t.Fatalf("suspend tenant: %v", err)
	}

	if _, err := service.SelectTenant(platformContext(), outcome.SelectionToken, acme.Code); errorCode(err) != 40100 {
		t.Fatalf("SelectTenant(suspended tenant) error = %#v", err)
	}
}

func TestAuthServiceSelectTenantRejectsExpiredTrialAfterChallenge(t *testing.T) {
	service, _, tenants, _ := discoveryAuthService(t)
	acme, _ := registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "password123", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "password123", "instructor",
	)

	outcome, err := service.BeginLogin(platformContext(), "shared@example.com", "password123")
	if err != nil || !outcome.RequiresTenantSelection {
		t.Fatalf("BeginLogin() = %#v, %v", outcome, err)
	}
	expiredAt := time.Now().UTC().Add(-time.Minute)
	acme.LifecycleStatus = "trial"
	acme.TrialEndsAt = &expiredAt
	if err := tenants.Update(context.Background(), acme); err != nil {
		t.Fatalf("expire tenant trial: %v", err)
	}

	if _, err := service.SelectTenant(platformContext(), outcome.SelectionToken, acme.Code); errorCode(err) != 40100 {
		t.Fatalf("SelectTenant(expired trial) error = %#v", err)
	}
}

func TestAuthServiceBeginLoginDoesNotRevealOrganizationsForWrongPassword(t *testing.T) {
	service, database, tenants, _ := discoveryAuthService(t)
	registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "password123", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "password123", "instructor",
	)

	outcome, err := service.BeginLogin(
		platformContext(),
		"shared@example.com",
		"wrong-password",
	)
	if errorCode(err) != 40100 || outcome != nil {
		t.Fatalf("outcome=%#v error=%#v", outcome, err)
	}
	var challengeCount int64
	if err := database.Model(&domain.LoginChallenge{}).
		Count(&challengeCount).Error; err != nil {
		t.Fatal(err)
	}
	if challengeCount != 0 {
		t.Fatalf("challenge count=%d want=0", challengeCount)
	}
}

func TestAuthServiceBeginLoginFiltersUnavailableMemberships(t *testing.T) {
	service, database, tenants, _ := discoveryAuthService(t)
	active, _ := registerDiscoveryUser(
		t, service, tenants, "active", "shared@example.com", "password123", "learner",
	)
	_, disabled := registerDiscoveryUser(
		t, service, tenants, "disabled", "shared@example.com", "password123", "learner",
	)
	if err := database.Model(&domain.User{}).
		Where("id = ?", disabled.ID).
		Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	suspended, _ := registerDiscoveryUser(
		t, service, tenants, "suspended", "shared@example.com", "password123", "instructor",
	)
	suspended.LifecycleStatus = "suspended"
	if err := tenants.Update(context.Background(), suspended); err != nil {
		t.Fatal(err)
	}

	outcome, err := service.BeginLogin(
		platformContext(),
		"shared@example.com",
		"password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.RequiresTenantSelection || outcome.Tenant.Code != active.Code {
		t.Fatalf("outcome=%#v", outcome)
	}
}

func TestAuthServiceBeginLoginKeepsExplicitPortalScope(t *testing.T) {
	service, _, tenants, _ := discoveryAuthService(t)
	acme, _ := registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "password123", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "password123", "learner",
	)
	ctx := tenantcontext.WithTenant(
		context.Background(),
		acme.Code,
		tenantcontext.SourceHeaderCode,
	)
	outcome, err := service.BeginLogin(ctx, "shared@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if outcome.RequiresTenantSelection || outcome.Tenant.Code != acme.Code {
		t.Fatalf("outcome=%#v", outcome)
	}
}

func TestAuthServiceBeginLoginKeepsPlatformSuperadmin(t *testing.T) {
	service, _, _, _ := discoveryAuthService(t)
	superadmin, _, err := service.BootstrapSuperadmin(
		context.Background(),
		"root@example.com",
		"Root",
		"password123",
	)
	if err != nil {
		t.Fatal(err)
	}

	outcome, err := service.BeginLogin(
		platformContext(),
		superadmin.Email,
		"password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.RequiresTenantSelection || outcome.User.Role != "superadmin" ||
		outcome.Tenant != nil || outcome.Pair == nil {
		t.Fatalf("outcome=%#v", outcome)
	}
}

func TestAuthServiceSelectTenantRejectsInvalidChallenges(t *testing.T) {
	service, database, tenants, _ := discoveryAuthService(t)
	registerDiscoveryUser(
		t, service, tenants, "acme", "shared@example.com", "password123", "learner",
	)
	registerDiscoveryUser(
		t, service, tenants, "bravo", "shared@example.com", "password123", "instructor",
	)

	if _, err := service.SelectTenant(
		platformContext(),
		"tampered",
		"acme",
	); errorCode(err) != 40100 {
		t.Fatalf("tampered error=%#v", err)
	}

	unlisted, err := service.BeginLogin(
		platformContext(),
		"shared@example.com",
		"password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SelectTenant(
		platformContext(),
		unlisted.SelectionToken,
		"charlie",
	); errorCode(err) != 40100 {
		t.Fatalf("unlisted error=%#v", err)
	}
	if _, err := service.SelectTenant(
		platformContext(),
		unlisted.SelectionToken,
		"acme",
	); errorCode(err) != 40100 {
		t.Fatalf("consumed unlisted token error=%#v", err)
	}

	expired, err := service.BeginLogin(
		platformContext(),
		"shared@example.com",
		"password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&domain.LoginChallenge{}).
		Where("token_hash = ?", security.HashRefreshToken(expired.SelectionToken)).
		Update("expires_at", time.Now().UTC().Add(-time.Second)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.SelectTenant(
		platformContext(),
		expired.SelectionToken,
		"acme",
	); errorCode(err) != 40100 {
		t.Fatalf("expired error=%#v", err)
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

func TestAuthServiceRefreshRejectsUnavailableTenantAfterLogin(t *testing.T) {
	cases := []struct {
		name            string
		status          int
		lifecycleStatus string
		trialEndsAt     *time.Time
	}{
		{name: "disabled active", status: 0, lifecycleStatus: "active"},
		{name: "suspended", status: 1, lifecycleStatus: "suspended"},
		{
			name:            "expired trial",
			status:          1,
			lifecycleStatus: "trial",
			trialEndsAt:     ptrTime(time.Now().UTC().Add(-time.Hour)),
		},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			database, tenantRepo, userRepo := serviceRepositories(t)
			refreshRepo := repository.NewRefreshTokenRepository(database)
			tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
			if err := tenantRepo.Create(context.Background(), tenant); err != nil {
				t.Fatalf("create tenant: %v", err)
			}
			service := NewAuthServiceWithRefreshTokens(userRepo, tenantRepo, refreshRepo, "secret")
			ctx := tenantcontext.WithTenant(context.Background(), tenant.Code, tenantcontext.SourceHeaderCode)
			if _, err := service.Register(ctx, "admin@example.com", "password123", "Admin", "tenant_admin"); err != nil {
				t.Fatalf("register: %v", err)
			}
			pair, err := service.LoginWithRefresh(ctx, "admin@example.com", "password123")
			if err != nil || pair.RefreshToken == "" {
				t.Fatalf("login tokens = %#v, %v", pair, err)
			}
			tenant.Status = item.status
			tenant.LifecycleStatus = item.lifecycleStatus
			tenant.TrialEndsAt = item.trialEndsAt
			if err := tenantRepo.Update(context.Background(), tenant); err != nil {
				t.Fatalf("make tenant unavailable: %v", err)
			}

			refreshed, err := service.Refresh(context.Background(), pair.RefreshToken)
			if errorCode(err) != 40100 || refreshed != nil {
				t.Fatalf("Refresh(unavailable tenant) = %#v, %#v", refreshed, err)
			}
		})
	}
}

func TestAuthServiceRefreshAllowsSuperadminWithoutTenant(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	service := NewAuthServiceWithRefreshTokens(
		userRepo, tenantRepo, repository.NewRefreshTokenRepository(database), "secret",
	)
	_, pair, err := service.BootstrapSuperadmin(
		context.Background(), "root@example.com", "Root", "password123",
	)
	if err != nil || pair.RefreshToken == "" {
		t.Fatalf("BootstrapSuperadmin() = %#v, %v", pair, err)
	}

	refreshed, err := service.Refresh(context.Background(), pair.RefreshToken)
	if err != nil || refreshed == nil || refreshed.AccessToken == "" {
		t.Fatalf("Refresh(superadmin) = %#v, %v", refreshed, err)
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
	if err := service.SendLoginCode(ctx, "13900139000"); err != nil {
		t.Fatalf("unknown phone login code = %v", err)
	}
	if err := service.SendLoginCode(ctx, "13800138000"); err != nil {
		t.Fatalf("login code = %v", err)
	}
	loginCode := capture.params["code"]
	if _, err := service.LoginWithCode(ctx, "13800138000", loginCode); err != nil {
		t.Fatalf("login with code = %v", err)
	}
	if _, err := service.LoginWithCode(ctx, "13800138000", loginCode); errorCode(err) != 40100 {
		t.Fatalf("reused login code = %#v", err)
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

func TestAuthServiceCurrentUserReturnsSafeExactProfile(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	_ = database
	tenant := &domain.Tenant{Code: "current", Name: "Current", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	phone := "13800138000"
	user := &domain.User{
		BaseModel: domain.BaseModel{ID: "current-user", TenantID: tenant.ID},
		Email:     "current@example.com", Phone: &phone, Password: "secret-hash",
		Name: "Current User", Role: "tenant_admin", Status: 1,
	}
	if err := users.Create(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	service := NewAuthService(users, tenants, "secret")
	ctx := tenantcontext.WithUser(
		context.Background(), user.ID, tenant.ID, "claim@example.com", "tenant_admin",
	)
	got, err := service.CurrentUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := AuthUser{
		ID: user.ID, TenantID: tenant.ID, Name: user.Name, Email: user.Email,
		Phone: &phone, Role: user.Role,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CurrentUser() = %#v, want %#v", got, want)
	}
	assertJSONKeys(t, AuthUser{}, []string{"id", "tenant_id", "name", "email", "phone", "role"})
}

func TestAuthServiceCurrentUserSupportsEmptyScopeSuperadmin(t *testing.T) {
	_, tenants, users := serviceRepositories(t)
	service := NewAuthService(users, tenants, "secret")
	user, _, err := service.BootstrapSuperadmin(
		context.Background(), "root-current@example.com", "Root Current", "password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.CurrentUser(tenantcontext.WithUser(
		context.Background(), user.ID, "", user.Email, "superadmin",
	))
	if err != nil || got.TenantID != "" || got.Role != "superadmin" {
		t.Fatalf("CurrentUser(superadmin) = %#v, %#v", got, err)
	}
}

func TestAuthServiceCurrentUserRejectsMissingDisabledAndMismatchedClaims(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	tenant := &domain.Tenant{Code: "claims", Name: "Claims", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	user := &domain.User{
		BaseModel: domain.BaseModel{ID: "claims-user", TenantID: tenant.ID},
		Email:     "claims@example.com", Password: "hash", Name: "Claims",
		Role: "instructor", Status: 1,
	}
	if err := users.Create(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	service := NewAuthService(users, tenants, "secret")

	if _, err := service.CurrentUser(context.Background()); errorCode(err) != 40100 {
		t.Fatalf("missing context error = %#v", err)
	}
	if _, err := service.CurrentUser(tenantcontext.WithUser(
		context.Background(), "missing", tenant.ID, "", "instructor",
	)); errorCode(err) != 40100 {
		t.Fatalf("missing user error = %#v", err)
	}
	if _, err := service.CurrentUser(tenantcontext.WithUser(
		context.Background(), user.ID, tenant.ID, "", "tenant_admin",
	)); errorCode(err) != 40300 {
		t.Fatalf("role mismatch error = %#v", err)
	}
	if _, err := service.CurrentUser(tenantcontext.WithUser(
		context.Background(), user.ID, "", "", "superadmin",
	)); errorCode(err) != 40300 {
		t.Fatalf("scope mismatch error = %#v", err)
	}
	if err := database.Model(&domain.User{}).Where("id = ?", user.ID).
		Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.CurrentUser(tenantcontext.WithUser(
		context.Background(), user.ID, tenant.ID, "", "instructor",
	)); errorCode(err) != 40300 {
		t.Fatalf("disabled user error = %#v", err)
	}
}

func TestAuthServicePasswordAndCodeLoginUseSameSafeUserPresenter(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	resetRepo := repository.NewPasswordResetRepository(database)
	tenant := &domain.Tenant{Code: "presenter", Name: "Presenter", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	service := NewAuthService(users, tenants, "secret")
	service.SetPasswordResetRepository(resetRepo)
	capture := &captureSMSSender{}
	service.SetSMSSender(capture)
	ctx := tenantcontext.WithTenant(context.Background(), tenant.Code, tenantcontext.SourceHeaderCode)
	if _, err := service.RegisterWithPhone(
		ctx, "presenter@example.com", "13800138009", "password123", "Presenter", "learner",
	); err != nil {
		t.Fatal(err)
	}
	passwordOutcome, err := service.BeginLogin(ctx, "presenter@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SendLoginCode(ctx, "13800138009"); err != nil {
		t.Fatal(err)
	}
	codeOutcome, err := service.LoginWithCode(ctx, "13800138009", capture.params["code"])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(passwordOutcome.User, codeOutcome.User) {
		t.Fatalf("password user=%#v code user=%#v", passwordOutcome.User, codeOutcome.User)
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

func discoveryAuthService(
	t *testing.T,
) (
	*AuthService,
	*gorm.DB,
	repository.TenantRepository,
	repository.UserRepository,
) {
	t.Helper()
	database, tenants, users := serviceRepositories(t)
	service := NewAuthServiceWithRefreshTokens(
		users,
		tenants,
		repository.NewRefreshTokenRepository(database),
		"secret",
	)
	service.SetLoginChallengeRepository(
		repository.NewLoginChallengeRepository(database),
	)
	return service, database, tenants, users
}

func registerDiscoveryUser(
	t *testing.T,
	service *AuthService,
	tenants repository.TenantRepository,
	code, email, password, role string,
) (*domain.Tenant, *domain.User) {
	t.Helper()
	tenant := &domain.Tenant{Code: code, Name: code, Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant %q: %v", code, err)
	}
	ctx := tenantcontext.WithTenant(
		context.Background(),
		code,
		tenantcontext.SourceHeaderCode,
	)
	user, err := service.Register(
		ctx,
		email,
		password,
		code,
		role,
	)
	if err != nil {
		t.Fatalf("register user for %q: %v", code, err)
	}
	return tenant, user
}

func platformContext() context.Context {
	return tenantcontext.WithTenant(
		context.Background(),
		tenantcontext.UnknownTenant,
		tenantcontext.SourceUnknown,
	)
}

func errorCode(err error) int {
	var appErr *errorsx.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return 0
}
