package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type TenantLimitService struct {
	tenants repository.TenantRepository
	plans   repository.PlanRepository
	users   repository.UserRepository
	courses repository.CourseRepository
	locks   sync.Map
}

func NewTenantLimitService(
	tenants repository.TenantRepository,
	plans repository.PlanRepository,
	users repository.UserRepository,
	courses repository.CourseRepository,
) *TenantLimitService {
	return &TenantLimitService{tenants: tenants, plans: plans, users: users, courses: courses}
}

func (service *TenantLimitService) ResolvePlan(
	ctx context.Context, tenantID string,
) (*domain.Plan, error) {
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, errorsx.Internal("find tenant failed")
	}
	if tenant.PlanID != nil && strings.TrimSpace(*tenant.PlanID) != "" {
		plan, err := service.plans.FindByID(ctx, *tenant.PlanID)
		if err == nil {
			return plan, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.Internal("find plan failed")
		}
	}
	plan, err := service.plans.FindDefault(ctx)
	if err != nil {
		return nil, errorsx.Internal("find default plan failed")
	}
	return plan, nil
}

func (service *TenantLimitService) WithEmployeeSlot(
	ctx context.Context, tenantID string, create func() error,
) error {
	return service.withTenantLock(tenantID, func() error {
		plan, err := service.ResolvePlan(ctx, tenantID)
		if err != nil {
			return err
		}
		if plan.MaxUsers > 0 {
			users, _, err := service.users.FindByTenant(ctx, tenantID, 0, 100000)
			if err != nil {
				return errorsx.Internal("count users failed")
			}
			active := 0
			for _, user := range users {
				if user.Status == 1 {
					active++
				}
			}
			if active >= plan.MaxUsers {
				return errorsx.Forbidden("员工数已达套餐上限，请升级套餐")
			}
		}
		return create()
	})
}

func (service *TenantLimitService) WithCourseSlot(
	ctx context.Context, tenantID string, create func() error,
) error {
	return service.withTenantLock(tenantID, func() error {
		plan, err := service.ResolvePlan(ctx, tenantID)
		if err != nil {
			return err
		}
		if plan.MaxCourses > 0 {
			_, total, err := service.courses.FindByTenant(ctx, tenantID, 0, 1)
			if err != nil {
				return errorsx.Internal("count courses failed")
			}
			if total >= int64(plan.MaxCourses) {
				return errorsx.Forbidden("课程数已达套餐上限，请升级套餐")
			}
		}
		return create()
	})
}

func (service *TenantLimitService) withTenantLock(tenantID string, fn func() error) error {
	value, _ := service.locks.LoadOrStore(tenantID, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	return fn()
}
