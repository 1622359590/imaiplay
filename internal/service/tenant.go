package service

import (
	"context"
	"errors"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type TenantService struct {
	tenants repository.TenantRepository
}

func NewTenantService(tenants repository.TenantRepository) *TenantService {
	return &TenantService{tenants: tenants}
}

func (service *TenantService) Create(
	ctx context.Context,
	code, name string,
) (*domain.Tenant, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	if _, err := service.tenants.FindByCode(ctx, code); err == nil {
		return nil, errorsx.Conflict("tenant code already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find tenant failed")
	}
	tenant := &domain.Tenant{Code: code, Name: name, Status: 1}
	if err := service.tenants.Create(ctx, tenant); err != nil {
		return nil, mapCreateError(
			err, "tenant code already exists", "create tenant failed",
		)
	}
	return tenant, nil
}

func (service *TenantService) List(
	ctx context.Context,
	offset, limit int,
) ([]domain.Tenant, int64, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, 0, err
	}
	items, err := service.tenants.FindAll(ctx)
	if err != nil {
		return nil, 0, errorsx.Internal("list tenants failed")
	}
	total := int64(len(items))
	offset, limit = pagination(offset, limit, len(items))
	return items[offset:limit], total, nil
}

func (service *TenantService) Get(
	ctx context.Context,
	id string,
) (*domain.Tenant, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	tenant, err := service.tenants.FindByID(ctx, id)
	return tenant, mapNotFound(err, "tenant not found")
}

func (service *TenantService) Update(
	ctx context.Context,
	id, name string,
	status int,
) (*domain.Tenant, error) {
	tenant, err := service.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	tenant.Name, tenant.Status = name, status
	if err := service.tenants.Update(ctx, tenant); err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	return tenant, nil
}

func (service *TenantService) Delete(ctx context.Context, id string) error {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return err
	}
	err := service.tenants.Delete(ctx, id)
	if err != nil && (strings.Contains(strings.ToLower(err.Error()), "foreign key") || strings.Contains(strings.ToLower(err.Error()), "constraint")) {
		return errorsx.Conflict("tenant has users and cannot be deleted")
	}
	return mapNotFound(err, "tenant not found")
}

func requireRole(ctx context.Context, required string) error {
	_, _, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != required {
		return errorsx.Forbidden("permission denied")
	}
	return nil
}

func pagination(offset, limit, length int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset > length {
		offset = length
	}
	end := offset + limit
	if end > length {
		end = length
	}
	return offset, end
}

func mapNotFound(err error, message string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errorsx.NotFound(message)
	}
	if err != nil {
		return errorsx.Internal("database operation failed")
	}
	return nil
}

func mapCreateError(err error, conflictMessage, internalMessage string) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return errorsx.Conflict(conflictMessage)
	}
	return errorsx.Internal(internalMessage)
}
