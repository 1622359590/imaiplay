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
