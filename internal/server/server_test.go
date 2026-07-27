package server

import (
	"encoding/json"
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
	})

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
