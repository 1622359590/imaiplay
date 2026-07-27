package middleware

import (
	"net"
	"net/http"
	"strings"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
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
