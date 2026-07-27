package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

func TestAuthHandlerRegisterAndLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	createTenant(t, tenantRepo)
	handler := NewAuthHandler(services.auth)
	router := gin.New()
	router.Use(middleware.Tenant())
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

	register := requestJSON(t, router, http.MethodPost, "/register",
		`{"email":"admin@example.com","password":"password123","name":"Admin","role":"tenant_admin"}`)
	if register.Code != http.StatusOK {
		t.Fatalf("register status = %d body=%s", register.Code, register.Body.String())
	}
	login := requestJSON(t, router, http.MethodPost, "/login",
		`{"email":"admin@example.com","password":"password123"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", login.Code, login.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if body.Code != 0 || body.Data.Token == "" {
		t.Fatalf("login body = %#v", body)
	}
}

func TestAuthHandlerRejectsSuperadminRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	createTenant(t, tenantRepo)
	handler := NewAuthHandler(services.auth)
	router := gin.New()
	router.Use(middleware.Tenant())
	router.POST("/register", handler.Register)

	response := requestJSON(t, router, http.MethodPost, "/register",
		`{"email":"root@example.com","password":"password123","name":"Root","role":"superadmin"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != 40000 || body.Message != "superadmin 不可通過公開註冊創建" {
		t.Fatalf("body = %#v", body)
	}
}

func TestAuthHandlerBootstrapSuperadminIsOneTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewAuthHandler(services.auth)
	router := gin.New()
	router.POST("/api/v1/bootstrap/superadmin", handler.BootstrapSuperadmin)
	first := requestJSON(t, router, http.MethodPost, "/api/v1/bootstrap/superadmin", `{"email":"root@example.com","name":"Root","password":"password123"}`)
	if first.Code != http.StatusOK {
		t.Fatalf("first bootstrap status=%d body=%s", first.Code, first.Body.String())
	}
	second := requestJSON(t, router, http.MethodPost, "/api/v1/bootstrap/superadmin", `{"email":"root2@example.com","name":"Root 2","password":"password123"}`)
	if second.Code != http.StatusConflict {
		t.Fatalf("second bootstrap status=%d body=%s", second.Code, second.Body.String())
	}
}

func requestJSON(
	t *testing.T,
	router http.Handler,
	method, path, body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Host = "acme.imaiplay.local"
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
