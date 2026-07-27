package migration

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrateCreatesTenantTable(t *testing.T) {
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
}
