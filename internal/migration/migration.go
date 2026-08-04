package migration

import (
	"sort"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type schemaMigration struct {
	Version   int       `gorm:"primaryKey"`
	AppliedAt time.Time `gorm:"not null"`
}

func (schemaMigration) TableName() string { return "schema_migrations" }

type migration struct {
	Version int
	Up      func(*gorm.DB) error
}

func AutoMigrate(database *gorm.DB) error {
	if err := database.AutoMigrate(&schemaMigration{}); err != nil {
		return err
	}
	registered := []migration{
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
	}
	sort.Slice(registered, func(i, j int) bool { return registered[i].Version < registered[j].Version })
	var applied []schemaMigration
	if err := database.Find(&applied).Error; err != nil {
		return err
	}
	done := make(map[int]struct{}, len(applied))
	for _, item := range applied {
		done[item.Version] = struct{}{}
	}
	for _, item := range registered {
		if _, ok := done[item.Version]; ok {
			continue
		}
		if err := database.Transaction(func(tx *gorm.DB) error {
			if err := item.Up(tx); err != nil {
				return err
			}
			return tx.Create(&schemaMigration{Version: item.Version, AppliedAt: time.Now().UTC()}).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

type courseLessonV1 struct {
	domain.BaseModel
	ChapterID       string `gorm:"index;not null"`
	Title           string `gorm:"not null"`
	ContentType     string `gorm:"not null"`
	ContentURL      string
	DurationSeconds int `gorm:"default:0"`
	SortOrder       int `gorm:"default:0"`
}

func (courseLessonV1) TableName() string { return "course_lessons" }

func migrateV1(database *gorm.DB) error {
	if err := database.AutoMigrate(
		&domain.Tenant{}, &domain.User{}, &domain.Course{}, &domain.CourseChapter{},
		&courseLessonV1{}, &domain.CourseEnrollment{}, &domain.LessonProgress{},
		&domain.ResourceCategory{}, &domain.Resource{}, &domain.RefreshToken{},
	); err != nil {
		return err
	}
	for _, statement := range []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email ON users (tenant_id, email)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_enrollments_tenant_course_user ON course_enrollments (tenant_id, course_id, user_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_progress_tenant_user_lesson ON lesson_progress (tenant_id, user_id, lesson_id)",
	} {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateV2(database *gorm.DB) error {
	if database.Migrator().HasColumn(&domain.CourseLesson{}, "resource_id") {
		return nil
	}
	return database.Migrator().AddColumn(&domain.CourseLesson{}, "ResourceID")
}

func migrateV3(database *gorm.DB) error {
	if !database.Migrator().HasColumn(&domain.User{}, "phone") {
		if err := database.Migrator().AddColumn(&domain.User{}, "Phone"); err != nil {
			return err
		}
	}
	if err := database.AutoMigrate(&domain.PasswordReset{}); err != nil {
		return err
	}
	return database.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_phone ON users (tenant_id, phone)").Error
}

func migrateV4(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.AuditLog{}); err != nil {
		return err
	}
	for _, statement := range []string{
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_created_at ON audit_logs (tenant_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_action_created_at ON audit_logs (action, created_at)",
	} {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateV5(database *gorm.DB) error {
	return database.AutoMigrate(&domain.Tenant{})
}

func migrateV6(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.Plan{}); err != nil {
		return err
	}
	if !database.Migrator().HasColumn(&domain.Tenant{}, "plan_id") {
		if err := database.Migrator().AddColumn(&domain.Tenant{}, "PlanID"); err != nil {
			return err
		}
	}
	var count int64
	if err := database.Model(&domain.Plan{}).Where("is_default = ?", true).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return database.Create(&domain.Plan{Name: "免费版", StorageQuotaBytes: 1024 * 1024 * 1024, Features: "{}", IsDefault: true, Status: 1}).Error
	}
	return nil
}

func migrateV7(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.Tenant{}); err != nil {
		return err
	}
	return database.Exec("UPDATE tenants SET lifecycle_status = 'active' WHERE lifecycle_status IS NULL OR lifecycle_status = ''").Error
}

// v8 records the storage configuration decision: secrets are kept in the
// encrypted local config file, so no storage table is required.
func migrateV8(_ *gorm.DB) error { return nil }

func migrateV9(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.Course{}, &domain.TenantOfficialCourse{}); err != nil {
		return err
	}
	return database.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_official_courses ON tenant_official_courses (tenant_id, course_id)").Error
}

func migrateV10(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.Tenant{}); err != nil {
		return err
	}
	return database.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_custom_domain ON tenants (custom_domain) WHERE custom_domain IS NOT NULL AND custom_domain <> ''").Error
}

func migrateV11(database *gorm.DB) error { return database.AutoMigrate(&domain.PasswordReset{}) }

// v12 allows the global superadmin to remain outside tenant scope. The
// application writes NULL for that user; tenant users retain the FK.
func migrateV12(database *gorm.DB) error {
	if database.Dialector.Name() == "sqlite" {
		if err := database.Migrator().AlterColumn(&nullableUserTenant{}, "TenantID"); err != nil {
			return err
		}
		for _, statement := range []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email ON users (tenant_id, email)",
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_phone ON users (tenant_id, phone)",
		} {
			if err := database.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if database.Dialector.Name() != "postgres" {
		return nil
	}
	return database.Exec("ALTER TABLE users ALTER COLUMN tenant_id DROP NOT NULL").Error
}

func migrateV13(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.LoginChallenge{}); err != nil {
		return err
	}
	for _, statement := range []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email_lookup ON users (LOWER(email))",
		"CREATE INDEX IF NOT EXISTS idx_users_phone_lookup ON users (phone)",
	} {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateV14(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.CourseMaterial{}); err != nil {
		return err
	}
	return database.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_course_materials_course_resource ON course_materials (course_id, resource_id)",
	).Error
}

type nullableUserTenant struct {
	TenantID *string `gorm:"column:tenant_id"`
}

func (nullableUserTenant) TableName() string { return "users" }
