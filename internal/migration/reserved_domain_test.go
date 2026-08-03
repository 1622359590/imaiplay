package migration

import (
	"context"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestClearReservedTenantDomainOnlyClearsMatchingTenant(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&domain.Tenant{}); err != nil {
		t.Fatal(err)
	}
	reserved := "PLAY.IMAI.WORK."
	custom := "learn.example.com"
	tenants := []*domain.Tenant{
		{Code: "reserved", Name: "Reserved", Status: 1, CustomDomain: &reserved},
		{Code: "custom", Name: "Custom", Status: 1, CustomDomain: &custom},
		{Code: "empty", Name: "Empty", Status: 1},
	}
	for _, tenant := range tenants {
		if err := database.Create(tenant).Error; err != nil {
			t.Fatal(err)
		}
	}

	affected, err := ClearReservedTenantDomain(
		database,
		" play.imai.work ",
	)
	if err != nil || affected != 1 {
		t.Fatalf("affected=%d err=%v", affected, err)
	}

	var got []domain.Tenant
	if err := database.WithContext(context.Background()).
		Order("code").
		Find(&got).Error; err != nil {
		t.Fatal(err)
	}
	for _, tenant := range got {
		switch tenant.Code {
		case "reserved":
			if tenant.CustomDomain != nil {
				t.Fatalf("reserved domain=%v want=nil", tenant.CustomDomain)
			}
		case "custom":
			if tenant.CustomDomain == nil ||
				*tenant.CustomDomain != custom {
				t.Fatalf("custom domain=%v want=%q", tenant.CustomDomain, custom)
			}
		}
	}
}

func TestClearReservedTenantDomainIgnoresEmptyReservedDomain(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&domain.Tenant{}); err != nil {
		t.Fatal(err)
	}
	affected, err := ClearReservedTenantDomain(database, " ")
	if err != nil || affected != 0 {
		t.Fatalf("affected=%d err=%v", affected, err)
	}
}
