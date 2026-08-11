package service

import (
	"context"
	"errors"
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

func (service *AuthService) Register(
	ctx context.Context,
	email, password, name, role string,
) (*domain.User, error) {
	return service.RegisterWithPhone(ctx, email, "", password, name, role)
}

func (service *AuthService) RegisterWithPhone(ctx context.Context, email, phone, password, name, role string) (*domain.User, error) {
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
	if service.employeeCapacity != nil {
		if err := service.employeeCapacity.EnsureEmployeeCapacity(ctx, tenant.ID); err != nil {
			return nil, err
		}
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := service.users.FindByEmailAndTenant(ctx, email, tenant.ID); err == nil {
		return nil, errorsx.Conflict("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find user failed")
	}
	phone = normalizePhone(phone)
	if phone != "" {
		if !validPhone(phone) {
			return nil, errorsx.BadRequest("invalid phone")
		}
		if _, err := service.users.FindByPhoneAndTenant(ctx, phone, tenant.ID); err == nil {
			return nil, errorsx.Conflict("phone already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.Internal("find user failed")
		}
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, errorsx.Internal("hash password failed")
	}
	user := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     email, Phone: nullablePhone(phone), Password: hash, Name: name, Role: role, Status: 1,
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
