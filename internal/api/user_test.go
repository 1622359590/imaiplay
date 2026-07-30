package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserHandlerCRUDAndRoleCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewUserHandler(services.users)
	router := gin.New()
	router.Use(asRole("tenant_admin", tenant.ID))
	router.POST("/users", handler.Create)
	router.GET("/users", handler.List)
	router.GET("/users/:id", handler.Get)
	router.PUT("/users/:id", handler.Update)
	router.DELETE("/users/:id", handler.Delete)

	created := requestJSON(t, router, http.MethodPost, "/users",
		`{"email":"learner@example.com","password":"password123","name":"Learner","role":"learner"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", created.Code, created.Body.String())
	}
	if bytes := created.Body.Bytes(); string(bytes) == "" {
		t.Fatal("create response is empty")
	}
	var createdBody struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	id := createdBody.Data.ID
	if id == "" {
		t.Fatal("create id is empty")
	}
	list := requestJSON(t, router, http.MethodGet, "/users", "")
	var body struct {
		Code int `json:"code"`
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if list.Code != http.StatusOK || body.Code != 0 || body.Data.Total != 1 {
		t.Fatalf("list status=%d body=%#v", list.Code, body)
	}
	if response := requestJSON(
		t, router, http.MethodGet, "/users/"+id, "",
	); response.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(
		t, router, http.MethodPut, "/users/"+id,
		`{"name":"Updated Learner","status":0,"password":"newpass123"}`,
	); response.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", response.Code, response.Body.String())
	}

	crossTenant := gin.New()
	crossTenant.Use(asRole("tenant_admin", "other-tenant"))
	crossTenant.GET("/users/:id", handler.Get)
	if response := requestJSON(
		t, crossTenant, http.MethodGet, "/users/"+id, "",
	); response.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(
		t, router, http.MethodDelete, "/users/"+id, "",
	); response.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(
		t, router, http.MethodGet, "/users/"+id, "",
	); response.Code != http.StatusNotFound {
		t.Fatalf("deleted get status = %d body=%s", response.Code, response.Body.String())
	}

	forbidden := gin.New()
	forbidden.Use(asRole("learner", tenant.ID))
	forbidden.GET("/users", handler.List)
	response := requestJSON(t, forbidden, http.MethodGet, "/users", "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d body=%s", response.Code, response.Body.String())
	}
}
