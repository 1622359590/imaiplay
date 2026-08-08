package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestCourseEnrollmentRepositoryCRUDAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseEnrollmentRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	seedEnrollmentAuthorizationFixtures(t, database)
	first := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "tenant-course-0", UserID: "learner-1", Status: 1,
	}
	second := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "tenant-course-0", UserID: "learner-2", Status: 1,
	}
	foreign := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"},
		CourseID:  "foreign-course", UserID: "foreign-learner", Status: 1,
	}
	for index, enrollment := range []*domain.CourseEnrollment{first, second, foreign} {
		ctx := tenantOne
		if index == 2 {
			ctx = tenantTwo
		}
		if err := repo.Create(ctx, enrollment); err != nil {
			t.Fatalf("Create(%s) error = %v", enrollment.UserID, err)
		}
	}
	if _, err := repo.FindByID(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant FindByID() error = %v", err)
	}
	items, err := repo.FindByCourse(tenantOne, "tenant-course-0")
	if err != nil || len(items) != 2 {
		t.Fatalf("FindByCourse() = %#v, %v", items, err)
	}
	items, err = repo.FindByUser(tenantOne, "learner-1")
	if err != nil || len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("FindByUser() = %#v, %v", items, err)
	}
	found, err := repo.FindByCourseAndUser(tenantOne, "tenant-course-0", "learner-1")
	if err != nil || found.ID != first.ID {
		t.Fatalf("FindByCourseAndUser() = %#v, %v", found, err)
	}
	duplicate := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "tenant-course-0", UserID: "learner-1", Status: 1,
	}
	if err := repo.Create(tenantOne, duplicate); err == nil {
		t.Fatal("Create(duplicate) error = nil")
	}
	if err := repo.Delete(tenantTwo, first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant Delete() error = %v", err)
	}
	if err := repo.Delete(tenantOne, first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestCourseEnrollmentRepositoryUpdateAssignmentIsTenantScoped(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := NewCourseEnrollmentRepository(database)
	base := context.Background()
	tenantOne := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	tenantTwo := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	seedEnrollmentAuthorizationFixtures(t, database)
	enrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: "tenant-course-0",
		UserID: "learner-1", Status: 1, AssignmentType: domain.AssignmentRequired,
	}
	if err := repo.Create(tenantOne, enrollment); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.UpdateAssignment(tenantTwo, enrollment.ID, domain.AssignmentOptional); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-tenant UpdateAssignment() error = %v", err)
	}
	if err := repo.UpdateAssignment(tenantOne, enrollment.ID, domain.AssignmentOptional); err != nil {
		t.Fatalf("UpdateAssignment() error = %v", err)
	}
	updated, err := repo.FindByID(tenantOne, enrollment.ID)
	if err != nil || updated.AssignmentType != domain.AssignmentOptional {
		t.Fatalf("FindByID(updated) = %#v, %v", updated, err)
	}
}

func TestCourseEnrollmentRepositoryCreateEnforcesActorScopeAndAssignmentType(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	seedEnrollmentAuthorizationFixtures(t, database)
	repo := NewCourseEnrollmentRepository(database)
	base := context.Background()
	admin := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	learner := usercontext.WithUser(base, "learner-1", "tenant-1", "", "learner")
	instructor := usercontext.WithUser(base, "instructor-1", "tenant-1", "", "instructor")

	adminEnrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{ID: "admin-enrollment", TenantID: "tenant-1"},
		CourseID:  "tenant-course-0", UserID: "learner-1", Status: 1,
		AssignmentType: domain.AssignmentOptional,
	}
	if err := repo.Create(admin, adminEnrollment); err != nil {
		t.Fatalf("tenant admin Create() error = %v", err)
	}
	selfEnrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{ID: "self-enrollment", TenantID: "tenant-1"},
		CourseID:  "tenant-course-1", UserID: "learner-1", Status: 1,
	}
	if err := repo.Create(learner, selfEnrollment); err != nil {
		t.Fatalf("learner self Create() error = %v", err)
	}
	if selfEnrollment.AssignmentType != domain.AssignmentRequired {
		t.Fatalf("learner assignment type = %q, want required", selfEnrollment.AssignmentType)
	}
	optionalSelfEnrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{ID: "optional-self-enrollment", TenantID: "tenant-1"},
		CourseID:  "tenant-course-2", UserID: "learner-1", Status: 1,
		AssignmentType: domain.AssignmentOptional,
	}
	if err := database.Model(&domain.Course{}).Where("id = ?", "tenant-course-2").
		Update("course_type", domain.CourseTypeOptional).Error; err != nil {
		t.Fatalf("set optional course type: %v", err)
	}
	if err := repo.Create(learner, optionalSelfEnrollment); err != nil {
		t.Fatalf("learner optional self Create() error = %v", err)
	}
	if optionalSelfEnrollment.AssignmentType != domain.CourseTypeOptional {
		t.Fatalf("learner explicit optional assignment type = %q, want optional", optionalSelfEnrollment.AssignmentType)
	}
	assertEnrollmentAssignment(t, database, optionalSelfEnrollment.ID, domain.CourseTypeOptional)
	officialEnrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{ID: "official-self-enrollment", TenantID: "tenant-1"},
		CourseID:  "official-enabled", UserID: "learner-1", Status: 1,
		AssignmentType: domain.AssignmentOptional,
	}
	if err := database.Model(&domain.Course{}).Where("id = ?", "official-enabled").
		Update("course_type", domain.CourseTypeOptional).Error; err != nil {
		t.Fatalf("set official optional course type: %v", err)
	}
	if err := repo.Create(learner, officialEnrollment); err != nil {
		t.Fatalf("learner official Create() error = %v", err)
	}
	if officialEnrollment.AssignmentType != domain.AssignmentOptional {
		t.Fatalf("learner official assignment type = %q, want optional", officialEnrollment.AssignmentType)
	}
	assertEnrollmentAssignment(t, database, officialEnrollment.ID, domain.AssignmentOptional)

	unauthorized := []struct {
		name       string
		ctx        context.Context
		enrollment *domain.CourseEnrollment
	}{
		{"instructor", instructor, enrollmentFixture("create-instructor", "tenant-1", "tenant-course-2", "learner-1", 1, domain.AssignmentRequired)},
		{"unauthenticated", base, enrollmentFixture("create-unauthenticated", "tenant-1", "tenant-course-3", "learner-1", 1, domain.AssignmentRequired)},
		{"forged tenant", admin, enrollmentFixture("create-forged-tenant", "tenant-2", "tenant-course-4", "learner-1", 1, domain.AssignmentRequired)},
		{"learner creates for another user", learner, enrollmentFixture("create-other-user", "tenant-1", "tenant-course-5", "learner-2", 1, domain.AssignmentRequired)},
		{"admin targets instructor", admin, enrollmentFixture("create-instructor-target", "tenant-1", "tenant-course-6", "instructor-1", 1, domain.AssignmentRequired)},
		{"admin targets foreign learner", admin, enrollmentFixture("create-foreign-user", "tenant-1", "tenant-course-7", "foreign-learner", 1, domain.AssignmentRequired)},
		{"learner inactive enrollment", learner, enrollmentFixture("create-inactive", "tenant-1", "tenant-course-8", "learner-1", 0, domain.AssignmentRequired)},
		{"learner foreign course", learner, enrollmentFixture("create-foreign-course", "tenant-1", "foreign-course", "learner-1", 1, domain.AssignmentRequired)},
		{"learner draft course", learner, enrollmentFixture("create-draft-course", "tenant-1", "draft-course", "learner-1", 1, domain.AssignmentRequired)},
		{"learner disabled account", usercontext.WithUser(base, "disabled-learner", "tenant-1", "", "learner"), enrollmentFixture("create-disabled-user", "tenant-1", "tenant-course-9", "disabled-learner", 1, domain.AssignmentRequired)},
		{"learner unactivated official", learner, enrollmentFixture("create-disabled-official", "tenant-1", "official-disabled", "learner-1", 1, domain.AssignmentRequired)},
	}
	for _, test := range unauthorized {
		t.Run(test.name, func(t *testing.T) {
			if err := repo.Create(test.ctx, test.enrollment); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Create() error = %v, want record not found", err)
			}
			assertEnrollmentMissing(t, database, test.enrollment.ID)
		})
	}
	for _, test := range []struct {
		name       string
		ctx        context.Context
		enrollment *domain.CourseEnrollment
	}{
		{"admin invalid assignment", admin, enrollmentFixture("create-invalid-admin", "tenant-1", "tenant-course-10", "learner-1", 1, "recommended")},
		{"learner invalid assignment", learner, enrollmentFixture("create-invalid-learner", "tenant-1", "tenant-course-11", "learner-1", 1, "recommended")},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := repo.Create(test.ctx, test.enrollment); err == nil {
				t.Fatal("Create() error = nil")
			}
			assertEnrollmentMissing(t, database, test.enrollment.ID)
		})
	}
}

func TestCourseEnrollmentRepositoryUpdateAndDeleteRequireTenantAdmin(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	seedEnrollmentAuthorizationFixtures(t, database)
	for _, enrollment := range []*domain.CourseEnrollment{
		enrollmentFixture("update-admin", "tenant-1", "tenant-course-0", "learner-1", 1, domain.AssignmentRequired),
		enrollmentFixture("update-instructor", "tenant-1", "tenant-course-1", "learner-1", 1, domain.AssignmentRequired),
		enrollmentFixture("update-unauthenticated", "tenant-1", "tenant-course-2", "learner-1", 1, domain.AssignmentRequired),
		enrollmentFixture("delete-admin", "tenant-1", "tenant-course-3", "learner-1", 1, domain.AssignmentRequired),
		enrollmentFixture("delete-instructor", "tenant-1", "tenant-course-4", "learner-1", 1, domain.AssignmentRequired),
		enrollmentFixture("delete-unauthenticated", "tenant-1", "tenant-course-5", "learner-1", 1, domain.AssignmentRequired),
	} {
		if err := database.Create(enrollment).Error; err != nil {
			t.Fatalf("create enrollment %s: %v", enrollment.ID, err)
		}
	}
	repo := NewCourseEnrollmentRepository(database)
	base := context.Background()
	admin := usercontext.WithUser(base, "admin-1", "tenant-1", "", "tenant_admin")
	foreignAdmin := usercontext.WithUser(base, "admin-2", "tenant-2", "", "tenant_admin")
	instructor := usercontext.WithUser(base, "instructor-1", "tenant-1", "", "instructor")

	if err := repo.UpdateAssignment(admin, "update-admin", domain.AssignmentOptional); err != nil {
		t.Errorf("tenant admin UpdateAssignment() error = %v", err)
	}
	for _, test := range []struct {
		name, id string
		ctx      context.Context
	}{
		{"instructor", "update-instructor", instructor},
		{"unauthenticated", "update-unauthenticated", base},
		{"cross tenant", "update-instructor", foreignAdmin},
	} {
		t.Run("update "+test.name, func(t *testing.T) {
			if err := repo.UpdateAssignment(test.ctx, test.id, domain.AssignmentOptional); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("UpdateAssignment() error = %v, want record not found", err)
			}
			assertEnrollmentAssignment(t, database, test.id, domain.AssignmentRequired)
		})
	}
	assertEnrollmentAssignment(t, database, "update-admin", domain.AssignmentOptional)

	if err := repo.Delete(admin, "delete-admin"); err != nil {
		t.Errorf("tenant admin Delete() error = %v", err)
	}
	assertEnrollmentMissing(t, database, "delete-admin")
	for _, test := range []struct {
		name, id string
		ctx      context.Context
	}{
		{"instructor", "delete-instructor", instructor},
		{"unauthenticated", "delete-unauthenticated", base},
		{"cross tenant", "delete-instructor", foreignAdmin},
	} {
		t.Run("delete "+test.name, func(t *testing.T) {
			if err := repo.Delete(test.ctx, test.id); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("Delete() error = %v, want record not found", err)
			}
			assertEnrollmentAssignment(t, database, test.id, domain.AssignmentRequired)
		})
	}
	if err := repo.UpdateAssignment(admin, "update-admin", "recommended"); err == nil {
		t.Error("invalid UpdateAssignment() error = nil")
	}
	if err := repo.UpdateAssignment(admin, "update-admin", ""); err == nil {
		t.Error("empty UpdateAssignment() error = nil")
	}
	assertEnrollmentAssignment(t, database, "update-admin", domain.AssignmentOptional)
}

func seedEnrollmentAuthorizationFixtures(t *testing.T, database *gorm.DB) {
	t.Helper()
	for _, tenant := range []*domain.Tenant{
		{ID: "tenant-1", Code: "tenant-1", Name: "Tenant 1", Status: 1},
		{ID: "tenant-2", Code: "tenant-2", Name: "Tenant 2", Status: 1},
	} {
		if err := database.Create(tenant).Error; err != nil {
			t.Fatalf("create tenant %s: %v", tenant.ID, err)
		}
	}
	for _, user := range []*domain.User{
		{BaseModel: domain.BaseModel{ID: "learner-1", TenantID: "tenant-1"}, Email: "learner-1@example.com", Password: "hash", Name: "Learner 1", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "learner-2", TenantID: "tenant-1"}, Email: "learner-2@example.com", Password: "hash", Name: "Learner 2", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{ID: "disabled-learner", TenantID: "tenant-1"}, Email: "disabled@example.com", Password: "hash", Name: "Disabled", Role: "learner", Status: 0},
		{BaseModel: domain.BaseModel{ID: "instructor-1", TenantID: "tenant-1"}, Email: "instructor@example.com", Password: "hash", Name: "Instructor", Role: "instructor", Status: 1},
		{BaseModel: domain.BaseModel{ID: "foreign-learner", TenantID: "tenant-2"}, Email: "foreign@example.com", Password: "hash", Name: "Foreign", Role: "learner", Status: 1},
	} {
		requestedStatus := user.Status
		if err := database.Create(user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.ID, err)
		}
		if requestedStatus == 0 {
			if err := database.Model(&domain.User{}).Where("id = ?", user.ID).Update("status", 0).Error; err != nil {
				t.Fatalf("disable user %s: %v", user.ID, err)
			}
		}
	}
	for index := 0; index < 12; index++ {
		course := &domain.Course{BaseModel: domain.BaseModel{ID: fmt.Sprintf("tenant-course-%d", index), TenantID: "tenant-1"}, Title: fmt.Sprintf("Course %d", index), Status: 1, CreatedBy: "admin-1"}
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", course.ID, err)
		}
	}
	for _, course := range []*domain.Course{
		{BaseModel: domain.BaseModel{ID: "draft-course", TenantID: "tenant-1"}, Title: "Draft", CreatedBy: "admin-1"},
		{BaseModel: domain.BaseModel{ID: "foreign-course", TenantID: "tenant-2"}, Title: "Foreign", Status: 1, CreatedBy: "admin-2"},
		{BaseModel: domain.BaseModel{ID: "official-enabled"}, Title: "Official enabled", Status: 1, CreatedBy: "root", IsOfficial: true},
		{BaseModel: domain.BaseModel{ID: "official-disabled"}, Title: "Official disabled", Status: 1, CreatedBy: "root", IsOfficial: true},
	} {
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", course.ID, err)
		}
	}
	if err := database.Create(&domain.TenantOfficialCourse{TenantID: "tenant-1", CourseID: "official-enabled", Enabled: true}).Error; err != nil {
		t.Fatalf("activate official course: %v", err)
	}
}

func enrollmentFixture(id, tenantID, courseID, userID string, status int, assignmentType string) *domain.CourseEnrollment {
	return &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{ID: id, TenantID: tenantID},
		CourseID:  courseID, UserID: userID, Status: status,
		AssignmentType: assignmentType,
	}
}

func assertEnrollmentMissing(t *testing.T, database *gorm.DB, id string) {
	t.Helper()
	var count int64
	if err := database.Model(&domain.CourseEnrollment{}).Where("id = ?", id).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("enrollment %s count = %d, error = %v", id, count, err)
	}
}

func assertEnrollmentAssignment(t *testing.T, database *gorm.DB, id, assignmentType string) {
	t.Helper()
	var enrollment domain.CourseEnrollment
	if err := database.First(&enrollment, "id = ?", id).Error; err != nil || enrollment.AssignmentType != assignmentType {
		t.Fatalf("enrollment %s = %#v, error = %v", id, enrollment, err)
	}
}
