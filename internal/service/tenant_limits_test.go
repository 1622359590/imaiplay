package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestUserServiceUsesDefaultPlanAndIgnoresDisabledEmployees(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	plans := repository.NewPlanRepository(database)
	defaultPlan, err := plans.FindDefault(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defaultPlan.MaxUsers = 1
	if err := plans.Update(context.Background(), defaultPlan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "default-limit", Name: "Default Limit", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	disabled := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID}, Email: "disabled@example.com",
		Password: "hash", Name: "Disabled", Role: "learner", Status: 0,
	}
	if err := users.Create(context.Background(), disabled); err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&domain.User{}).Where("id = ?", disabled.ID).Update("status", 0).Error; err != nil {
		t.Fatal(err)
	}
	service := NewUserService(users, UserLimitRepositories{Tenants: tenants, Plans: plans})
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	if _, err := service.Create(ctx, "active@example.com", "password123", "Active", "learner"); err != nil {
		t.Fatalf("active employee should fit default plan: %v", err)
	}
	if _, err := service.Create(ctx, "second@example.com", "password123", "Second", "learner"); errorCode(err) != 40300 {
		t.Fatalf("second active employee error = %#v", err)
	}
}

func TestTenantLimitServiceSerializesEmployeeSlots(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	plans := repository.NewPlanRepository(database)
	plan := &domain.Plan{Name: "One", MaxUsers: 1, Status: 1}
	if err := plans.Create(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "serialized", Name: "Serialized", Status: 1, PlanID: &plan.ID}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	limits := NewTenantLimitService(tenants, plans, users, repository.NewCourseRepository(database))
	var created atomic.Int32
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = limits.WithEmployeeSlot(context.Background(), tenant.ID, func() error {
				created.Add(1)
				return users.Create(context.Background(), &domain.User{
					BaseModel: domain.BaseModel{TenantID: tenant.ID}, Email: "user-" + string(rune('a'+created.Load())) + "@example.com",
					Password: "hash", Name: "User", Role: "learner", Status: 1,
				})
			})
		}()
	}
	wg.Wait()
	if created.Load() != 1 {
		t.Fatalf("created callbacks = %d, want 1", created.Load())
	}
}

func TestCourseServiceEnforcesTenantCourseLimit(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	plans := repository.NewPlanRepository(database)
	plan := &domain.Plan{Name: "One course", MaxCourses: 1, Status: 1}
	if err := plans.Create(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "course-limit", Name: "Course Limit", Status: 1, PlanID: &plan.ID}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	courses := repository.NewCourseRepository(database)
	service := NewCourseService(
		courses,
		repository.NewCourseChapterRepository(database),
		repository.NewCourseLessonRepository(database),
		repository.NewCourseEnrollmentRepository(database),
	).WithTenantLimits(NewTenantLimitService(tenants, plans, users, courses))
	ctx := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	if _, err := service.Create(ctx, "First", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, "Second", "", ""); errorCode(err) != 40300 {
		t.Fatalf("second course error = %#v", err)
	}
}
