package migration

import (
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrateCreatesTenantAndUserTables(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if !database.Migrator().HasTable(&domain.Tenant{}) {
		t.Fatal("AutoMigrate() did not create tenants table")
	}
	if !database.Migrator().HasTable(&domain.User{}) {
		t.Fatal("AutoMigrate() did not create users table")
	}
	if !database.Migrator().HasConstraint(&domain.User{}, "Tenant") {
		t.Fatal("AutoMigrate() did not create users tenant foreign key")
	}
	for name, model := range map[string]interface{}{
		"courses":                 &domain.Course{},
		"course_chapters":         &domain.CourseChapter{},
		"course_lessons":          &domain.CourseLesson{},
		"course_enrollments":      &domain.CourseEnrollment{},
		"lesson_progress":         &domain.LessonProgress{},
		"resources":               &domain.Resource{},
		"resource_categories":     &domain.ResourceCategory{},
		"refresh_tokens":          &domain.RefreshToken{},
		"password_resets":         &domain.PasswordReset{},
		"audit_logs":              &domain.AuditLog{},
		"plans":                   &domain.Plan{},
		"tenant_official_courses": &domain.TenantOfficialCourse{},
		"login_challenges":        &domain.LoginChallenge{},
		"course_materials":        &domain.CourseMaterial{},
		"course_categories":       &domain.CourseCategory{},
		"learning_daily_stats":    &domain.LearningDailyStat{},
		"learning_time_reports":   &domain.LearningTimeReport{},
		"tenant_demo_records":     &domain.TenantDemoRecord{},
	} {
		if !database.Migrator().HasTable(model) {
			t.Fatalf("AutoMigrate() did not create %s table", name)
		}
	}
	if !database.Migrator().HasTable("schema_migrations") || !database.Migrator().HasColumn(&domain.CourseLesson{}, "resource_id") {
		t.Fatal("versioned migrations did not create schema metadata or resource_id")
	}
	if !database.Migrator().HasColumn(&domain.Tenant{}, "BrandName") {
		t.Fatal("AutoMigrate() did not create tenants.brand_name")
	}
	var count int64
	if err := database.Table("schema_migrations").Count(&count).Error; err != nil || count != 18 {
		t.Fatalf("schema migrations count = %d, err=%v", count, err)
	}
	for _, name := range []string{
		"idx_users_email_lookup",
		"idx_users_phone_lookup",
	} {
		if !database.Migrator().HasIndex(&domain.User{}, name) {
			t.Fatalf("AutoMigrate() did not create %s", name)
		}
	}
	if !database.Migrator().HasIndex(&domain.CourseMaterial{}, "idx_course_materials_course_resource") {
		t.Fatal("AutoMigrate() did not create idx_course_materials_course_resource")
	}
	if err := AutoMigrate(database); err != nil {
		t.Fatalf("repeat AutoMigrate() error = %v", err)
	}
	if err := database.Table("schema_migrations").Count(&count).Error; err != nil || count != 18 {
		t.Fatalf("repeat schema migrations count = %d, err=%v", count, err)
	}
}

func TestMigrationV16AddsLearningWorkspaceSchema(t *testing.T) {
	database := migrateTestDatabaseThroughV15(t)
	if err := database.Exec(
		"INSERT INTO course_enrollments (id, tenant_id, course_id, user_id, status) VALUES (?, ?, ?, ?, ?)",
		"legacy-enrollment", "tenant-1", "course-1", "learner-1", 1,
	).Error; err != nil {
		t.Fatal(err)
	}

	if err := AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() from v15 error = %v", err)
	}

	for name, check := range map[string]bool{
		"courses.category_id":                database.Migrator().HasColumn(&domain.Course{}, "CategoryID"),
		"course_enrollments.assignment_type": database.Migrator().HasColumn(&domain.CourseEnrollment{}, "AssignmentType"),
		"course_categories":                  database.Migrator().HasTable(&domain.CourseCategory{}),
		"learning_daily_stats":               database.Migrator().HasTable(&domain.LearningDailyStat{}),
		"learning_time_reports":              database.Migrator().HasTable(&domain.LearningTimeReport{}),
		"tenant_demo_records":                database.Migrator().HasTable(&domain.TenantDemoRecord{}),
		"idx_course_categories_scope_name":   database.Migrator().HasIndex(&domain.CourseCategory{}, "idx_course_categories_scope_name"),
		"idx_learning_daily_user_date":       database.Migrator().HasIndex(&domain.LearningDailyStat{}, "idx_learning_daily_user_date"),
		"idx_learning_report_idempotency":    database.Migrator().HasIndex(&domain.LearningTimeReport{}, "idx_learning_report_idempotency"),
		"idx_learning_daily_ranking":         database.Migrator().HasIndex(&domain.LearningDailyStat{}, "idx_learning_daily_ranking"),
		"idx_tenant_demo_record":             database.Migrator().HasIndex(&domain.TenantDemoRecord{}, "idx_tenant_demo_record"),
	} {
		if !check {
			t.Fatalf("migration v16 did not create %s", name)
		}
	}

	var enrollment domain.CourseEnrollment
	if err := database.First(&enrollment, "id = ?", "legacy-enrollment").Error; err != nil {
		t.Fatal(err)
	}
	if enrollment.AssignmentType != domain.AssignmentRequired {
		t.Fatalf("legacy assignment type = %q, want %q", enrollment.AssignmentType, domain.AssignmentRequired)
	}

	category := &domain.CourseCategory{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		Name:      "销售", NormalizedName: "销售",
	}
	if err := database.Create(category).Error; err != nil {
		t.Fatal(err)
	}
	duplicateCategory := *category
	duplicateCategory.ID = ""
	if err := database.Create(&duplicateCategory).Error; err == nil {
		t.Fatal("expected duplicate category to fail")
	}
	otherTenantCategory := *category
	otherTenantCategory.ID = ""
	otherTenantCategory.TenantID = "tenant-2"
	if err := database.Create(&otherTenantCategory).Error; err != nil {
		t.Fatalf("same category name in another tenant failed: %v", err)
	}

	daily := &domain.LearningDailyStat{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner-1", StudyDate: "2026-08-05", DurationSeconds: 30,
	}
	if err := database.Create(daily).Error; err != nil {
		t.Fatal(err)
	}
	duplicateDaily := *daily
	duplicateDaily.ID = ""
	if err := database.Create(&duplicateDaily).Error; err == nil {
		t.Fatal("expected duplicate daily stat to fail")
	}

	report := &domain.LearningTimeReport{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner-1", LessonID: "lesson-1", ReportID: "report-1", WatchedSecondsDelta: 15,
	}
	if err := database.Create(report).Error; err != nil {
		t.Fatal(err)
	}
	duplicateReport := *report
	duplicateReport.ID = ""
	if err := database.Create(&duplicateReport).Error; err == nil {
		t.Fatal("expected duplicate report to fail")
	}

	demoRecord := &domain.TenantDemoRecord{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		BatchID:   "batch-1", RecordType: "course", RecordID: "course-1",
	}
	if err := database.Create(demoRecord).Error; err != nil {
		t.Fatal(err)
	}
	duplicateDemoRecord := *demoRecord
	duplicateDemoRecord.ID = ""
	duplicateDemoRecord.BatchID = "batch-2"
	if err := database.Create(&duplicateDemoRecord).Error; err == nil {
		t.Fatal("expected duplicate demo record to fail")
	}
}

func TestMigrationV16IsIdempotent(t *testing.T) {
	database := migrateTestDatabaseThroughV15(t)
	if err := AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	optional := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  "course-optional", UserID: "learner-optional",
		AssignmentType: domain.AssignmentOptional,
	}
	if err := database.Create(optional).Error; err != nil {
		t.Fatal(err)
	}

	if err := migrateV16(database); err != nil {
		t.Fatalf("first direct migration v16 rerun error = %v", err)
	}
	if err := migrateV16(database); err != nil {
		t.Fatalf("second direct migration v16 rerun error = %v", err)
	}
	if err := AutoMigrate(database); err != nil {
		t.Fatalf("repeat AutoMigrate() error = %v", err)
	}

	var stored domain.CourseEnrollment
	if err := database.First(&stored, "id = ?", optional.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.AssignmentType != domain.AssignmentOptional {
		t.Fatalf("optional assignment changed to %q", stored.AssignmentType)
	}
	var count int64
	if err := database.Model(&schemaMigration{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 18 {
		t.Fatalf("schema migrations count = %d, want 18", count)
	}
}

func migrateTestDatabaseThroughV15(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&schemaMigration{}); err != nil {
		t.Fatal(err)
	}
	migrations := []migration{
		{Version: 1, Up: migrateV1},
		{Version: 2, Up: migrateV2},
		{Version: 3, Up: migrateV3},
		{Version: 4, Up: migrateV4},
		{Version: 5, Up: migrateV5},
		{Version: 6, Up: migrateV6},
		{Version: 7, Up: migrateV7},
		{Version: 8, Up: migrateV8},
		{Version: 9, Up: migrateV9},
		{Version: 10, Up: migrateV10},
		{Version: 11, Up: migrateV11},
		{Version: 12, Up: migrateV12},
		{Version: 13, Up: migrateV13},
		{Version: 14, Up: migrateV14},
		{Version: 15, Up: migrateV15},
	}
	for _, item := range migrations {
		if err := database.Transaction(func(tx *gorm.DB) error {
			if err := item.Up(tx); err != nil {
				return err
			}
			return tx.Create(&schemaMigration{Version: item.Version, AppliedAt: time.Now().UTC()}).Error
		}); err != nil {
			t.Fatalf("migrate database through v%d: %v", item.Version, err)
		}
	}
	for _, column := range []struct {
		model interface{}
		name  string
	}{
		{model: &domain.Course{}, name: "CategoryID"},
		{model: &domain.CourseEnrollment{}, name: "AssignmentType"},
	} {
		if database.Migrator().HasColumn(column.model, column.name) {
			if err := database.Migrator().DropColumn(column.model, column.name); err != nil {
				t.Fatalf("drop post-v15 column %s: %v", column.name, err)
			}
		}
	}
	return database
}

func TestAutoMigrateRepairsMissingActiveDefaultPlan(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&domain.Plan{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Where("version = ?", 15).Delete(&schemaMigration{}).Error; err != nil {
		t.Fatal(err)
	}

	if err := AutoMigrate(database); err != nil {
		t.Fatal(err)
	}

	var count int64
	if err := database.Model(&domain.Plan{}).Where("is_default = ? AND status = ?", true, 1).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("active default plan count = %d, want 1", count)
	}
}
