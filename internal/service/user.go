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
	users   repository.UserRepository
	tenants repository.TenantRepository
	plans   repository.PlanRepository
	limits  *TenantLimitService
}

type EmployeeCapacityChecker interface {
	WithEmployeeSlot(ctx context.Context, tenantID string, create func() error) error
}

type UserLimitRepositories struct {
	Tenants repository.TenantRepository
	Plans   repository.PlanRepository
}

func NewUserService(users repository.UserRepository, limits ...UserLimitRepositories) *UserService {
	service := &UserService{users: users}
	if len(limits) > 0 {
		service.tenants = limits[0].Tenants
		service.plans = limits[0].Plans
		service.limits = NewTenantLimitService(
			limits[0].Tenants, limits[0].Plans, users, nil,
		)
	}
	return service
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
	var user *domain.User
	err = service.WithEmployeeSlot(ctx, tenantID, func() error {
		var createErr error
		user, createErr = service.createWithPhone(ctx, tenantID, email, phone, password, name, role)
		return createErr
	})
	return user, err
}

func (service *UserService) createWithPhone(
	ctx context.Context, tenantID, email, phone, password, name, role string,
) (*domain.User, error) {
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

func (service *UserService) WithEmployeeSlot(
	ctx context.Context, tenantID string, create func() error,
) error {
	if service.limits == nil {
		return create()
	}
	return service.limits.WithEmployeeSlot(ctx, tenantID, create)
}

func (service *UserService) EnsureEmployeeCapacity(ctx context.Context, tenantID string) error {
	return service.WithEmployeeSlot(ctx, tenantID, func() error { return nil })
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
	id, name string, status int, password string,
) (*domain.User, error) {
	user, err := service.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Name, user.Status = name, status
	if password != "" {
		if len(password) < 8 {
			return nil, errorsx.BadRequest("password must be at least 8 characters")
		}
		hash, err := security.HashPassword(password)
		if err != nil {
			return nil, errorsx.Internal("hash password failed")
		}
		user.Password = hash
	}
	if err := service.users.Update(ctx, user); err != nil {
		return nil, mapNotFound(err, "user not found")
	}
	return user, nil
}

func (service *UserService) Delete(ctx context.Context, id string) error {
	if _, err := tenantAdminID(ctx); err != nil {
		return err
	}
	return mapNotFound(service.users.Deactivate(ctx, id), "user not found")
}

func (service *UserService) ResetTenantAdminPassword(ctx context.Context, id, password string) error {
	_, _, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "superadmin" {
		return errorsx.Forbidden("permission denied")
	}
	if len(password) < 8 {
		return errorsx.BadRequest("password must be at least 8 characters")
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return errorsx.Internal("hash password failed")
	}
	return mapNotFound(service.users.UpdatePassword(ctx, id, hash), "tenant admin not found")
}

func tenantAdminID(ctx context.Context) (string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return "", errorsx.Forbidden("permission denied")
	}
	return tenantID, nil
}
