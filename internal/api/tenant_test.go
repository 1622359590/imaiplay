package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTenantHandlerCRUDAndRoleCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewTenantHandler(services.tenants)
	router := gin.New()
	router.Use(asRole("superadmin", ""))
	router.POST("/tenants", handler.Create)
	router.GET("/tenants", handler.List)
	router.GET("/tenants/:id", handler.Get)
	router.PUT("/tenants/:id", handler.Update)
	router.DELETE("/tenants/:id", handler.Delete)

	created := requestJSON(
		t, router, http.MethodPost, "/tenants", `{"code":"acme","name":"Acme"}`,
	)
	if created.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", created.Code, created.Body.String())
	}
	var createdBody struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createdBody.Data["code"] != "acme" {
		t.Fatalf("create data = %#v, want lowercase code field", createdBody.Data)
	}
	id, ok := createdBody.Data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("create id = %#v", createdBody.Data["id"])
	}
	list := requestJSON(t, router, http.MethodGet, "/tenants", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", list.Code, list.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if body.Code != 0 || body.Data.Total != 1 {
		t.Fatalf("list body = %#v", body)
	}
	if response := requestJSON(
		t, router, http.MethodGet, "/tenants/"+id, "",
	); response.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(
		t, router, http.MethodPut, "/tenants/"+id,
		`{"name":"Acme Academy","status":0}`,
	); response.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(
		t, router, http.MethodDelete, "/tenants/"+id, "",
	); response.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(
		t, router, http.MethodGet, "/tenants/"+id, "",
	); response.Code != http.StatusNotFound {
		t.Fatalf("deleted get status = %d body=%s", response.Code, response.Body.String())
	}

	forbidden := gin.New()
	forbidden.Use(asRole("tenant_admin", "tenant-1"))
	forbidden.GET("/tenants", handler.List)
	response := requestJSON(t, forbidden, http.MethodGet, "/tenants", "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d body=%s", response.Code, response.Body.String())
	}
}
