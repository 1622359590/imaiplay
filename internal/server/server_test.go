package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1622359590/imaiplay/internal/config"
)

func TestHealthIncludesApplicationAndTenant(t *testing.T) {
	router := New(config.Config{
		ServerPort: "8080",
		AppName:    "imaiplay",
		AppVersion: "0.1.0",
	}, func() error { return nil })

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
			router := New(config.Config{}, tt.dbCheck)
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
