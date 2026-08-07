package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

const DefaultPrimaryColor = "#4F46E5"

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type TenantThemeService struct{ tenants repository.TenantRepository }

type ThemeUpdate struct {
	PrimaryColor            string
	LogoURL                 string
	WelcomeText             string
	BrowserTitle            string
	SelectedBackgroundColor string
	SelectedTextColor       string
	SelectedIconColor       string
}

func NewTenantThemeService(tenants repository.TenantRepository) *TenantThemeService {
	return &TenantThemeService{tenants: tenants}
}

func (service *TenantThemeService) Get(ctx context.Context) (*domain.Tenant, error) {
	_, tenantID, _, _, authenticated := tenantcontext.UserFromContext(ctx)
	code, _ := tenantcontext.TenantFromContext(ctx)
	var tenant *domain.Tenant
	var err error
	if authenticated && tenantID != "" {
		tenant, err = service.tenants.FindByID(ctx, tenantID)
	} else {
		if code == "" || code == tenantcontext.UnknownTenant {
			return defaultTheme(), nil
		}
		tenant, err = service.tenants.FindByCode(ctx, code)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return defaultTheme(), nil
		}
		return nil, errorsx.Internal("load tenant theme failed")
	}
	return themeWithDefaults(tenant), nil
}

func (service *TenantThemeService) Update(ctx context.Context, update ThemeUpdate) (*domain.Tenant, error) {
	_, tenantID, _, role, ok := tenantcontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" {
		return nil, errorsx.Forbidden("permission denied")
	}
	colors := []struct {
		name  string
		value *string
	}{
		{name: "primary_color", value: &update.PrimaryColor},
		{name: "selected_background_color", value: &update.SelectedBackgroundColor},
		{name: "selected_text_color", value: &update.SelectedTextColor},
		{name: "selected_icon_color", value: &update.SelectedIconColor},
	}
	for _, color := range colors {
		*color.value = strings.TrimSpace(*color.value)
		if *color.value != "" && !hexColorPattern.MatchString(*color.value) {
			return nil, errorsx.BadRequest(color.name + " must be a six-digit hex color")
		}
		*color.value = strings.ToUpper(*color.value)
	}
	if len(update.LogoURL) > 500 || len(update.WelcomeText) > 255 || len(update.BrowserTitle) > 255 {
		return nil, errorsx.BadRequest("theme value is too long")
	}
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	tenant.PrimaryColor = update.PrimaryColor
	tenant.SelectedBackgroundColor = update.SelectedBackgroundColor
	tenant.SelectedTextColor = update.SelectedTextColor
	tenant.SelectedIconColor = update.SelectedIconColor
	tenant.LogoURL = strings.TrimSpace(update.LogoURL)
	tenant.WelcomeText = strings.TrimSpace(update.WelcomeText)
	tenant.BrowserTitle = strings.TrimSpace(update.BrowserTitle)
	if err := service.tenants.UpdateTheme(ctx, tenant); err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	return themeWithDefaults(tenant), nil
}

func defaultTheme() *domain.Tenant {
	return themeWithDefaults(&domain.Tenant{PrimaryColor: DefaultPrimaryColor})
}

func themeWithDefaults(tenant *domain.Tenant) *domain.Tenant {
	copy := *tenant
	if copy.PrimaryColor == "" || !hexColorPattern.MatchString(copy.PrimaryColor) {
		copy.PrimaryColor = DefaultPrimaryColor
	}
	copy.PrimaryColor = strings.ToUpper(copy.PrimaryColor)
	if copy.SelectedBackgroundColor == "" || !hexColorPattern.MatchString(copy.SelectedBackgroundColor) {
		copy.SelectedBackgroundColor = copy.PrimaryColor
	}
	copy.SelectedBackgroundColor = strings.ToUpper(copy.SelectedBackgroundColor)
	if copy.SelectedTextColor == "" || !hexColorPattern.MatchString(copy.SelectedTextColor) {
		copy.SelectedTextColor = highestContrastText(copy.SelectedBackgroundColor)
	} else {
		copy.SelectedTextColor = strings.ToUpper(copy.SelectedTextColor)
	}
	if copy.SelectedIconColor == "" || !hexColorPattern.MatchString(copy.SelectedIconColor) {
		copy.SelectedIconColor = copy.SelectedTextColor
	} else {
		copy.SelectedIconColor = strings.ToUpper(copy.SelectedIconColor)
	}
	return &copy
}

func highestContrastText(background string) string {
	r, g, b := hexRGB(background)
	luminance := 0.2126*linearColor(r) + 0.7152*linearColor(g) + 0.0722*linearColor(b)
	if (luminance+0.05)/0.05 >= 1.05/(luminance+0.05) {
		return "#000000"
	}
	return "#ffffff"
}

func hexRGB(color string) (float64, float64, float64) {
	values := make([]float64, 3)
	for index := range values {
		value, err := strconv.ParseUint(color[1+index*2:3+index*2], 16, 8)
		if err != nil {
			panic(fmt.Sprintf("invalid normalized color %q", color))
		}
		values[index] = float64(value) / 255
	}
	return values[0], values[1], values[2]
}

func linearColor(value float64) float64 {
	if value <= 0.04045 {
		return value / 12.92
	}
	return ((value + 0.055) / 1.055) * ((value + 0.055) / 1.055) * ((value + 0.055) / 1.055)
}
