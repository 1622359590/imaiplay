package migration

import (
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

func AutoMigrate(database *gorm.DB) error {
	if err := database.AutoMigrate(
		&domain.Tenant{},
		&domain.User{},
		&domain.Course{},
		&domain.CourseChapter{},
		&domain.CourseLesson{},
		&domain.CourseEnrollment{},
		&domain.LessonProgress{},
		&domain.ResourceCategory{},
		&domain.Resource{},
	); err != nil {
		return err
	}
	if err := database.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email " +
			"ON users (tenant_id, email)",
	).Error; err != nil {
		return err
	}
	if err := database.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_enrollments_tenant_course_user " +
			"ON course_enrollments (tenant_id, course_id, user_id)",
	).Error; err != nil {
		return err
	}
	return database.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_progress_tenant_user_lesson " +
			"ON lesson_progress (tenant_id, user_id, lesson_id)",
	).Error
}
