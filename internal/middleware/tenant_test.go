package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
)

type tenantRepositoryStub struct{ tenant *domain.Tenant }

func (stub tenantRepositoryStub) Create(context.Context, *domain.Tenant) error { return nil }
func (stub tenantRepositoryStub) FindByID(context.Context, string) (*domain.Tenant, error) {
	return nil, errors.New("not implemented")
}
func (stub tenantRepositoryStub) FindByCode(context.Context, string) (*domain.Tenant, error) {
	return nil, errors.New("not implemented")
}
func (stub tenantRepositoryStub) FindByCustomDomain(_ context.Context, customDomain string) (*domain.Tenant, error) {
	if stub.tenant != nil && stub.tenant.CustomDomain != nil && customDomain == *stub.tenant.CustomDomain {
		return stub.tenant, nil
	}
	return nil, errors.New("not found")
}

func TestTenantWithRepositoryPrefersCustomDomainOverHeader(t *testing.T) {
	customDomain := "academy.example.com"
	router := gin.New()
	router.Use(TenantWithRepository(tenantRepositoryStub{
		tenant: &domain.Tenant{Code: "academy", CustomDomain: &customDomain},
	}))
	router.GET("/", func(c *gin.Context) {
		code, source := tenantcontext.TenantFromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"code": code, "source": source})
	})

	request := httptest.NewRequest(http.MethodGet, "http://academy.example.com/", nil)
	request.Host = customDomain
	request.Header.Set("X-Tenant-Code", "wrong-tenant")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	var got struct {
		Code   string `json:"code"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != "academy" || got.Source != tenantcontext.SourceCustomDomain {
		t.Fatalf("tenant = (%q, %q), want (%q, %q)", got.Code, got.Source, "academy", tenantcontext.SourceCustomDomain)
	}
}

func TestTenantWithRepositoryForAdminHostLeavesTenantUnknown(t *testing.T) {
	router := gin.New()
	router.Use(TenantWithRepositoryForAdminHost(tenantRepositoryStub{}, "play.imai.work"))
	router.GET("/", func(c *gin.Context) {
		code, source := tenantcontext.TenantFromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"code": code, "source": source})
	})

	request := httptest.NewRequest(http.MethodGet, "https://play.imai.work/", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	var got struct {
		Code   string `json:"code"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != tenantcontext.UnknownTenant || got.Source != tenantcontext.SourceUnknown {
		t.Fatalf("tenant = (%q, %q), want unknown tenant", got.Code, got.Source)
	}
}

func (stub tenantRepositoryStub) FindAll(context.Context) ([]domain.Tenant, error) {
	return nil, errors.New("not implemented")
}
func (stub tenantRepositoryStub) Update(context.Context, *domain.Tenant) error      { return nil }
func (stub tenantRepositoryStub) UpdateTheme(context.Context, *domain.Tenant) error { return nil }
func (stub tenantRepositoryStub) UpdatePlan(context.Context, *domain.Tenant) error  { return nil }
func (stub tenantRepositoryStub) Delete(context.Context, string) error              { return nil }

var _ repository.TenantRepository = tenantRepositoryStub{}

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
