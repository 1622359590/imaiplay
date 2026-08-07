package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestUserImportHandlerReturnsPartialResultWithoutPasswords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewUserHandler(services.users)
	router := gin.New()
	router.Use(asRole("tenant_admin", tenant.ID))
	router.POST("/users/import", handler.Import)
	contents := "姓名,邮箱,手机号（可选）,角色（可选）,初始密码\n" +
		"张三,zhang@example.com,,学员,password1\n" +
		"弱密码,weak@example.com,,学员,short\n"

	response := requestUserImport(t, router, "users.csv", contents)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Data struct {
			Total     int `json:"total"`
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Total != 2 || body.Data.Succeeded != 1 || body.Data.Failed != 1 {
		t.Fatalf("body = %#v", body)
	}
	for _, password := range []string{"password1", "short"} {
		if strings.Contains(response.Body.String(), password) {
			t.Fatalf("response exposes password %q: %s", password, response.Body.String())
		}
	}
}

func TestUserImportHandlerRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewUserHandler(services.users)

	forbidden := gin.New()
	forbidden.Use(asRole("learner", tenant.ID))
	forbidden.POST("/users/import", handler.Import)
	if response := requestUserImport(t, forbidden, "users.csv", ""); response.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d body=%s", response.Code, response.Body.String())
	}

	router := gin.New()
	router.Use(asRole("tenant_admin", tenant.ID))
	router.POST("/users/import", handler.Import)
	request := httptest.NewRequest(http.MethodPost, "/users/import", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("missing file status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestUserImport(t, router, "users.txt", "not a spreadsheet"); response.Code != http.StatusBadRequest {
		t.Fatalf("file type status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestUserImportHandlerAppliesLimitToFileInsteadOfMultipartEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewUserHandler(services.users)
	router := gin.New()
	router.Use(asRole("tenant_admin", tenant.ID))
	router.POST("/users/import", handler.Import)
	header := "姓名,邮箱,手机号（可选）,角色（可选）,初始密码\n"
	lineSuffix := ",,,,\n"
	contents := header + strings.Repeat(" ", (10<<20)-len(header)-len(lineSuffix)) + lineSuffix

	if response := requestUserImport(t, router, "users.csv", contents); response.Code != http.StatusOK {
		t.Fatalf("10MB file status = %d body=%s", response.Code, response.Body.String())
	}
	if response := requestUserImport(t, router, "users.csv", contents+"x"); response.Code != http.StatusBadRequest {
		t.Fatalf("oversized file status = %d body=%s", response.Code, response.Body.String())
	}
}

func requestUserImport(t *testing.T, router http.Handler, filename, contents string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/users/import", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
