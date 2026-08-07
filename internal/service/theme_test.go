package service

import (
	"context"
	"testing"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTenantThemeUsesAuthenticatedTenant(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewTenantRepository(database)
	first := &domain.Tenant{ID: "tenant-one", Code: "one", Name: "One", PrimaryColor: "#111111"}
	second := &domain.Tenant{ID: "tenant-two", Code: "two", Name: "Two", PrimaryColor: "#222222"}
	if err := repo.Create(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	service := NewTenantThemeService(repo)
	ctx := tenantcontext.WithTenant(context.Background(), "two", tenantcontext.SourceHeaderCode)
	ctx = tenantcontext.WithUser(ctx, "user-one", first.ID, "admin@one.test", "tenant_admin")
	theme, err := service.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if theme.PrimaryColor != first.PrimaryColor {
		t.Fatalf("theme color = %s, want %s", theme.PrimaryColor, first.PrimaryColor)
	}
}

func TestTenantThemeDefaultsWhenUnset(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewTenantRepository(database)
	tenant := &domain.Tenant{ID: "tenant-default", Code: "default", Name: "Default"}
	if err := repo.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := tenantcontext.WithTenant(context.Background(), tenant.Code, tenantcontext.SourceHeaderCode)
	theme, err := NewTenantThemeService(repo).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if theme.PrimaryColor != DefaultPrimaryColor {
		t.Fatalf("theme color = %s, want default %s", theme.PrimaryColor, DefaultPrimaryColor)
	}
}

func TestTenantThemeBrandNameIsTrimmedAndTenantScoped(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewTenantRepository(database)
	first := &domain.Tenant{ID: "tenant-brand-one", Code: "brand-one", Name: "One"}
	second := &domain.Tenant{ID: "tenant-brand-two", Code: "brand-two", Name: "Two"}
	for _, tenant := range []*domain.Tenant{first, second} {
		if err := repo.Create(context.Background(), tenant); err != nil {
			t.Fatal(err)
		}
	}
	ctx := tenantcontext.WithUser(context.Background(), "admin", first.ID, "admin@one.test", "tenant_admin")
	theme, err := NewTenantThemeService(repo).Update(ctx, "#111111", "", "", "", "  Acme Academy  ")
	if err != nil || theme.BrandName != "Acme Academy" {
		t.Fatalf("Update() = %#v, %v", theme, err)
	}
	stored, err := repo.FindByID(context.Background(), first.ID)
	if err != nil || stored.BrandName != "Acme Academy" {
		t.Fatalf("stored brand = %q, err=%v", stored.BrandName, err)
	}
	other, err := repo.FindByID(context.Background(), second.ID)
	if err != nil || other.BrandName != "" {
		t.Fatalf("other tenant brand = %q, err=%v", other.BrandName, err)
	}
}
