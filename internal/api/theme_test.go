package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

type themeServiceStub struct{ brandName string }

func (stub *themeServiceStub) Get(context.Context) (*domain.Tenant, error) {
	return &domain.Tenant{PrimaryColor: "#4F46E5", BrandName: "Acme Academy"}, nil
}

func (stub *themeServiceStub) Update(
	_ context.Context, primary, _, _, _, brand string,
) (*domain.Tenant, error) {
	stub.brandName = brand
	return &domain.Tenant{PrimaryColor: primary, BrandName: brand}, nil
}

func TestThemeHandlerReturnsBrandName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewThemeHandler(&themeServiceStub{})
	router := gin.New()
	router.GET("/theme", handler.Get)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/theme", nil))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"brand_name":"Acme Academy"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestThemeHandlerUpdatesBrandName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &themeServiceStub{}
	handler := NewThemeHandler(service)
	router := gin.New()
	router.PUT("/theme", func(c *gin.Context) {
		ctx := usercontext.WithUser(c.Request.Context(), "admin", "tenant-1", "admin@example.com", "tenant_admin")
		c.Request = c.Request.WithContext(ctx)
		handler.Update(c)
	})

	request := httptest.NewRequest(
		http.MethodPut, "/theme",
		strings.NewReader(`{"primary_color":"#3582E1","brand_name":"Sales School"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK || service.brandName != "Sales School" {
		t.Fatalf("status=%d brand=%q body=%s", response.Code, service.brandName, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"brand_name":"Sales School"`) {
		t.Fatalf("response missing brand_name: %s", response.Body.String())
	}
}
