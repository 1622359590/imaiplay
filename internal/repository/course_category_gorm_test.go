package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestCourseCategoryRepositoryCRUDAndScopeIsolation(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewCourseCategoryRepository(database)
	ctx := context.Background()
	tenantCategory := &domain.CourseCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "Sales", NormalizedName: "sales", SortOrder: 2, Status: 1,
	}
	otherTenantCategory := &domain.CourseCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-2"},
		Name:      "Sales", NormalizedName: "sales", Status: 1,
	}
	platformCategory := &domain.CourseCategory{
		Name: "Official", NormalizedName: "official", Status: 1,
	}
	for _, category := range []*domain.CourseCategory{
		tenantCategory, otherTenantCategory, platformCategory,
	} {
		if err := repo.Create(ctx, category); err != nil {
			t.Fatalf("Create(%q) error = %v", category.TenantID, err)
		}
	}

	items, err := repo.FindByTenant(ctx, "tenant-1")
	if err != nil || len(items) != 1 || items[0].ID != tenantCategory.ID {
		t.Fatalf("FindByTenant(tenant-1) = %#v, %v", items, err)
	}
	platform, err := repo.FindByTenant(ctx, "")
	if err != nil || len(platform) != 1 || platform[0].ID != platformCategory.ID {
		t.Fatalf("FindByTenant(platform) = %#v, %v", platform, err)
	}
	if _, err := repo.FindByID(ctx, "tenant-2", tenantCategory.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-scope FindByID() error = %v", err)
	}

	tenantCategory.Name = "Revenue"
	tenantCategory.NormalizedName = "revenue"
	if err := repo.Update(ctx, "tenant-2", tenantCategory); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-scope Update() error = %v", err)
	}
	if err := repo.Update(ctx, "tenant-1", tenantCategory); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err := repo.FindByID(ctx, "tenant-1", tenantCategory.ID)
	if err != nil || found.Name != "Revenue" || found.NormalizedName != "revenue" {
		t.Fatalf("FindByID(updated) = %#v, %v", found, err)
	}
	if err := repo.Delete(ctx, "tenant-2", tenantCategory.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-scope Delete() error = %v", err)
	}
	if err := repo.Delete(ctx, "tenant-1", tenantCategory.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestCourseCategoryRepositoryRejectsReferencedDelete(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewCourseCategoryRepository(database)
	category := &domain.CourseCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "Sales", NormalizedName: "sales", Status: 1,
	}
	if err := repo.Create(context.Background(), category); err != nil {
		t.Fatal(err)
	}
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Title:     "Referenced", CreatedBy: "admin-1", CategoryID: &category.ID,
	}
	if err := database.Create(course).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(context.Background(), "tenant-1", category.ID); !errors.Is(err, ErrCourseCategoryInUse) {
		t.Fatalf("Delete(referenced) error = %v", err)
	}
}

func TestTenantDemoRecordRepositoryBatchLifecycleAndTenantScope(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewTenantDemoRecordRepository(database)
	records := []domain.TenantDemoRecord{
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, BatchID: "batch-1", RecordType: DemoRecordCourse, RecordID: "course-1"},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, BatchID: "batch-1", RecordType: DemoRecordUser, RecordID: "learner-1"},
	}
	if err := repo.RegisterBatch(context.Background(), records); err != nil {
		t.Fatalf("RegisterBatch() error = %v", err)
	}
	hasRecords, err := repo.HasRecords(context.Background(), "tenant-1")
	if err != nil || !hasRecords {
		t.Fatalf("HasRecords(tenant-1) = %v, %v", hasRecords, err)
	}
	foreignHasRecords, err := repo.HasRecords(context.Background(), "tenant-2")
	if err != nil || foreignHasRecords {
		t.Fatalf("HasRecords(tenant-2) = %v, %v", foreignHasRecords, err)
	}
	listed, err := repo.ListByTenant(context.Background(), "tenant-1")
	if err != nil || len(listed) != 2 {
		t.Fatalf("ListByTenant() = %#v, %v", listed, err)
	}
	if err := repo.DeleteBatch(context.Background(), "tenant-2", "batch-1"); err != nil {
		t.Fatalf("cross-tenant DeleteBatch() error = %v", err)
	}
	listed, err = repo.ListByTenant(context.Background(), "tenant-1")
	if err != nil || len(listed) != 2 {
		t.Fatalf("records after cross-tenant delete = %#v, %v", listed, err)
	}
	if err := repo.DeleteBatch(context.Background(), "tenant-1", "batch-1"); err != nil {
		t.Fatalf("DeleteBatch() error = %v", err)
	}
	hasRecords, err = repo.HasRecords(context.Background(), "tenant-1")
	if err != nil || hasRecords {
		t.Fatalf("HasRecords(after delete) = %v, %v", hasRecords, err)
	}
}
