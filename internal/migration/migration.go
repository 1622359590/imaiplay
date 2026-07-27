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
