package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type themeServiceStub struct {
	update service.ThemeUpdate
	theme  *domain.Tenant
}

func (stub *themeServiceStub) Get(context.Context) (*domain.Tenant, error) {
	return stub.theme, nil
}

func (stub *themeServiceStub) Update(_ context.Context, update service.ThemeUpdate) (*domain.Tenant, error) {
	stub.update = update
	return stub.theme, nil
}

func TestThemeHandlerReturnsBrandName(t *testing.T) {
	stub := &themeServiceStub{theme: &domain.Tenant{PrimaryColor: "#4F46E5", BrandName: "Acme Academy"}}
	router := gin.New()
	router.GET("/theme", NewThemeHandler(stub).Get)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/theme", nil))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"brand_name":"Acme Academy"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestThemeHandlerUpdatesBrandNameAndIndependentSelectionColors(t *testing.T) {
	stub := &themeServiceStub{theme: &domain.Tenant{
		PrimaryColor:            "#3582E1",
		SelectedBackgroundColor: "#FFF1F0",
		SelectedTextColor:       "#C5221F",
		SelectedIconColor:       "#8C1D18",
		BrandName:               "Sales School",
	}}
	router := gin.New()
	router.Use(asRole("tenant_admin", "tenant-one"))
	router.PUT("/backend/v1/theme", NewThemeHandler(stub).Update)
	request := httptest.NewRequest(http.MethodPut, "/backend/v1/theme", strings.NewReader(`{
		"primary_color":"#3582E1",
		"selected_background_color":"#FFF1F0",
		"selected_text_color":"#C5221F",
		"selected_icon_color":"#8C1D18",
		"brand_name":"Sales School"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if stub.update.SelectedBackgroundColor != "#FFF1F0" ||
		stub.update.SelectedTextColor != "#C5221F" ||
		stub.update.SelectedIconColor != "#8C1D18" ||
		stub.update.BrandName != "Sales School" {
		t.Fatalf("update=%#v", stub.update)
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data["selected_background_color"] != "#FFF1F0" ||
		body.Data["selected_text_color"] != "#C5221F" ||
		body.Data["selected_icon_color"] != "#8C1D18" ||
		body.Data["brand_name"] != "Sales School" {
		t.Fatalf("body=%#v", body)
	}
}

func TestThemeHandlerPreservesCrossClientThemeContract(t *testing.T) {
	tests := []struct {
		name  string
		theme *domain.Tenant
	}{
		{
			name: "populated brand fields",
			theme: &domain.Tenant{
				PrimaryColor:            "#3582E1",
				SelectedBackgroundColor: "#FFF1F0",
				SelectedTextColor:       "#C5221F",
				SelectedIconColor:       "#8C1D18",
				LogoURL:                 "https://cdn.example.test/logo.svg",
				WelcomeText:             "欢迎来到 Acme 学院",
				BrowserTitle:            "Acme Learning",
				BrandName:               "Acme Academy",
			},
		},
		{
			name: "empty optional brand fields",
			theme: &domain.Tenant{
				PrimaryColor:            "#123456",
				SelectedBackgroundColor: "#654321",
				SelectedTextColor:       "#FFFFFF",
				SelectedIconColor:       "#000000",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &themeServiceStub{theme: test.theme}
			router := gin.New()
			handler := NewThemeHandler(stub)
			router.GET("/theme", handler.Get)
			router.Use(asRole("tenant_admin", "tenant-one"))
			router.PUT("/theme", handler.Update)

			getResponse := httptest.NewRecorder()
			router.ServeHTTP(getResponse, httptest.NewRequest(http.MethodGet, "/theme", nil))
			assertThemeContractResponse(t, getResponse, test.theme)

			payload, err := json.Marshal(themeContractRequest(test.theme))
			if err != nil {
				t.Fatal(err)
			}
			putResponse := httptest.NewRecorder()
			putRequest := httptest.NewRequest(http.MethodPut, "/theme", strings.NewReader(string(payload)))
			putRequest.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(putResponse, putRequest)
			assertThemeContractResponse(t, putResponse, test.theme)
			if got := themeContractRequestFromUpdate(stub.update); !themeContractMapsEqual(got, themeContractRequest(test.theme)) {
				t.Fatalf("update = %#v, want %#v", got, themeContractRequest(test.theme))
			}
		})
	}
}

func themeContractRequest(theme *domain.Tenant) map[string]string {
	return map[string]string{
		"primary_color":             theme.PrimaryColor,
		"selected_background_color": theme.SelectedBackgroundColor,
		"selected_text_color":       theme.SelectedTextColor,
		"selected_icon_color":       theme.SelectedIconColor,
		"logo_url":                  theme.LogoURL,
		"welcome_text":              theme.WelcomeText,
		"browser_title":             theme.BrowserTitle,
		"brand_name":                theme.BrandName,
	}
}

func themeContractRequestFromUpdate(update service.ThemeUpdate) map[string]string {
	return map[string]string{
		"primary_color":             update.PrimaryColor,
		"selected_background_color": update.SelectedBackgroundColor,
		"selected_text_color":       update.SelectedTextColor,
		"selected_icon_color":       update.SelectedIconColor,
		"logo_url":                  update.LogoURL,
		"welcome_text":              update.WelcomeText,
		"browser_title":             update.BrowserTitle,
		"brand_name":                update.BrandName,
	}
}

func assertThemeContractResponse(t *testing.T, response *httptest.ResponseRecorder, want *domain.Tenant) {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !themeContractMapsEqual(body.Data, themeContractRequest(want)) {
		t.Fatalf("body data = %#v, want %#v", body.Data, themeContractRequest(want))
	}
}

func themeContractMapsEqual(got, want map[string]string) bool {
	if len(got) != len(want) {
		return false
	}
	for field, wantValue := range want {
		if got[field] != wantValue {
			return false
		}
	}
	return true
}
