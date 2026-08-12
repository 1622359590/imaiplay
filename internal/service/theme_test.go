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

func TestTenantThemeUpdateNormalizesCrossClientThemeContract(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.NewTenantRepository(database)
	themeService := NewTenantThemeService(repo)

	tests := []struct {
		name  string
		input ThemeUpdate
		want  ThemeUpdate
	}{
		{
			name: "normalizes every shared theme field",
			input: ThemeUpdate{
				PrimaryColor:            " #3582e1 ",
				SelectedBackgroundColor: " #fff1f0 ",
				SelectedTextColor:       " #c5221f ",
				SelectedIconColor:       " #8c1d18 ",
				LogoURL:                 " https://cdn.example.test/logo.svg ",
				WelcomeText:             " 欢迎来到 Acme 学院 ",
				BrowserTitle:            " Acme Learning ",
				BrandName:               " Acme Academy ",
			},
			want: ThemeUpdate{
				PrimaryColor:            "#3582E1",
				SelectedBackgroundColor: "#FFF1F0",
				SelectedTextColor:       "#C5221F",
				SelectedIconColor:       "#8C1D18",
				LogoURL:                 "https://cdn.example.test/logo.svg",
				WelcomeText:             "欢迎来到 Acme 学院",
				BrowserTitle:            "Acme Learning",
				BrandName:               "Acme Academy",
			},
		},
		{
			name: "keeps optional brand fields empty",
			input: ThemeUpdate{
				PrimaryColor:            "#123456",
				SelectedBackgroundColor: "#654321",
				SelectedTextColor:       "#FFFFFF",
				SelectedIconColor:       "#000000",
				LogoURL:                 "  ",
				WelcomeText:             "\t",
				BrowserTitle:            "\n",
				BrandName:               "  ",
			},
			want: ThemeUpdate{
				PrimaryColor:            "#123456",
				SelectedBackgroundColor: "#654321",
				SelectedTextColor:       "#FFFFFF",
				SelectedIconColor:       "#000000",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tenant := &domain.Tenant{ID: "tenant-" + strings.ReplaceAll(test.name, " ", "-"), Code: "theme-" + strings.ReplaceAll(test.name, " ", "-"), Name: "Theme"}
			if err := repo.Create(context.Background(), tenant); err != nil {
				t.Fatal(err)
			}
			ctx := tenantcontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
			updated, err := themeService.Update(ctx, test.input)
			if err != nil {
				t.Fatal(err)
			}
			assertThemeUpdate(t, "updated", themeUpdateFromTenant(updated), test.want)
			persisted, err := repo.FindByID(context.Background(), tenant.ID)
			if err != nil {
				t.Fatal(err)
			}
			assertThemeUpdate(t, "persisted", themeUpdateFromTenant(persisted), test.want)
			got, err := themeService.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			assertThemeUpdate(t, "get", themeUpdateFromTenant(got), test.want)
		})
	}
}

func TestTenantThemeUpdateRejectsInvalidCrossClientColors(t *testing.T) {
	_, tenants, _ := serviceRepositories(t)
	tenant := &domain.Tenant{ID: "theme-contract-validation", Code: "theme-contract-validation", Name: "Theme"}
	if err := tenants.Create(context.Background(), tenant); err != nil {
		t.Fatal(err)
	}
	ctx := tenantcontext.WithUser(context.Background(), "admin", tenant.ID, "admin@example.com", "tenant_admin")
	themeService := NewTenantThemeService(tenants)
	tests := []struct {
		name  string
		field string
		input ThemeUpdate
	}{
		{name: "primary", field: "primary_color", input: ThemeUpdate{PrimaryColor: "blue"}},
		{name: "selected background", field: "selected_background_color", input: ThemeUpdate{SelectedBackgroundColor: "#123"}},
		{name: "selected text", field: "selected_text_color", input: ThemeUpdate{SelectedTextColor: "#12345678"}},
		{name: "selected icon", field: "selected_icon_color", input: ThemeUpdate{SelectedIconColor: "nope"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := themeService.Update(ctx, test.input)
			if errorCode(err) != 40000 || !strings.Contains(err.Error(), test.field) {
				t.Fatalf("Update() error = %#v, want bad request naming %s", err, test.field)
			}
		})
	}
}

func themeUpdateFromTenant(theme *domain.Tenant) ThemeUpdate {
	return ThemeUpdate{
		PrimaryColor:            theme.PrimaryColor,
		SelectedBackgroundColor: theme.SelectedBackgroundColor,
		SelectedTextColor:       theme.SelectedTextColor,
		SelectedIconColor:       theme.SelectedIconColor,
		LogoURL:                 theme.LogoURL,
		WelcomeText:             theme.WelcomeText,
		BrowserTitle:            theme.BrowserTitle,
		BrandName:               theme.BrandName,
	}
}

func assertThemeUpdate(t *testing.T, subject string, got, want ThemeUpdate) {
	t.Helper()
	if got != want {
		t.Fatalf("%s theme = %#v, want %#v", subject, got, want)
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
	theme, err := NewTenantThemeService(repo).Update(ctx, ThemeUpdate{PrimaryColor: "#111111", BrandName: "  Acme Academy  "})
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
