package service

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestPlanServiceAdministrationAndAssignment(t *testing.T) {
	database, tenants, _ := serviceRepositories(t)
	plans := repository.NewPlanRepository(database)
	resources := repository.NewResourceRepository(database)
	service := NewPlanService(plans, tenants, resources)
	root := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	admin := usercontext.WithUser(context.Background(), "admin", "tenant-1", "admin@example.com", "tenant_admin")

	if _, _, err := service.List(admin, 0, 10); errorCode(err) != 40300 {
		t.Fatalf("List(non-admin) error = %#v", err)
	}
	if _, err := service.Create(root, &domain.Plan{}); errorCode(err) != 40000 {
		t.Fatalf("Create(empty) error = %#v", err)
	}
	plan, err := service.Create(root, &domain.Plan{Name: "Team", MaxUsers: 10, MaxCourses: 3})
	if err != nil || plan.Features != "{}" {
		t.Fatalf("Create() = %#v, %v", plan, err)
	}
	if _, _, err := service.List(root, 0, 20); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	plan.Name, plan.MaxUsers = "Team Plus", 20
	if _, err := service.Update(root, plan); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	tenant := &domain.Tenant{ID: "tenant-plan-service", Code: "plan-service", Name: "Tenant", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if _, err := service.Assign(root, tenant.ID, plan.ID); err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	current, err := service.Current(usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin"))
	if err != nil || current["quota_bytes"] != plan.StorageQuotaBytes {
		t.Fatalf("Current() = %#v, %v", current, err)
	}
	if err := service.Delete(root, plan.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if errorCode(service.Delete(root, plan.ID)) != 40400 {
		t.Fatal("Delete(missing) did not return not found")
	}
}

func TestCourseOfficialServiceFlows(t *testing.T) {
	fixture := newCourseFixture(t)
	root := courseContext("root", "", "superadmin")
	admin := courseContext("admin", "tenant-1", "tenant_admin")
	if _, err := fixture.courses.CreateOfficial(admin, "Official", "", ""); errorCode(err) != 40300 {
		t.Fatalf("CreateOfficial(non-admin) error = %#v", err)
	}
	official, err := fixture.courses.CreateOfficial(root, "Official", "Description", "")
	if err != nil || !official.IsOfficial {
		t.Fatalf("CreateOfficial() = %#v, %v", official, err)
	}
	if items, total, err := fixture.courses.OfficialList(root, 0, 10); err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("OfficialList(root) = %#v, %d, %v", items, total, err)
	}
	if err := fixture.courses.EnableOfficial(admin, official.ID, true); err != nil {
		t.Fatalf("EnableOfficial() error = %v", err)
	}
	if err := fixture.courses.EnableOfficial(root, official.ID, true); errorCode(err) != 40300 {
		t.Fatalf("EnableOfficial(root) error = %#v", err)
	}
}

func TestTenantThemeServiceUpdateValidation(t *testing.T) {
	_, tenants, _ := serviceRepositories(t)
	tenant := &domain.Tenant{ID: "theme-service", Code: "theme-service", Name: "Theme", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewTenantThemeService(tenants)
	admin := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	if _, err := service.Update(admin, "bad", "", ""); errorCode(err) != 40000 {
		t.Fatalf("Update(invalid color) error = %#v", err)
	}
	updated, err := service.Update(admin, "#abcdef", " /logo.png ", " Welcome ")
	if err != nil || updated.PrimaryColor != "#abcdef" || updated.LogoURL != "/logo.png" || updated.WelcomeText != "Welcome" {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
}

func TestAuthBootstrapAndIssueTokens(t *testing.T) {
	_, tenants, users := serviceRepositories(t)
	auth := NewAuthService(users, tenants, "coverage-secret")
	root := context.Background()
	user, pair, err := auth.BootstrapSuperadmin(root, " ROOT@EXAMPLE.COM ", " Root ", "password123")
	if err != nil || user.Role != "superadmin" || pair.AccessToken == "" {
		t.Fatalf("BootstrapSuperadmin() = %#v, %#v, %v", user, pair, err)
	}
	if _, _, err := auth.BootstrapSuperadmin(root, "second@example.com", "Second", "password123"); errorCode(err) != 40900 {
		t.Fatalf("duplicate bootstrap error = %#v", err)
	}
	if _, _, err := auth.BootstrapSuperadmin(root, "bad@example.com", "", "short"); errorCode(err) != 40900 {
		t.Fatalf("second bootstrap guard error = %#v", err)
	}
	plain, err := auth.IssueTokens(root, user)
	if err != nil || plain.AccessToken == "" {
		t.Fatalf("IssueTokens() = %#v, %v", plain, err)
	}
}

func TestAuthBootstrapValidationAndPlanGuards(t *testing.T) {
	_, tenants, users := serviceRepositories(t)
	auth := NewAuthService(users, tenants, "coverage-secret")
	if _, _, err := auth.BootstrapSuperadmin(context.Background(), "bad@example.com", "", "short"); errorCode(err) != 40000 {
		t.Fatalf("invalid bootstrap error = %#v", err)
	}
	database, tenantRepo, _ := serviceRepositories(t)
	planRepo := repository.NewPlanRepository(database)
	planService := NewPlanService(planRepo, tenantRepo, repository.NewResourceRepository(database))
	root := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	if _, err := planService.Create(root, &domain.Plan{Name: "Bad", MaxUsers: -1}); errorCode(err) != 40000 {
		t.Fatalf("negative plan error = %#v", err)
	}
	if _, err := planService.Update(root, &domain.Plan{ID: "missing"}); errorCode(err) != 40000 {
		t.Fatalf("invalid update error = %#v", err)
	}
	if _, err := planService.Assign(root, "missing-tenant", "missing-plan"); errorCode(err) != 40400 {
		t.Fatalf("missing plan assignment error = %#v", err)
	}
}

func TestUserServicePhoneValidationAndLifecycle(t *testing.T) {
	_, tenants, users := serviceRepositories(t)
	tenant := &domain.Tenant{ID: "user-service", Code: "user-service", Name: "Users", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	service := NewUserService(users)
	admin := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	if _, err := service.CreateWithPhone(admin, "learner@example.com", "bad", "password123", "Learner", "learner"); errorCode(err) != 40000 {
		t.Fatalf("invalid phone error = %#v", err)
	}
	user, err := service.CreateWithPhone(admin, "learner@example.com", "13800138000", "password123", "Learner", "learner")
	if err != nil || user.Phone == nil {
		t.Fatalf("CreateWithPhone() = %#v, %v", user, err)
	}
	if _, err := service.Create(admin, "learner@example.com", "password123", "Duplicate", "learner"); errorCode(err) != 40900 {
		t.Fatalf("duplicate user error = %#v", err)
	}
}

func TestResourceServiceListAndOpen(t *testing.T) {
	service := newResourceService(t, t.TempDir())
	ctx := courseContext("admin", "tenant-resource", "tenant_admin")
	resource, err := service.Upload(ctx, "guide.pdf", bytes.NewReader([]byte("%PDF-1.7\n")), 9)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	items, total, err := service.List(ctx, 0, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("List() = %#v, %d, %v", items, total, err)
	}
	body, contentType, name, err := service.Open(ctx, resource.ID)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer body.Close()
	data, _ := io.ReadAll(body)
	if string(data) != "%PDF-1.7\n" || contentType != "application/pdf" || name != "guide.pdf" {
		t.Fatalf("Open() = %q, %q, %q", data, contentType, name)
	}
	if _, _, _, err := service.Open(courseContext("other", "other-tenant", "tenant_admin"), resource.ID); errorCode(err) != 40400 {
		t.Fatalf("cross-tenant Open() error = %#v", err)
	}
}

func TestCourseServiceDeleteAndOfficialGuards(t *testing.T) {
	fixture := newCourseFixture(t)
	admin := courseContext("admin", "tenant-1", "tenant_admin")
	course, err := fixture.courses.Create(admin, "Delete me", "", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := fixture.courses.Delete(admin, course.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if errorCode(fixture.courses.Delete(admin, course.ID)) != 40400 {
		t.Fatal("Delete(missing) did not return not found")
	}
}

func TestTenantLifecycleAndAccessibility(t *testing.T) {
	_, tenants, _ := serviceRepositories(t)
	tenant := &domain.Tenant{ID: "lifecycle-service", Code: "lifecycle-service", Name: "Lifecycle", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	root := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	if _, err := NewTenantService(tenants).UpdateLifecycle(root, tenant.ID, "invalid", nil); errorCode(err) != 40000 {
		t.Fatalf("invalid lifecycle error = %#v", err)
	}
	expires := time.Now().UTC().Add(time.Hour)
	updated, err := NewTenantService(tenants).UpdateLifecycle(root, tenant.ID, "trial", &expires)
	if err != nil || updated.LifecycleStatus != "trial" || updated.TrialEndsAt == nil {
		t.Fatalf("UpdateLifecycle() = %#v, %v", updated, err)
	}
	if ok, _ := TenantAccessible(&domain.Tenant{Status: 0}, time.Now()); ok {
		t.Fatal("disabled tenant was accessible")
	}
	if ok, _ := TenantAccessible(&domain.Tenant{LifecycleStatus: "suspended"}, time.Now()); ok {
		t.Fatal("suspended tenant was accessible")
	}
	expired := time.Now().Add(-time.Minute)
	if ok, _ := TenantAccessible(&domain.Tenant{LifecycleStatus: "trial", TrialEndsAt: &expired}, time.Now()); ok {
		t.Fatal("expired trial was accessible")
	}
	if ok, reason := TenantAccessible(&domain.Tenant{Status: 1}, time.Now()); !ok || reason != "" {
		t.Fatalf("active tenant accessibility = %v, %q", ok, reason)
	}
}

func TestServiceSmallBranches(t *testing.T) {
	if got := resourceContentType("unknown"); got != "application/octet-stream" {
		t.Fatalf("resourceContentType() = %q", got)
	}
	if defaultTheme().PrimaryColor != DefaultPrimaryColor {
		t.Fatal("defaultTheme() did not set primary color")
	}
	if errorCode(mapCreateError(nil, "conflict", "internal")) != 50000 {
		t.Fatal("mapCreateError(nil) did not return internal error")
	}
	database, tenants, _ := serviceRepositories(t)
	planService := NewPlanService(repository.NewPlanRepository(database), tenants, repository.NewResourceRepository(database))
	admin := usercontext.WithUser(context.Background(), "admin", "tenant", "admin@example.com", "tenant_admin")
	if _, err := planService.Current(admin); errorCode(err) != 40400 {
		t.Fatalf("Current(missing tenant) error = %#v", err)
	}
	if errorCode(planService.Delete(admin, "missing")) != 40300 {
		t.Fatalf("Delete(non-admin) error = %#v", planService.Delete(admin, "missing"))
	}
	fixture := newCourseFixture(t)
	if _, _, err := fixture.courses.OfficialList(admin, 0, 10); err != nil {
		t.Fatalf("OfficialList(tenant admin) error = %v", err)
	}
}

func TestEnrollmentServiceListAndRemove(t *testing.T) {
	database, tenants, users := serviceRepositories(t)
	tenant := &domain.Tenant{ID: "enrollment-service", Code: "enrollment-service", Name: "Enrollment", Status: 1}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	userRepo := users
	learner := &domain.User{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Email: "learner@enroll.test", Name: "Learner", Role: "learner", Status: 1}
	if err := userRepo.Create(context.Background(), learner); err != nil {
		t.Fatalf("create learner: %v", err)
	}
	courses := repository.NewCourseRepository(database)
	course := &domain.Course{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Title: "Course", CreatedBy: "admin", Status: 1}
	if err := courses.Create(context.Background(), course); err != nil {
		t.Fatalf("create course: %v", err)
	}
	enrollments := repository.NewCourseEnrollmentRepository(database)
	enrollmentService := NewEnrollmentService(enrollments, courses, userRepo)
	admin := usercontext.WithUser(context.Background(), "admin", tenant.ID, "admin@enroll.test", "tenant_admin")
	enrollment, err := enrollmentService.Enroll(admin, course.ID, learner.ID)
	if err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}
	items, err := enrollmentService.ListByCourse(admin, course.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListByCourse() = %#v, %v", items, err)
	}
	if err := enrollmentService.Remove(admin, enrollment.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if errorCode(enrollmentService.Remove(admin, enrollment.ID)) != 40400 {
		t.Fatal("Remove(missing) did not return not found")
	}
}

func TestTenantRegistrationSuperadminFlow(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	service := NewTenantRegistrationService(database, "registration-secret")
	root := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	result, err := service.CreateForSuperadmin(root, "Managed Tenant", "managed@example.com", "", "Managed Admin", "password123", "")
	if err != nil || result.Tenant == nil || result.Token != "" {
		t.Fatalf("CreateForSuperadmin() = %#v, %v", result, err)
	}
	if _, err := service.CreateForSuperadmin(root, "Bad Plan", "badplan@example.com", "", "Admin", "password123", "missing-plan"); errorCode(err) != 40000 {
		t.Fatalf("invalid plan error = %#v", err)
	}
	admin := usercontext.WithUser(context.Background(), result.User.ID, result.Tenant.ID, result.User.Email, result.User.Role)
	if err := service.ClearDemoData(admin); err != nil {
		t.Fatalf("ClearDemoData() error = %v", err)
	}
}

func TestServicePermissionGuards(t *testing.T) {
	ctx := context.Background()
	database, tenants, users := serviceRepositories(t)
	plan := NewPlanService(repository.NewPlanRepository(database), tenants, repository.NewResourceRepository(database))
	_, _, _ = plan.List(ctx, 0, 1)
	_, _ = plan.Create(ctx, &domain.Plan{Name: "x"})
	_, _ = plan.Update(ctx, &domain.Plan{Name: "x"})
	_ = plan.Delete(ctx, "id")
	_, _ = plan.Assign(ctx, "tenant", "plan")
	_, _ = plan.Current(ctx)
	_ = plan.CheckStorage(ctx, "tenant", 1)

	user := NewUserService(users)
	_, _ = user.Create(ctx, "a@example.com", "password123", "A", "learner")
	_, _, _ = user.List(ctx, 0, 1)
	_, _ = user.Get(ctx, "id")
	_, _ = user.Update(ctx, "id", "A", 1)
	_ = user.Delete(ctx, "id")

	courseFixture := newCourseFixture(t)
	_, _ = courseFixture.courses.CreateOfficial(ctx, "x", "", "")
	_, _, _ = courseFixture.courses.OfficialList(ctx, 0, 1)
	_ = courseFixture.courses.EnableOfficial(ctx, "id", true)
	_, _, _ = courseFixture.courses.ListPublished(ctx, 0, 1)
	_, _ = courseFixture.courses.GetPublishedDetail(ctx, "id")

	resource := newResourceService(t, t.TempDir())
	_, _, _ = resource.List(ctx, 0, 1)
	_, _, _, _ = resource.File(ctx, "id", t.TempDir())
	_, _, _, _ = resource.Open(ctx, "id")

	theme := NewTenantThemeService(tenants)
	_, _ = theme.Update(ctx, "#ffffff", "", "")

	auth := NewAuthService(users, tenants, "secret")
	_, _ = auth.IssueTokens(ctx, &domain.User{BaseModel: domain.BaseModel{ID: "id"}, Role: "learner"})
	_, _ = auth.LoginWithCode(ctx, "13800138000", "000000")
	_, _ = auth.Login(ctx, "", "")
	_ = auth.SendLoginCode(ctx, "bad")
	_ = auth.ForgotPassword(ctx, "bad")
	_ = auth.ResetPassword(ctx, "bad", "000000", "short")
	_, _ = auth.Refresh(ctx, "")
	_ = auth.Logout(ctx, "")
	_ = auth.Logout(usercontext.WithUser(ctx, "id", "tenant", "id@example.com", "learner"), "")
	_, _ = courseFixture.courses.Update(ctx, "id", "", "", "", 9)
	_ = resourceContentType("image")
	_ = resourceContentType("video")
	_ = validPhone("13800138000")
	_ = nullablePhone("13800138000")
	_ = tenantCodeSlug("Coverage Tenant")
	_, _ = pagination(-1, 0, 1)
}
