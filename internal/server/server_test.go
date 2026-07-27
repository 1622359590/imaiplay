package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/1622359590/imaiplay/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHealthIncludesApplicationAndTenant(t *testing.T) {
	router := New(config.Config{
		ServerPort: "8080",
		AppName:    "imaiplay",
		AppVersion: "0.1.0",
	}, func() error { return nil }, Dependencies{})

	request := httptest.NewRequest(http.MethodGet, "http://localhost/health", nil)
	request.Host = "tenant1.imaiplay.local"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var got struct {
		Status  string `json:"status"`
		AppName string `json:"app_name"`
		Version string `json:"version"`
		Tenant  struct {
			Code   string `json:"code"`
			Source string `json:"source"`
		} `json:"tenant"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != "ok" || got.AppName != "imaiplay" || got.Version != "0.1.0" {
		t.Fatalf("application fields = %#v", got)
	}
	if got.Tenant.Code != "tenant1" || got.Tenant.Source != "subdomain" {
		t.Fatalf("tenant = %#v", got.Tenant)
	}
}

func TestDatabaseHealth(t *testing.T) {
	tests := []struct {
		name         string
		dbCheck      func() error
		wantHTTPCode int
		wantStatus   string
		wantDatabase string
	}{
		{
			name:         "connected",
			dbCheck:      func() error { return nil },
			wantHTTPCode: http.StatusOK,
			wantStatus:   "ok",
			wantDatabase: "connected",
		},
		{
			name:         "disconnected",
			dbCheck:      func() error { return errors.New("connection lost") },
			wantHTTPCode: http.StatusServiceUnavailable,
			wantStatus:   "error",
			wantDatabase: "disconnected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := New(config.Config{}, tt.dbCheck, Dependencies{})
			request := httptest.NewRequest(http.MethodGet, "http://localhost/health/db", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantHTTPCode {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantHTTPCode)
			}
			var got struct {
				Status   string `json:"status"`
				Database string `json:"database"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Status != tt.wantStatus || got.Database != tt.wantDatabase {
				t.Fatalf("response = %#v, want status=%q database=%q",
					got, tt.wantStatus, tt.wantDatabase)
			}
		})
	}
}

func TestBackendRoutesRequireJWTAndRole(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	tenantRepo := repository.NewTenantRepository(database)
	userRepo := repository.NewUserRepository(database)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	deps := Dependencies{
		AuthService:   service.NewAuthService(userRepo, tenantRepo, "secret"),
		TenantService: service.NewTenantService(tenantRepo),
		UserService:   service.NewUserService(userRepo),
	}
	router := New(config.Config{JWTSecret: "secret"}, func() error { return nil }, deps)

	assertRouteStatus(t, router, "/backend/v1/tenants", "", http.StatusUnauthorized)
	tenantAdminToken, err := security.GenerateToken(
		"admin", tenant.ID, "admin@example.com", "tenant_admin", "secret",
	)
	if err != nil {
		t.Fatalf("generate tenant admin token: %v", err)
	}
	assertRouteStatus(
		t, router, "/backend/v1/tenants", tenantAdminToken, http.StatusForbidden,
	)
	superadminToken, err := security.GenerateToken(
		"root", "", "root@example.com", "superadmin", "secret",
	)
	if err != nil {
		t.Fatalf("generate superadmin token: %v", err)
	}
	assertRouteStatus(
		t, router, "/backend/v1/tenants", superadminToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/backend/v1/users", tenantAdminToken, http.StatusOK,
	)
}

func assertRouteStatus(
	t *testing.T,
	handler http.Handler,
	path, token string,
	want int,
) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Host = "acme.imaiplay.local"
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s status = %d, want %d body=%s",
			path, response.Code, want, response.Body.String())
	}
}
