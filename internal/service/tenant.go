package service

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

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

func (service *TenantService) UpdateLifecycle(ctx context.Context, id, status string, trialEndsAt *time.Time) (*domain.Tenant, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	if status != "trial" && status != "active" && status != "suspended" && status != "deleted" {
		return nil, errorsx.BadRequest("invalid tenant lifecycle status")
	}
	tenant, err := service.tenants.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	tenant.LifecycleStatus, tenant.TrialEndsAt = status, trialEndsAt
	if err := service.tenants.Update(ctx, tenant); err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	return tenant, nil
}

func TenantAccessible(tenant *domain.Tenant, now time.Time) (bool, string) {
	status := tenant.LifecycleStatus
	if status == "" {
		if tenant.Status == 0 {
			return false, "tenant is suspended"
		}
		return true, ""
	}
	if status == "suspended" || status == "deleted" {
		return false, "tenant is " + status
	}
	if status == "trial" && tenant.TrialEndsAt != nil && !now.Before(*tenant.TrialEndsAt) {
		return false, "tenant trial expired"
	}
	return true, ""
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

func (service *TenantService) SetCustomDomain(ctx context.Context, id, customDomain string) (*domain.Tenant, error) {
	_, currentTenant, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (role != "superadmin" && role != "tenant_admin") {
		return nil, errorsx.Forbidden("permission denied")
	}
	if role == "tenant_admin" {
		id = currentTenant
	}
	domainName := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(customDomain), "."))
	if domainName != "" && !validCustomDomain(domainName) {
		return nil, errorsx.BadRequest("invalid custom domain")
	}
	tenant, err := service.tenants.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	if domainName != "" {
		other, findErr := service.tenants.FindByCustomDomain(ctx, domainName)
		if findErr == nil && other.ID != tenant.ID {
			return nil, errorsx.Conflict("custom domain already exists")
		}
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, errorsx.Internal("find custom domain failed")
		}
		tenant.CustomDomain = &domainName
	} else {
		tenant.CustomDomain = nil
	}
	if err := service.tenants.Update(ctx, tenant); err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	return tenant, nil
}

func validCustomDomain(value string) bool {
	if len(value) > 253 || strings.Contains(value, "://") || net.ParseIP(value) != nil {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if !(char == '-' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9') {
				return false
			}
		}
	}
	return strings.Contains(value, ".")
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
