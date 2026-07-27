package service

import (
	"context"
	"errors"
	"strings"
	"time"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

type AuthService struct {
	users         repository.UserRepository
	tenants       repository.TenantRepository
	jwtSecret     string
	refreshTokens repository.RefreshTokenRepository
}

func NewAuthService(
	users repository.UserRepository,
	tenants repository.TenantRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{users: users, tenants: tenants, jwtSecret: jwtSecret}
}

func NewAuthServiceWithRefreshTokens(users repository.UserRepository, tenants repository.TenantRepository, refreshTokens repository.RefreshTokenRepository, jwtSecret string) *AuthService {
	service := NewAuthService(users, tenants, jwtSecret)
	service.refreshTokens = refreshTokens
	return service
}

type TokenPair struct {
	AccessToken  string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (service *AuthService) Register(
	ctx context.Context,
	email, password, name, role string,
) (*domain.User, error) {
	if role == "superadmin" {
		return nil, errorsx.BadRequest("superadmin 不可通過公開註冊創建")
	}
	if !isTenantRole(role) {
		return nil, errorsx.BadRequest("invalid role")
	}
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return nil, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := service.users.FindByEmailAndTenant(ctx, email, tenant.ID); err == nil {
		return nil, errorsx.Conflict("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find user failed")
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, errorsx.Internal("hash password failed")
	}
	user := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     email, Password: hash, Name: name, Role: role, Status: 1,
	}
	if err := service.users.Create(ctx, user); err != nil {
		return nil, mapCreateError(err, "email already exists", "create user failed")
	}
	return user, nil
}

func (service *AuthService) BootstrapSuperadmin(ctx context.Context, email, name, password string) (*domain.User, *TokenPair, error) {
	users, _, err := service.users.FindByTenant(ctx, "", 0, 100000)
	if err != nil {
		return nil, nil, errorsx.Internal("find superadmin failed")
	}
	for _, user := range users {
		if user.Role == "superadmin" {
			return nil, nil, errorsx.Conflict("superadmin already initialized")
		}
	}
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" || name == "" || len(password) < 8 {
		return nil, nil, errorsx.BadRequest("email, name and password are required")
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, nil, errorsx.Internal("hash password failed")
	}
	user := &domain.User{BaseModel: domain.BaseModel{TenantID: ""}, Email: email, Password: hash, Name: name, Role: "superadmin", Status: 1}
	if err := service.users.Create(ctx, user); err != nil {
		return nil, nil, errorsx.Internal("create superadmin failed")
	}
	pair, err := service.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (service *AuthService) Login(
	ctx context.Context,
	email, password string,
) (string, error) {
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return "", err
	}
	user, err := service.users.FindByEmailAndTenant(
		ctx, strings.ToLower(strings.TrimSpace(email)), tenant.ID,
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errorsx.Unauthorized("invalid email or password")
	}
	if err != nil {
		return "", errorsx.Internal("find user failed")
	}
	if !security.CheckPassword(password, user.Password) {
		return "", errorsx.Unauthorized("invalid email or password")
	}
	if user.Status != 1 {
		return "", errorsx.Forbidden("user is disabled")
	}
	pair, err := service.issueTokens(ctx, user)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}

func (service *AuthService) LoginWithRefresh(ctx context.Context, email, password string) (*TokenPair, error) {
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.users.FindByEmailAndTenant(ctx, strings.ToLower(strings.TrimSpace(email)), tenant.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) || err == nil && !security.CheckPassword(password, user.Password) {
		return nil, errorsx.Unauthorized("invalid email or password")
	}
	if err != nil {
		return nil, errorsx.Internal("find user failed")
	}
	if user.Status != 1 {
		return nil, errorsx.Forbidden("user is disabled")
	}
	return service.issueTokens(ctx, user)
}

func (service *AuthService) IssueTokens(ctx context.Context, user *domain.User) (*TokenPair, error) {
	return service.issueTokens(ctx, user)
}

func (service *AuthService) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	if service.refreshTokens == nil || raw == "" {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	token, err := service.refreshTokens.FindValidByHash(ctx, security.HashRefreshToken(raw))
	if err != nil || token.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	userCtx := tenantcontext.WithUser(ctx, token.UserID, token.TenantID, "", "")
	user, err := service.users.FindByID(userCtx, token.UserID)
	if err != nil || user.Status != 1 {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	if err := service.refreshTokens.Revoke(ctx, token.TokenHash); err != nil {
		return nil, errorsx.Internal("revoke refresh token failed")
	}
	return service.issueTokens(userCtx, user)
}

func (service *AuthService) Logout(ctx context.Context, raw string) error {
	if service.refreshTokens == nil {
		return nil
	}
	userID, _, _, _, ok := tenantcontext.UserFromContext(ctx)
	if !ok {
		return errorsx.Unauthorized("authentication required")
	}
	if raw != "" {
		token, err := service.refreshTokens.FindValidByHash(ctx, security.HashRefreshToken(raw))
		if err != nil || token.UserID != userID {
			return errorsx.Unauthorized("invalid refresh token")
		}
		if err := service.refreshTokens.Revoke(ctx, token.TokenHash); err != nil {
			return errorsx.Unauthorized("invalid refresh token")
		}
		return nil
	}
	return service.refreshTokens.RevokeAllForUser(ctx, userID)
}

func (service *AuthService) issueTokens(ctx context.Context, user *domain.User) (*TokenPair, error) {
	access, err := security.GenerateToken(user.ID, user.TenantID, user.Email, user.Role, service.jwtSecret)
	if err != nil {
		return nil, errorsx.Internal("generate token failed")
	}
	pair := &TokenPair{AccessToken: access, ExpiresAt: time.Now().UTC().Add(24 * time.Hour)}
	if service.refreshTokens == nil {
		return pair, nil
	}
	plain, hash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, errorsx.Internal("generate refresh token failed")
	}
	if err := service.refreshTokens.Create(ctx, &domain.RefreshToken{BaseModel: domain.BaseModel{TenantID: user.TenantID}, UserID: user.ID, TokenHash: hash, ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour)}); err != nil {
		return nil, errorsx.Internal("store refresh token failed")
	}
	pair.RefreshToken = plain
	return pair, nil
}

func (service *AuthService) currentTenant(
	ctx context.Context,
) (*domain.Tenant, error) {
	code, _ := tenantcontext.TenantFromContext(ctx)
	tenant, err := service.tenants.FindByCode(ctx, code)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.NotFound("tenant not found")
	}
	if err != nil {
		return nil, errorsx.Internal("find tenant failed")
	}
	if tenant.Status != 1 {
		return nil, errorsx.Forbidden("tenant is disabled")
	}
	return tenant, nil
}

func isTenantRole(role string) bool {
	return role == "tenant_admin" || role == "instructor" || role == "learner"
}
