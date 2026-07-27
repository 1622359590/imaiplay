package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/gin-gonic/gin"
)

func TestTenantIdentification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		host       string
		headers    map[string]string
		wantCode   string
		wantSource string
	}{
		{
			name:       "subdomain",
			host:       "tenant1.imaiplay.local",
			wantCode:   "tenant1",
			wantSource: tenantcontext.SourceSubdomain,
		},
		{
			name:       "subdomain with port",
			host:       "tenant1.imaiplay.local:8080",
			wantCode:   "tenant1",
			wantSource: tenantcontext.SourceSubdomain,
		},
		{
			name:       "ip address",
			host:       "127.0.0.1",
			wantCode:   tenantcontext.UnknownTenant,
			wantSource: tenantcontext.SourceUnknown,
		},
		{
			name:       "base domain",
			host:       "imaiplay.local",
			wantCode:   tenantcontext.UnknownTenant,
			wantSource: tenantcontext.SourceUnknown,
		},
		{
			name:       "tenant id header",
			host:       "localhost",
			headers:    map[string]string{"X-Tenant-ID": "tenant2"},
			wantCode:   "tenant2",
			wantSource: tenantcontext.SourceHeaderID,
		},
		{
			name:       "tenant code header",
			host:       "localhost",
			headers:    map[string]string{"X-Tenant-Code": "tenant3"},
			wantCode:   "tenant3",
			wantSource: tenantcontext.SourceHeaderCode,
		},
		{
			name:       "unknown",
			host:       "localhost",
			wantCode:   tenantcontext.UnknownTenant,
			wantSource: tenantcontext.SourceUnknown,
		},
		{
			name: "source priority",
			host: "tenant1.imaiplay.local",
			headers: map[string]string{
				"X-Tenant-ID":   "tenant2",
				"X-Tenant-Code": "tenant3",
			},
			wantCode:   "tenant1",
			wantSource: tenantcontext.SourceSubdomain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Tenant())
			router.GET("/", func(c *gin.Context) {
				code, source := tenantcontext.TenantFromContext(c.Request.Context())
				c.JSON(http.StatusOK, gin.H{"code": code, "source": source})
			})

			request := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			request.Host = tt.host
			for key, value := range tt.headers {
				request.Header.Set(key, value)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			var got struct {
				Code   string `json:"code"`
				Source string `json:"source"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Code != tt.wantCode || got.Source != tt.wantSource {
				t.Fatalf("tenant = (%q, %q), want (%q, %q)",
					got.Code, got.Source, tt.wantCode, tt.wantSource)
			}
		})
	}
}
