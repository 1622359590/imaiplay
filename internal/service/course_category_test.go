package service

import (
	"context"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestNormalizeCourseCategoryName(t *testing.T) {
	display, normalized, err := NormalizeCourseCategoryName("  Sales   Enablement ")
	if err != nil || display != "Sales Enablement" || normalized != "sales enablement" {
		t.Fatalf("normalized category = %q/%q err=%v", display, normalized, err)
	}
	display, normalized, err = NormalizeCourseCategoryName("  ＳＡＬＥＳ\tStraße  ")
	if err != nil || display != "SALES Straße" || normalized != "sales strasse" {
		t.Fatalf("unicode normalized category = %q/%q err=%v", display, normalized, err)
	}
	for _, value := range []string{"\t\n", strings.Repeat("界", 65)} {
		if _, _, err := NormalizeCourseCategoryName(value); errorCode(err) != 40000 {
			t.Fatalf("NormalizeCourseCategoryName(%q) error = %#v", value, err)
		}
	}
}

func TestCourseCategoryServiceEnforcesRolesAndScopes(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	service := NewCourseCategoryService(repository.NewCourseCategoryRepository(database))
	adminOne := courseContext("admin-1", "tenant-1", "tenant_admin")
	adminTwo := courseContext("admin-2", "tenant-2", "tenant_admin")
	instructor := courseContext("instructor-1", "tenant-1", "instructor")
	learner := courseContext("learner-1", "tenant-1", "learner")
	root := courseContext("root", "", "superadmin")

	created, err := service.Create(adminOne, "  Sales  Enablement ", 3, 1)
	if err != nil || created.TenantID != "tenant-1" || created.Name != "Sales Enablement" || created.NormalizedName != "sales enablement" {
		t.Fatalf("Create() = %#v, %v", created, err)
	}
	items, err := service.List(instructor)
	if err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("List(instructor) = %#v, %v", items, err)
	}
	otherItems, err := service.List(adminTwo)
	if err != nil || len(otherItems) != 0 {
		t.Fatalf("List(other tenant) = %#v, %v", otherItems, err)
	}
	if _, err := service.Create(instructor, "Forbidden", 0, 1); errorCode(err) != 40300 {
		t.Fatalf("Create(instructor) error = %#v", err)
	}
	if _, err := service.List(learner); errorCode(err) != 40300 {
		t.Fatalf("List(learner) error = %#v", err)
	}
	if _, err := service.List(root); errorCode(err) != 40300 {
		t.Fatalf("List(superadmin tenant scope) error = %#v", err)
	}
	if _, err := service.CreatePlatform(adminOne, "Forbidden", 0, 1); errorCode(err) != 40300 {
		t.Fatalf("CreatePlatform(tenant admin) error = %#v", err)
	}
	platform, err := service.CreatePlatform(root, "  Official  ", 1, 1)
	if err != nil || platform.TenantID != "" || platform.Name != "Official" {
		t.Fatalf("CreatePlatform() = %#v, %v", platform, err)
	}
	platformItems, err := service.ListPlatform(root)
	if err != nil || len(platformItems) != 1 || platformItems[0].ID != platform.ID {
		t.Fatalf("ListPlatform() = %#v, %v", platformItems, err)
	}
}

func TestCourseCategoryServiceDuplicateUpdateAndReferencedDelete(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	service := NewCourseCategoryService(repository.NewCourseCategoryRepository(database))
	admin := courseContext("admin-1", "tenant-1", "tenant_admin")
	category, err := service.Create(admin, "Straße", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(admin, " STRASSE ", 0, 1); errorCode(err) != 40900 {
		t.Fatalf("Create(duplicate normalized name) error = %#v", err)
	}
	inactive, err := service.Create(admin, "Inactive", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	var storedInactive domain.CourseCategory
	if err := database.First(&storedInactive, "id = ?", inactive.ID).Error; err != nil {
		t.Fatal(err)
	}
	if storedInactive.Status != 0 {
		t.Fatalf("stored inactive category status=%d, want 0", storedInactive.Status)
	}
	updated, err := service.Update(admin, category.ID, "  Sales ", 7, 0)
	if err != nil || updated.Name != "Sales" || updated.NormalizedName != "sales" || updated.SortOrder != 7 || updated.Status != 0 {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if _, err := service.Update(admin, category.ID, "Sales", 0, 2); errorCode(err) != 40000 {
		t.Fatalf("Update(invalid status) error = %#v", err)
	}
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Title:     "Referenced", CreatedBy: "admin-1", CategoryID: &category.ID,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(admin, category.ID); errorCode(err) != 40900 {
		t.Fatalf("Delete(referenced) error = %#v", err)
	}
	foreign := courseContext("admin-2", "tenant-2", "tenant_admin")
	if _, err := service.Update(foreign, category.ID, "Foreign", 0, 1); errorCode(err) != 40400 {
		t.Fatalf("Update(cross-tenant) error = %#v", err)
	}
	if err := service.Delete(foreign, category.ID); errorCode(err) != 40400 {
		t.Fatalf("Delete(cross-tenant) error = %#v", err)
	}
	if _, err := service.Create(context.Background(), "Anonymous", 0, 1); errorCode(err) != 40300 {
		t.Fatalf("Create(anonymous) error = %#v", err)
	}
}
