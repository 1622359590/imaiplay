package api

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResourceCategoryHandlerCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewResourceCategoryHandler(services.resourceCategories)
	router := gin.New()
	router.Use(asUser("tenant_admin", tenant.ID, "admin-1"))
	router.POST("/resource-categories", handler.Create)
	router.GET("/resource-categories", handler.List)
	router.PUT("/resource-categories/:id", handler.Update)
	router.DELETE("/resource-categories/:id", handler.Delete)

	created := requestJSON(t, router, http.MethodPost, "/resource-categories",
		`{"name":"Videos"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("Create status=%d body=%s", created.Code, created.Body.String())
	}
	id := responseID(t, created.Body.Bytes())
	if response := requestJSON(
		t, router, http.MethodGet, "/resource-categories", "",
	); response.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(t, router, http.MethodPut,
		"/resource-categories/"+id, `{"name":"Updated","parent_id":null}`); response.Code != http.StatusOK {
		t.Fatalf("Update status=%d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(t, router, http.MethodDelete,
		"/resource-categories/"+id, ""); response.Code != http.StatusOK {
		t.Fatalf("Delete status=%d body=%s", response.Code, response.Body.String())
	}
}
