package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type portalResolverAPIStub struct {
	portal           *service.Portal
	err              error
	sessionPortal    *service.Portal
	sessionErr       error
	resolvedTenantID string
}

func (stub portalResolverAPIStub) Resolve(
	context.Context,
	string,
	string,
) (*service.Portal, error) {
	return stub.portal, stub.err
}

func (stub *portalResolverAPIStub) ResolveByTenantID(
	_ context.Context,
	tenantID string,
) (*service.Portal, error) {
	stub.resolvedTenantID = tenantID
	return stub.sessionPortal, stub.sessionErr
}

func TestPortalHandlerReturnsPublicMetadata(t *testing.T) {
	router := gin.New()
	handler := NewPortalHandler(&portalResolverAPIStub{
		portal: &service.Portal{
			TenantID:                "tenant-acme",
			Code:                    "acme",
			Name:                    "Acme",
			SelectedBackgroundColor: "#FFF1F0",
			SelectedTextColor:       "#C5221F",
			SelectedIconColor:       "#8C1D18",
			DefaultPortalURL:        "https://play.imai.work/t/acme",
		},
	})
	router.GET("/api/v1/portal", handler.Get)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/portal?tenant_code=acme",
		nil,
	)
	request.Host = "play.imai.work"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data["tenant_id"] != "tenant-acme" || body.Data["code"] != "acme" ||
		body.Data["default_portal_url"] != "https://play.imai.work/t/acme" {
		t.Fatalf("body=%#v", body)
	}
	if body.Data["selected_background_color"] != "#FFF1F0" ||
		body.Data["selected_text_color"] != "#C5221F" ||
		body.Data["selected_icon_color"] != "#8C1D18" {
		t.Fatalf("selection colors=%#v", body)
	}
}

func TestPortalHandlerPreservesCrossClientThemeContract(t *testing.T) {
	tests := []struct {
		name   string
		portal *service.Portal
		want   map[string]string
	}{
		{
			name: "populated brand fields",
			portal: &service.Portal{
				TenantID:                "tenant-acme",
				Code:                    "acme",
				Name:                    "Acme",
				PrimaryColor:            "#3582E1",
				SelectedBackgroundColor: "#FFF1F0",
				SelectedTextColor:       "#C5221F",
				SelectedIconColor:       "#8C1D18",
				LogoURL:                 "https://cdn.example.test/logo.svg",
				WelcomeText:             "欢迎来到 Acme 学院",
				BrowserTitle:            "Acme Learning",
				BrandName:               "Acme Academy",
			},
			want: map[string]string{
				"primary_color":             "#3582E1",
				"selected_background_color": "#FFF1F0",
				"selected_text_color":       "#C5221F",
				"selected_icon_color":       "#8C1D18",
				"logo_url":                  "https://cdn.example.test/logo.svg",
				"welcome_text":              "欢迎来到 Acme 学院",
				"browser_title":             "Acme Learning",
				"brand_name":                "Acme Academy",
			},
		},
		{
			name: "empty optional brand fields",
			portal: &service.Portal{
				TenantID:                "tenant-empty",
				Code:                    "empty",
				Name:                    "Empty",
				PrimaryColor:            "#123456",
				SelectedBackgroundColor: "#654321",
				SelectedTextColor:       "#FFFFFF",
				SelectedIconColor:       "#000000",
			},
			want: map[string]string{
				"primary_color":             "#123456",
				"selected_background_color": "#654321",
				"selected_text_color":       "#FFFFFF",
				"selected_icon_color":       "#000000",
				"logo_url":                  "",
				"welcome_text":              "",
				"browser_title":             "",
				"brand_name":                "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/api/v1/portal", NewPortalHandler(&portalResolverAPIStub{portal: test.portal}).Get)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/portal?tenant_code="+test.portal.Code, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			var body struct {
				Data map[string]string `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			for field, want := range test.want {
				if got := body.Data[field]; got != want {
					t.Fatalf("%s = %q, want %q; body=%s", field, got, want, response.Body.String())
				}
			}
		})
	}
}

func TestPortalHandlerMapsResolutionErrors(t *testing.T) {
	router := gin.New()
	handler := NewPortalHandler(&portalResolverAPIStub{
		err: errorsx.NotFound("tenant portal not found"),
	})
	router.GET("/api/v1/portal", handler.Get)

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/portal", nil),
	)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestPortalHandlerReturnsStablePortalErrorContract(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "not found",
			err:        errorsx.NotFound("tenant portal not found"),
			wantStatus: http.StatusNotFound,
			wantError:  "portal_not_found",
		},
		{
			name:       "suspended",
			err:        errorsx.Forbidden("tenant is suspended"),
			wantStatus: http.StatusForbidden,
			wantError:  "portal_suspended",
		},
		{
			name:       "trial expired",
			err:        errorsx.Forbidden("tenant trial expired"),
			wantStatus: http.StatusForbidden,
			wantError:  "portal_trial_expired",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			handler := NewPortalHandler(&portalResolverAPIStub{err: test.err})
			router.GET("/api/v1/portal", handler.Get)

			response := httptest.NewRecorder()
			router.ServeHTTP(
				response,
				httptest.NewRequest(http.MethodGet, "/api/v1/portal", nil),
			)
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error != test.wantError {
				t.Fatalf("error=%q want=%q body=%s", body.Error, test.wantError, response.Body.String())
			}
		})
	}
}

func TestPortalSessionHandlerReturnsAuthenticatedLearnerPortal(t *testing.T) {
	resolver := &portalResolverAPIStub{
		sessionPortal: &service.Portal{
			TenantID: "tenant-acme", Code: "acme", Name: "Acme",
		},
	}
	router := gin.New()
	router.Use(asRole("learner", "tenant-acme"))
	router.GET("/api/v1/portal/session", NewPortalHandler(resolver).GetSession)

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/portal/session", nil),
	)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if resolver.resolvedTenantID != "tenant-acme" {
		t.Fatalf("resolved tenant=%q", resolver.resolvedTenantID)
	}
	if !strings.Contains(response.Body.String(), `"code":"acme"`) {
		t.Fatalf("body=%s", response.Body.String())
	}
}

func TestPortalSessionHandlerRejectsRoleOrMissingTenant(t *testing.T) {
	for _, test := range []struct {
		name     string
		role     string
		tenantID string
	}{
		{name: "superadmin", role: "superadmin"},
		{name: "learner without tenant", role: "learner"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolver := &portalResolverAPIStub{}
			router := gin.New()
			router.Use(asRole(test.role, test.tenantID))
			router.GET(
				"/api/v1/portal/session",
				NewPortalHandler(resolver).GetSession,
			)

			response := httptest.NewRecorder()
			router.ServeHTTP(
				response,
				httptest.NewRequest(http.MethodGet, "/api/v1/portal/session", nil),
			)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if resolver.resolvedTenantID != "" {
				t.Fatalf("resolver called with %q", resolver.resolvedTenantID)
			}
		})
	}
}
