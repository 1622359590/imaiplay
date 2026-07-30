package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminHostAllowsConfiguredManagementHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AdminHost("play.imai.work"))
	router.GET("/backend/v1/users", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "http://play.imai.work/backend/v1/users", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
	}
}

func TestAdminHostRejectsTenantCustomDomain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AdminHost("play.imai.work"))
	router.GET("/backend/v1/users", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "https://academy.example.com/backend/v1/users", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNotFound)
	}
}
