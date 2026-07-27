package service

import (
	"context"
	"errors"
	"strings"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

type AuthService struct {
	users     repository.UserRepository
	tenants   repository.TenantRepository
	jwtSecret string
}

func NewAuthService(
	users repository.UserRepository,
	tenants repository.TenantRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{users: users, tenants: tenants, jwtSecret: jwtSecret}
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
	token, err := security.GenerateToken(
		user.ID, user.TenantID, user.Email, user.Role, service.jwtSecret,
	)
	if err != nil {
		return "", errorsx.Internal("generate token failed")
	}
	return token, nil
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
