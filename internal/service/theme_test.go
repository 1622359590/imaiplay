package service

import (
	"context"
	"strings"
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
	if theme.SelectedBackgroundColor != DefaultPrimaryColor ||
		theme.SelectedTextColor != "#ffffff" ||
		theme.SelectedIconColor != "#ffffff" {
		t.Fatalf("selection colors = %#v", theme)
	}
}

func TestTenantThemeDefaultsChooseHighestContrastText(t *testing.T) {
	theme := themeWithDefaults(&domain.Tenant{PrimaryColor: "#3582E1"})
	if theme.SelectedTextColor != "#000000" || theme.SelectedIconColor != "#000000" {
		t.Fatalf("selection colors = %s/%s, want highest-contrast black", theme.SelectedTextColor, theme.SelectedIconColor)
	}
}

func TestTenantThemeUpdatePersistsIndependentSelectionColors(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewTenantRepository(database)
	tenant := &domain.Tenant{ID: "tenant-selection", Code: "selection", Name: "Selection"}
	if err := repo.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := tenantcontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	themeService := NewTenantThemeService(repo)
	updated, err := themeService.Update(ctx, ThemeUpdate{
		PrimaryColor:            "#3582e1",
		SelectedBackgroundColor: "#fff1f0",
		SelectedTextColor:       "#c5221f",
		SelectedIconColor:       "#8c1d18",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.PrimaryColor != "#3582E1" ||
		updated.SelectedBackgroundColor != "#FFF1F0" ||
		updated.SelectedTextColor != "#C5221F" ||
		updated.SelectedIconColor != "#8C1D18" {
		t.Fatalf("updated theme = %#v", updated)
	}
	persisted, err := repo.FindByID(context.Background(), tenant.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.SelectedBackgroundColor != "#FFF1F0" ||
		persisted.SelectedTextColor != "#C5221F" ||
		persisted.SelectedIconColor != "#8C1D18" {
		t.Fatalf("persisted theme = %#v", persisted)
	}
}

func TestTenantThemeUpdateRejectsInvalidSelectionColors(t *testing.T) {
	_, tenants, _ := serviceRepositories(t)
	tenant := &domain.Tenant{ID: "theme-validation", Code: "theme-validation", Name: "Theme"}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := tenantcontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	themeService := NewTenantThemeService(tenants)
	tests := []struct {
		name  string
		field string
		value ThemeUpdate
	}{
		{name: "background", field: "selected_background_color", value: ThemeUpdate{SelectedBackgroundColor: "blue"}},
		{name: "text", field: "selected_text_color", value: ThemeUpdate{SelectedTextColor: "#123"}},
		{name: "icon", field: "selected_icon_color", value: ThemeUpdate{SelectedIconColor: "#12345678"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := themeService.Update(ctx, test.value)
			if errorCode(err) != 40000 || !strings.Contains(err.Error(), test.field) {
				t.Fatalf("Update() error = %#v, want bad request naming %s", err, test.field)
			}
		})
	}
}
