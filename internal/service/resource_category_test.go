package service

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestResourceCategoryServiceCRUDAndParentValidation(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	service := NewResourceCategoryService(
		repository.NewResourceCategoryRepository(database),
	)
	admin := courseContext("admin-1", "tenant-1", "tenant_admin")
	parent, err := service.Create(admin, "Videos", nil)
	if err != nil {
		t.Fatalf("Create(parent) error = %v", err)
	}
	child, err := service.Create(admin, "Onboarding", &parent.ID)
	if err != nil || child.ParentID == nil || *child.ParentID != parent.ID {
		t.Fatalf("Create(child) = %#v, %v", child, err)
	}
	items, err := service.List(admin)
	if err != nil || len(items) != 2 {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	if _, err := service.Update(
		admin, parent.ID, "Videos", &parent.ID,
	); errorCode(err) != 40000 {
		t.Fatalf("Update(self parent) error = %#v", err)
	}
	if _, err := service.Update(
		admin, parent.ID, "Videos", &child.ID,
	); errorCode(err) != 40000 {
		t.Fatalf("Update(cyclic parent) error = %#v", err)
	}
	if err := service.Delete(admin, parent.ID); errorCode(err) != 40900 {
		t.Fatalf("Delete(parent with child) error = %#v", err)
	}
	updated, err := service.Update(admin, child.ID, "Orientation", nil)
	if err != nil || updated.Name != "Orientation" || updated.ParentID != nil {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if err := service.Delete(admin, child.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestResourceCategoryServiceRejectsInvalidRoleAndForeignParent(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	repo := repository.NewResourceCategoryRepository(database)
	service := NewResourceCategoryService(repo)
	learner := courseContext("learner-1", "tenant-1", "learner")
	if _, err := service.Create(learner, "Videos", nil); errorCode(err) != 40300 {
		t.Fatalf("Create(learner) error = %#v", err)
	}
	instructor := courseContext("instructor-1", "tenant-1", "instructor")
	if _, err := service.Create(instructor, "Videos", nil); errorCode(err) != 40300 {
		t.Fatalf("Create(instructor) error = %#v", err)
	}
	if _, err := service.Update(instructor, "category", "Changed", nil); errorCode(err) != 40300 {
		t.Fatalf("Update(instructor) error = %#v", err)
	}
	if err := service.Delete(instructor, "category"); errorCode(err) != 40300 {
		t.Fatalf("Delete(instructor) error = %#v", err)
	}
	foreign := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"}, Name: "Foreign",
	}
	if err := database.Create(foreign).Error; err != nil {
		t.Fatalf("create foreign category: %v", err)
	}
	admin := courseContext("admin-1", "tenant-1", "tenant_admin")
	if _, err := service.Create(
		admin, "Child", &foreign.ID,
	); errorCode(err) != 40400 {
		t.Fatalf("Create(foreign parent) error = %#v", err)
	}
	if _, err := service.Create(admin, "   ", nil); errorCode(err) != 40000 {
		t.Fatalf("Create(empty name) error = %#v", err)
	}
}
