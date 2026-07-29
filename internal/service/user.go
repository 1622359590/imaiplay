package service

import (
	"context"
	"errors"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (service *UserService) Create(
	ctx context.Context,
	email, password, name, role string,
) (*domain.User, error) {
	return service.CreateWithPhone(ctx, email, "", password, name, role)
}

func (service *UserService) CreateWithPhone(ctx context.Context, email, phone, password, name, role string) (*domain.User, error) {
	tenantID, err := tenantAdminID(ctx)
	if err != nil {
		return nil, err
	}
	if !isTenantRole(role) {
		return nil, errorsx.BadRequest("invalid role")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := service.users.FindByEmailAndTenant(ctx, email, tenantID); err == nil {
		return nil, errorsx.Conflict("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find user failed")
	}
	phone = normalizePhone(phone)
	if phone != "" {
		if !validPhone(phone) {
			return nil, errorsx.BadRequest("invalid phone")
		}
		if _, err := service.users.FindByPhoneAndTenant(ctx, phone, tenantID); err == nil {
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
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Email:     email, Phone: nullablePhone(phone), Password: hash, Name: name, Role: role, Status: 1,
	}
	if err := service.users.Create(ctx, user); err != nil {
		return nil, mapCreateError(err, "email already exists", "create user failed")
	}
	return user, nil
}

func (service *UserService) List(
	ctx context.Context,
	offset, limit int,
) ([]domain.User, int64, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (role != "tenant_admin" && role != "superadmin") {
		return nil, 0, errorsx.Forbidden("permission denied")
	}
	var users []domain.User
	var total int64
	var err error
	if role == "superadmin" {
		users, total, err = service.users.FindAll(ctx, offset, limit)
	} else {
		if tenantID == "" {
			return nil, 0, errorsx.Forbidden("permission denied")
		}
		users, total, err = service.users.FindByTenant(ctx, tenantID, offset, limit)
	}
	if err != nil {
		return nil, 0, errorsx.Internal("list users failed")
	}
	return users, total, nil
}

func (service *UserService) Get(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	if _, err := tenantAdminID(ctx); err != nil {
		return nil, err
	}
	user, err := service.users.FindByID(ctx, id)
	return user, mapNotFound(err, "user not found")
}

func (service *UserService) Update(
	ctx context.Context,
	id, name string,
	status int,
) (*domain.User, error) {
	user, err := service.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Name, user.Status = name, status
	if err := service.users.Update(ctx, user); err != nil {
		return nil, mapNotFound(err, "user not found")
	}
	return user, nil
}

func (service *UserService) Delete(ctx context.Context, id string) error {
	if _, err := tenantAdminID(ctx); err != nil {
		return err
	}
	return mapNotFound(service.users.Delete(ctx, id), "user not found")
}

func tenantAdminID(ctx context.Context) (string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return "", errorsx.Forbidden("permission denied")
	}
	return tenantID, nil
}
