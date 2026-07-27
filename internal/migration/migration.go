package migration

import (
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

func AutoMigrate(database *gorm.DB) error {
	return database.AutoMigrate(&domain.Tenant{})
}
