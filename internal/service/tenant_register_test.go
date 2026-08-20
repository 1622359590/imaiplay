package service

import (
	"context"
	"errors"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantRegistrationRegistersEverySeededDemoRecordInOneBatch(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	registration := NewTenantRegistrationService(database, "secret")
	result, err := registration.Register(context.Background(), "Registry", "admin@registry.test", "Admin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	records, err := repository.NewTenantDemoRecordRepository(database).ListByTenant(context.Background(), result.Tenant.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 12 {
		t.Fatalf("registered demo record count=%d, want 12: %#v", len(records), records)
	}
	var engagementStates []domain.LearnerEngagementState
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&engagementStates).Error; err != nil {
		t.Fatal(err)
	}
	if len(engagementStates) != 2 {
		t.Fatalf("demo learner engagement states=%d, want 2", len(engagementStates))
	}
	for _, state := range engagementStates {
		if state.FirstLoginAt != nil || state.WelcomeSeenAt != nil {
			t.Fatalf("demo learner should be new before first login: %#v", state)
		}
	}
	batches := map[string]bool{}
	types := map[string]int{}
	for _, record := range records {
		batches[record.BatchID] = true
		types[record.RecordType]++
	}
	if len(batches) != 1 {
		t.Fatalf("demo batches=%#v, want exactly one", batches)
	}
	wantTypes := map[string]int{
		repository.DemoRecordCourse: 1, repository.DemoRecordCourseChapter: 2,
		repository.DemoRecordCourseLesson: 4, repository.DemoRecordUser: 3,
		repository.DemoRecordResource: 2,
	}
	for recordType, want := range wantTypes {
		if types[recordType] != want {
			t.Fatalf("registered %s count=%d, want %d", recordType, types[recordType], want)
		}
	}
	wantIDs := make(map[string]map[string]bool)
	for recordType, model := range map[string]interface{}{
		repository.DemoRecordCourse:        &domain.Course{},
		repository.DemoRecordCourseChapter: &domain.CourseChapter{},
		repository.DemoRecordCourseLesson:  &domain.CourseLesson{},
		repository.DemoRecordUser:          &domain.User{},
		repository.DemoRecordResource:      &domain.Resource{},
	} {
		var ids []string
		query := database.Model(model).Where("tenant_id = ?", result.Tenant.ID)
		if recordType == repository.DemoRecordUser {
			query = query.Where("role IN ?", []string{"learner", "instructor"})
		}
		if err := query.Pluck("id", &ids).Error; err != nil {
			t.Fatal(err)
		}
		wantIDs[recordType] = make(map[string]bool, len(ids))
		for _, id := range ids {
			wantIDs[recordType][id] = true
		}
	}
	for _, record := range records {
		if !wantIDs[record.RecordType][record.RecordID] {
			t.Fatalf("registered demo record does not match a seeded row: %#v", record)
		}
		delete(wantIDs[record.RecordType], record.RecordID)
	}
	for recordType, ids := range wantIDs {
		if len(ids) != 0 {
			t.Fatalf("unregistered seeded %s IDs=%#v", recordType, ids)
		}
	}
}

func TestTenantRegistrationRollsBackSeedDataWhenDemoRegistrationFails(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	database.Callback().Create().Before("gorm:create").Register("fail_demo_registration", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "tenant_demo_records" {
			tx.AddError(errors.New("demo registry unavailable"))
		}
	})
	registration := NewTenantRegistrationService(database, "secret")
	if _, err := registration.Register(context.Background(), "Rollback", "admin@rollback.test", "Admin", "password123"); err == nil {
		t.Fatal("Register() error=nil, want demo registration failure")
	}
	for _, model := range []interface{}{&domain.Tenant{}, &domain.User{}, &domain.Course{}, &domain.Resource{}, &domain.TenantDemoRecord{}} {
		var count int64
		if err := database.Model(model).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%T count=%d after rollback, want 0", model, count)
		}
	}
}

func TestClearDemoDataDeletesOnlyRegisteredTenantIDs(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	registration := NewTenantRegistrationService(database, "secret")
	result, err := registration.Register(context.Background(), "Safe Clear", "admin@safe.test", "Admin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	var seededUsers []domain.User
	if err := database.Where("tenant_id = ? AND role IN ?", result.Tenant.ID, []string{"learner", "instructor"}).Find(&seededUsers).Error; err != nil {
		t.Fatal(err)
	}
	for index := range seededUsers {
		legacyEmail := seededUsers[index].Email
		if err := database.Model(&seededUsers[index]).Update("email", "registered-"+legacyEmail).Error; err != nil {
			t.Fatal(err)
		}
		businessUser := &domain.User{
			BaseModel: domain.BaseModel{TenantID: result.Tenant.ID}, Email: legacyEmail,
			Password: "business", Name: seededUsers[index].Name, Role: seededUsers[index].Role, Status: 1,
		}
		if err := database.Create(businessUser).Error; err != nil {
			t.Fatal(err)
		}
	}
	businessCourse := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		Title:     demoCourseTitle, CreatedBy: result.User.ID,
	}
	if err := database.Create(businessCourse).Error; err != nil {
		t.Fatal(err)
	}
	businessResources := []*domain.Resource{
		{BaseModel: domain.BaseModel{TenantID: result.Tenant.ID}, Name: demoImageName, ResourceType: "image", URL: "/business/image", CreatedBy: result.User.ID},
		{BaseModel: domain.BaseModel{TenantID: result.Tenant.ID}, Name: demoDocumentName, ResourceType: "document", URL: "/business/document", CreatedBy: result.User.ID},
	}
	for _, resource := range businessResources {
		if err := database.Create(resource).Error; err != nil {
			t.Fatal(err)
		}
	}
	foreignTenant := &domain.Tenant{Code: "foreign", Name: "Foreign", Status: 1}
	if err := database.Create(foreignTenant).Error; err != nil {
		t.Fatal(err)
	}
	foreignResource := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: foreignTenant.ID}, Name: demoImageName,
		ResourceType: "image", URL: "/foreign/image", CreatedBy: "foreign-admin",
	}
	if err := database.Create(foreignResource).Error; err != nil {
		t.Fatal(err)
	}
	var batchID string
	if err := database.Model(&domain.TenantDemoRecord{}).Where("tenant_id = ?", result.Tenant.ID).Limit(1).Pluck("batch_id", &batchID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&domain.TenantDemoRecord{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID}, BatchID: batchID,
		RecordType: repository.DemoRecordResource, RecordID: foreignResource.ID,
	}).Error; err != nil {
		t.Fatal(err)
	}
	adminCtx := usercontext.WithUser(
		context.Background(), result.User.ID, result.Tenant.ID, "admin@safe.test", "tenant_admin",
	)
	if err := registration.ClearDemoData(adminCtx); err != nil {
		t.Fatal(err)
	}
	var remainingEngagementStates int64
	if err := database.Model(&domain.LearnerEngagementState{}).
		Where("tenant_id = ?", result.Tenant.ID).Count(&remainingEngagementStates).Error; err != nil {
		t.Fatal(err)
	}
	if remainingEngagementStates != 0 {
		t.Fatalf("demo learner engagement states after clear=%d, want 0", remainingEngagementStates)
	}
	var remaining int64
	if err := database.Model(&domain.Course{}).Where("id = ?", businessCourse.ID).Count(&remaining).Error; err != nil || remaining != 1 {
		t.Fatalf("unregistered business course count=%d error=%v", remaining, err)
	}
	for _, resource := range append(businessResources, foreignResource) {
		if err := database.Model(&domain.Resource{}).Where("id = ?", resource.ID).Count(&remaining).Error; err != nil || remaining != 1 {
			t.Fatalf("preserved resource %s count=%d error=%v", resource.ID, remaining, err)
		}
	}
	for _, email := range []string{"learner1@example.com", "learner2@example.com", "instructor@example.com"} {
		if err := database.Model(&domain.User{}).Where("tenant_id = ? AND email = ?", result.Tenant.ID, email).Count(&remaining).Error; err != nil || remaining != 1 {
			t.Fatalf("unregistered business user %s count=%d error=%v", email, remaining, err)
		}
	}
	hasRecords, err := repository.NewTenantDemoRecordRepository(database).HasRecords(context.Background(), result.Tenant.ID)
	if err != nil || hasRecords {
		t.Fatalf("HasRecords(after clear)=%v, %v", hasRecords, err)
	}
}

func TestClearDemoDataRollsBackDeletesWhenRegistrationRemovalFails(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	registration := NewTenantRegistrationService(database, "secret")
	result, err := registration.Register(context.Background(), "Atomic Clear", "admin@atomic.test", "Admin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	database.Callback().Delete().Before("gorm:delete").Register("fail_demo_record_delete", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "tenant_demo_records" {
			tx.AddError(errors.New("demo registry delete failed"))
		}
	})
	adminCtx := usercontext.WithUser(context.Background(), result.User.ID, result.Tenant.ID, result.User.Email, "tenant_admin")
	if err := registration.ClearDemoData(adminCtx); err == nil {
		t.Fatal("ClearDemoData() error=nil, want registration removal failure")
	}
	for _, model := range []interface{}{&domain.Course{}, &domain.User{}, &domain.Resource{}, &domain.TenantDemoRecord{}} {
		var count int64
		if err := database.Model(model).Where("tenant_id = ?", result.Tenant.ID).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Fatalf("%T count=0 after rollback", model)
		}
	}
}

func TestTenantRegistrationCreatesTenantAdminAndDemoData(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	service := NewTenantRegistrationService(database, "secret")

	result, err := service.Register(context.Background(), "Acme 公司", "ADMIN@ACME.COM", "管理员", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if result.Tenant.Code != "acme" || result.User.Role != "tenant_admin" || result.Token == "" {
		t.Fatalf("registration result = %#v", result)
	}
	var users []domain.User
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&users).Error; err != nil {
		t.Fatalf("find users: %v", err)
	}
	if len(users) != 4 {
		t.Fatalf("user count = %d, want 4", len(users))
	}
	var courses []domain.Course
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&courses).Error; err != nil {
		t.Fatalf("find courses: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != demoCourseTitle {
		t.Fatalf("courses = %#v", courses)
	}
	var resources []domain.Resource
	if err := database.Where("tenant_id = ?", result.Tenant.ID).Find(&resources).Error; err != nil {
		t.Fatalf("find resources: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("resource count = %d, want 2", len(resources))
	}
}

func TestTenantRegistrationUsesUniqueSlugAndClearsDemoData(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	service := NewTenantRegistrationService(database, "secret")
	first, err := service.Register(context.Background(), "Acme Inc", "one@example.com", "One", "password123")
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	second, err := service.Register(context.Background(), "Acme Inc", "two@example.com", "Two", "password123")
	if err != nil {
		t.Fatalf("second register: %v", err)
	}
	if first.Tenant.Code != "acme-inc" || second.Tenant.Code == first.Tenant.Code {
		t.Fatalf("tenant codes = %q, %q", first.Tenant.Code, second.Tenant.Code)
	}
	ctx := withUserContext(context.Background(), first.Tenant.ID, first.User.ID, "tenant_admin")
	if err := service.ClearDemoData(ctx); err != nil {
		t.Fatalf("clear demo data: %v", err)
	}
	var count int64
	if err := database.Model(&domain.Course{}).Where("tenant_id = ?", first.Tenant.ID).Count(&count).Error; err != nil {
		t.Fatalf("count courses: %v", err)
	}
	if count != 0 {
		t.Fatalf("course count after clear = %d", count)
	}
}

func TestSuperadminCanCreateTenantWithoutTokenAndChoosePlan(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	plan := &domain.Plan{Name: "专业版", StorageQuotaBytes: 99, Status: 1}
	if err := database.Create(plan).Error; err != nil {
		t.Fatal(err)
	}
	service := NewTenantRegistrationService(database, "secret")
	ctx := usercontext.WithUser(context.Background(), "root", "", "root@example.com", "superadmin")
	result, err := service.CreateForSuperadmin(ctx, "代建客户", "admin@example.com", "", "客户管理员", "password123", plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != "" || result.Tenant.PlanID == nil || *result.Tenant.PlanID != plan.ID {
		t.Fatalf("result = %#v", result)
	}
	var users int64
	if err := database.Model(&domain.User{}).Where("tenant_id = ?", result.Tenant.ID).Count(&users).Error; err != nil {
		t.Fatal(err)
	}
	if users != 1 {
		t.Fatalf("user count = %d, want only the tenant admin", users)
	}
	var courses int64
	if err := database.Model(&domain.Course{}).Where("tenant_id = ?", result.Tenant.ID).Count(&courses).Error; err != nil {
		t.Fatal(err)
	}
	if courses != 0 {
		t.Fatalf("demo course count = %d, want 0", courses)
	}
}

func TestTenantCodeSlug(t *testing.T) {
	cases := map[string]string{"Acme Inc": "acme-inc", "  ACME!! ": "acme", "测试": "t-"}
	for input, prefix := range cases {
		if got := tenantCodeSlug(input); len(got) == 0 || (prefix != "t-" && got != prefix) || (prefix == "t-" && len(got) < 3) {
			t.Fatalf("tenantCodeSlug(%q) = %q", input, got)
		}
	}
}

func withUserContext(ctx context.Context, tenantID, userID, role string) context.Context {
	return usercontext.WithUser(ctx, userID, tenantID, userID+"@example.com", role)
}
