package service

import (
	"context"
	"regexp"
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

func (service *TenantThemeService) Update(ctx context.Context, primaryColor, logoURL, welcomeText, browserTitle, brandName string) (*domain.Tenant, error) {
	_, tenantID, _, role, ok := tenantcontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" {
		return nil, errorsx.Forbidden("permission denied")
	}
	primaryColor = strings.TrimSpace(primaryColor)
	logoURL = strings.TrimSpace(logoURL)
	welcomeText = strings.TrimSpace(welcomeText)
	browserTitle = strings.TrimSpace(browserTitle)
	brandName = strings.TrimSpace(brandName)
	if primaryColor != "" && !hexColorPattern.MatchString(primaryColor) {
		return nil, errorsx.BadRequest("primary_color must be a six-digit hex color")
	}
	if len(logoURL) > 500 || len(welcomeText) > 255 || len(browserTitle) > 255 || len([]rune(brandName)) > 50 {
		return nil, errorsx.BadRequest("theme value is too long")
	}
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	tenant.PrimaryColor, tenant.LogoURL, tenant.WelcomeText, tenant.BrowserTitle, tenant.BrandName = primaryColor, logoURL, welcomeText, browserTitle, brandName
	if err := service.tenants.UpdateTheme(ctx, tenant); err != nil {
		return nil, mapNotFound(err, "tenant not found")
	}
	return themeWithDefaults(tenant), nil
}

func defaultTheme() *domain.Tenant { return &domain.Tenant{PrimaryColor: DefaultPrimaryColor} }

func themeWithDefaults(tenant *domain.Tenant) *domain.Tenant {
	copy := *tenant
	if copy.PrimaryColor == "" || !hexColorPattern.MatchString(copy.PrimaryColor) {
		copy.PrimaryColor = DefaultPrimaryColor
	}
	return &copy
}
