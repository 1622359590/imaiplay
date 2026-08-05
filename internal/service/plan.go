package service

import (
	"context"
	"errors"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type PlanService struct {
	plans     repository.PlanRepository
	tenants   repository.TenantRepository
	resources repository.ResourceRepository
}

func NewPlanService(plans repository.PlanRepository, tenants repository.TenantRepository, resources repository.ResourceRepository) *PlanService {
	return &PlanService{plans: plans, tenants: tenants, resources: resources}
}

func (service *PlanService) List(ctx context.Context, offset, limit int) ([]domain.Plan, int64, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, 0, err
	}
	return service.plans.List(ctx, offset, limit)
}

func (service *PlanService) Create(ctx context.Context, plan *domain.Plan) (*domain.Plan, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	if plan.Name == "" {
		return nil, errorsx.BadRequest("plan name is required")
	}
	if plan.StorageQuotaBytes < 0 || plan.MaxUsers < 0 || plan.MaxCourses < 0 {
		return nil, errorsx.BadRequest("plan quotas cannot be negative")
	}
	if plan.Features == "" {
		plan.Features = "{}"
	}
	if err := service.plans.Create(ctx, plan); err != nil {
		return nil, errorsx.Internal("create plan failed")
	}
	return plan, nil
}

func (service *PlanService) Update(ctx context.Context, plan *domain.Plan) (*domain.Plan, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	if plan.Name == "" || plan.StorageQuotaBytes < 0 || plan.MaxUsers < 0 || plan.MaxCourses < 0 {
		return nil, errorsx.BadRequest("invalid plan")
	}
	existing, err := service.plans.FindByID(ctx, plan.ID)
	if err != nil {
		return nil, mapNotFound(err, "plan not found")
	}
	if existing.IsDefault {
		plan.IsDefault = true
		if plan.Status != 1 {
			return nil, errorsx.BadRequest("default plan must remain enabled")
		}
	}
	if plan.Features == "" {
		plan.Features = existing.Features
	}
	if err := service.plans.Update(ctx, plan); err != nil {
		return nil, mapNotFound(err, "plan not found")
	}
	return plan, nil
}

func (service *PlanService) Delete(ctx context.Context, id string) error {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return err
	}
	plan, err := service.plans.FindByID(ctx, id)
	if err != nil {
		return mapNotFound(err, "plan not found")
	}
	if plan.IsDefault {
		return errorsx.BadRequest("default plan cannot be deleted")
	}
	return mapNotFound(service.plans.Delete(ctx, id), "plan not found")
}

func (service *PlanService) Assign(ctx context.Context, tenantID, planID string) (*domain.Tenant, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	if _, err := service.plans.FindByID(ctx, planID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.NotFound("plan not found")
		}
		return nil, errorsx.Internal("find plan failed")
	}
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	tenant.PlanID = &planID
	if err := service.tenants.UpdatePlan(ctx, tenant); err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	return tenant, nil
}

func (service *PlanService) Current(ctx context.Context) (map[string]interface{}, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return nil, errorsx.Forbidden("permission denied")
	}
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	var plan *domain.Plan
	if tenant.PlanID != nil {
		plan, err = service.plans.FindByID(ctx, *tenant.PlanID)
	} else {
		plan, err = service.plans.FindDefault(ctx)
	}
	if err != nil {
		return nil, errorsx.Internal("load current plan failed")
	}
	used, err := service.resources.TotalSizeByTenant(ctx, tenantID)
	if err != nil {
		return nil, errorsx.Internal("load storage usage failed")
	}
	return map[string]interface{}{"plan": plan, "used_bytes": used, "quota_bytes": plan.StorageQuotaBytes}, nil
}

func (service *PlanService) CheckStorage(ctx context.Context, tenantID string, incoming int64) error {
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return errorsx.Internal("load tenant plan failed")
	}
	var plan *domain.Plan
	if tenant.PlanID != nil {
		plan, err = service.plans.FindByID(ctx, *tenant.PlanID)
	} else {
		plan, err = service.plans.FindDefault(ctx)
	}
	if err != nil {
		return errorsx.Internal("load tenant plan failed")
	}
	if plan.StorageQuotaBytes <= 0 {
		return nil
	}
	used, err := service.resources.TotalSizeByTenant(ctx, tenantID)
	if err != nil {
		return errorsx.Internal("load storage usage failed")
	}
	if used+incoming > plan.StorageQuotaBytes {
		return errorsx.BadRequest("存储配额不足，请升级套餐")
	}
	return nil
}
