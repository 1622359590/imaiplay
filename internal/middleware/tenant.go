package middleware

import (
	"net"
	"net/http"
	"strings"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
)

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		code, source := tenantFromRequest(c.Request)
		ctx := tenantcontext.WithTenant(c.Request.Context(), code, source)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func TenantWithRepository(tenants repository.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		code, source := tenantFromRequestWithRepository(c.Request, tenants)
		ctx := tenantcontext.WithTenant(c.Request.Context(), code, source)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// TenantWithRepositoryForAdminHost keeps the global admin host outside tenant
// resolution so superadmin authentication can run without a tenant.
func TenantWithRepositoryForAdminHost(tenants repository.TenantRepository, adminHost string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if adminHost != "" && requestHost(c.Request.Host) == requestHost(adminHost) {
			code, source := tenantcontext.UnknownTenant, tenantcontext.SourceUnknown
			if headerID := strings.TrimSpace(c.GetHeader("X-Tenant-ID")); headerID != "" {
				code, source = headerID, tenantcontext.SourceHeaderID
			} else if headerCode := strings.TrimSpace(c.GetHeader("X-Tenant-Code")); headerCode != "" {
				code, source = headerCode, tenantcontext.SourceHeaderCode
			}
			c.Request = c.Request.WithContext(tenantcontext.WithTenant(c.Request.Context(), code, source))
			c.Next()
			return
		}
		TenantWithRepository(tenants)(c)
	}
}

func tenantFromRequestWithRepository(request *http.Request, tenants repository.TenantRepository) (string, string) {
	if tenants != nil {
		if host := requestHost(request.Host); host != "" {
			if tenant, err := tenants.FindByCustomDomain(request.Context(), host); err == nil {
				return tenant.Code, tenantcontext.SourceCustomDomain
			}
		}
	}
	return tenantFromRequest(request)
}

func tenantFromRequest(request *http.Request) (string, string) {
	if code := subdomain(request.Host); code != "" {
		return code, tenantcontext.SourceSubdomain
	}
	if code := strings.TrimSpace(request.Header.Get("X-Tenant-ID")); code != "" {
		return code, tenantcontext.SourceHeaderID
	}
	if code := strings.TrimSpace(request.Header.Get("X-Tenant-Code")); code != "" {
		return code, tenantcontext.SourceHeaderCode
	}
	return tenantcontext.UnknownTenant, tenantcontext.SourceUnknown
}

func requestHost(raw string) string {
	host := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(raw)), ".")
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	if net.ParseIP(host) != nil {
		return ""
	}
	return host
}

func subdomain(rawHost string) string {
	host := strings.TrimSuffix(strings.ToLower(rawHost), ".")
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	if net.ParseIP(host) != nil {
		return ""
	}

	parts := strings.Split(host, ".")
	if len(parts) < 3 || parts[0] == "" {
		return ""
	}
	return parts[0]
}
