package migration

import (
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

func AutoMigrate(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.Tenant{}, &domain.User{}); err != nil {
		return err
	}
	return database.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email " +
			"ON users (tenant_id, email)",
	).Error
}
