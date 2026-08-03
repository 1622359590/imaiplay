package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
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

func TestAuthHandlerPlatformLoginSelectsOrganization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	for _, code := range []string{"tenant-one", "tenant-two"} {
		tenant := &domain.Tenant{Code: code, Name: code, Status: 1}
		if err := tenantRepo.Create(context.Background(), tenant); err != nil {
			t.Fatalf("create tenant %q: %v", code, err)
		}
		ctx := tenantcontext.WithTenant(context.Background(), code, tenantcontext.SourceHeaderCode)
		if _, err := services.auth.Register(
			ctx, "shared-admin@example.com", "password123", code, "tenant_admin",
		); err != nil {
			t.Fatalf("register tenant admin %q: %v", code, err)
		}
	}

	handler := NewAuthHandler(services.auth)
	router := gin.New()
	router.Use(middleware.TenantWithRepositoryForAdminHost(tenantRepo, "play.imai.work"))
	router.POST("/login", handler.Login)
	router.POST("/select-tenant", handler.SelectTenant)
	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(`{"identifier":"shared-admin@example.com","password":"password123"}`),
	)
	request.Host = "play.imai.work"
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			RequiresTenantSelection bool   `json:"requires_tenant_selection"`
			SelectionToken          string `json:"selection_token"`
			Organizations           []struct {
				Code string `json:"code"`
				Role string `json:"role"`
			} `json:"organizations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != 0 || !body.Data.RequiresTenantSelection ||
		body.Data.SelectionToken == "" || len(body.Data.Organizations) != 2 {
		t.Fatalf("body = %#v", body)
	}

	selectRequest := httptest.NewRequest(
		http.MethodPost,
		"/select-tenant",
		bytes.NewBufferString(`{"selection_token":"`+
			body.Data.SelectionToken+`","tenant_code":"tenant-one"}`),
	)
	selectRequest.Host = "play.imai.work"
	selectRequest.Header.Set("Content-Type", "application/json")
	selected := httptest.NewRecorder()
	router.ServeHTTP(selected, selectRequest)
	if selected.Code != http.StatusOK {
		t.Fatalf(
			"select status = %d body=%s",
			selected.Code,
			selected.Body.String(),
		)
	}
	var selectedBody struct {
		Data struct {
			Token  string `json:"token"`
			Tenant struct {
				Code string `json:"code"`
			} `json:"tenant"`
		} `json:"data"`
	}
	if err := json.Unmarshal(selected.Body.Bytes(), &selectedBody); err != nil {
		t.Fatalf("decode selected response: %v", err)
	}
	if selectedBody.Data.Token == "" ||
		selectedBody.Data.Tenant.Code != "tenant-one" {
		t.Fatalf("selected body = %#v", selectedBody)
	}
}

func TestAuthHandlerPlatformLoginDoesNotRevealOrganizationsForWrongPassword(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	for _, code := range []string{"tenant-one", "tenant-two"} {
		tenant := &domain.Tenant{Code: code, Name: code, Status: 1}
		if err := tenantRepo.Create(context.Background(), tenant); err != nil {
			t.Fatal(err)
		}
		ctx := tenantcontext.WithTenant(
			context.Background(),
			code,
			tenantcontext.SourceHeaderCode,
		)
		if _, err := services.auth.Register(
			ctx,
			"shared@example.com",
			"password123",
			code,
			"learner",
		); err != nil {
			t.Fatal(err)
		}
	}
	handler := NewAuthHandler(services.auth)
	router := gin.New()
	router.Use(
		middleware.TenantWithRepositoryForAdminHost(
			tenantRepo,
			"play.imai.work",
		),
	)
	router.POST("/login", handler.Login)
	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(
			`{"identifier":"shared@example.com","password":"wrong"}`,
		),
	)
	request.Host = "play.imai.work"
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("organizations")) ||
		bytes.Contains(response.Body.Bytes(), []byte("tenant-one")) ||
		bytes.Contains(response.Body.Bytes(), []byte("tenant-two")) {
		t.Fatalf("wrong-password response leaked organizations: %s", response.Body.String())
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
