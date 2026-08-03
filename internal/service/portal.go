package service

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type Portal struct {
	TenantID         string `json:"tenant_id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	LogoURL          string `json:"logo_url"`
	PrimaryColor     string `json:"primary_color"`
	WelcomeText      string `json:"welcome_text"`
	DefaultPortalURL string `json:"default_portal_url"`
	CustomDomainURL  string `json:"custom_domain_url,omitempty"`
}

type PortalService struct {
	tenants      repository.TenantRepository
	platformHost string
}

func NewPortalService(
	tenants repository.TenantRepository,
	platformHost string,
) *PortalService {
	platformHost = normalizePortalHost(platformHost)
	if platformHost == "" {
		platformHost = "play.imai.work"
	}
	return &PortalService{tenants: tenants, platformHost: platformHost}
}

func (service *PortalService) Resolve(
	ctx context.Context,
	tenantCode string,
	host string,
) (*Portal, error) {
	if service.tenants == nil {
		return nil, errorsx.Internal("resolve tenant portal failed")
	}
	tenantCode = strings.ToLower(strings.TrimSpace(tenantCode))
	host = normalizePortalHost(host)
	var tenant *domain.Tenant
	var err error
	switch {
	case host != "" && host != service.platformHost:
		tenant, err = service.tenants.FindByCustomDomain(ctx, host)
	case tenantCode != "":
		tenant, err = service.tenants.FindByCode(ctx, tenantCode)
	default:
		return nil, errorsx.NotFound("tenant portal not found")
	}
	return service.portalFromResult(tenant, err)
}

func (service *PortalService) ResolveByTenantID(
	ctx context.Context,
	tenantID string,
) (*Portal, error) {
	if service.tenants == nil {
		return nil, errorsx.Internal("resolve tenant portal failed")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, errorsx.NotFound("tenant portal not found")
	}
	tenant, err := service.tenants.FindByID(ctx, tenantID)
	return service.portalFromResult(tenant, err)
}

func (service *PortalService) portalFromResult(
	tenant *domain.Tenant,
	err error,
) (*Portal, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.NotFound("tenant portal not found")
	}
	if err != nil {
		return nil, errorsx.Internal("resolve tenant portal failed")
	}
	if accessible, reason := TenantAccessible(
		tenant,
		time.Now().UTC(),
	); !accessible {
		return nil, errorsx.Forbidden(reason)
	}
	return service.portalFromTenant(tenant), nil
}

func (service *PortalService) portalFromTenant(tenant *domain.Tenant) *Portal {
	theme := themeWithDefaults(tenant)
	welcomeText := strings.TrimSpace(theme.WelcomeText)
	if welcomeText == "" {
		welcomeText = "欢迎来到 " + tenant.Name
	}
	portal := &Portal{
		TenantID:         tenant.ID,
		Code:             tenant.Code,
		Name:             tenant.Name,
		LogoURL:          strings.TrimSpace(theme.LogoURL),
		PrimaryColor:     theme.PrimaryColor,
		WelcomeText:      welcomeText,
		DefaultPortalURL: "https://" + service.platformHost + "/t/" + tenant.Code,
	}
	if tenant.CustomDomain != nil {
		customDomain := normalizePortalHost(*tenant.CustomDomain)
		if customDomain != "" && customDomain != service.platformHost {
			portal.CustomDomainURL = "https://" + customDomain
		}
	}
	return portal
}

func normalizePortalHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "://") {
		if parsed, err := url.Parse(raw); err == nil {
			raw = parsed.Host
		}
	}
	host := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(raw), "."))
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	return strings.Trim(host, "[]")
}
