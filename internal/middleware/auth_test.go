package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/gin-gonic/gin"
)

func TestAuthAcceptsValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token, err := security.GenerateToken(
		"user-1", "tenant-1", "admin@example.com", "tenant_admin", "secret",
	)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	router := gin.New()
	router.Use(Auth("secret"))
	router.GET("/", func(c *gin.Context) {
		userID, tenantID, email, role, ok :=
			usercontext.UserFromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID, "tenant_id": tenantID,
			"email": email, "role": role, "ok": ok,
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["user_id"] != "user-1" || body["tenant_id"] != "tenant-1" ||
		body["ok"] != true {
		t.Fatalf("body = %#v", body)
	}
}

func TestAuthRejectsMissingOrInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		header string
	}{
		{"missing", ""},
		{"invalid", "Bearer invalid"},
		{"wrong scheme", "Basic abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Auth("secret"))
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Authorization", tt.header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", response.Code)
			}
			var body struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Code != 40100 {
				t.Fatalf("code = %d, want 40100", body.Code)
			}
		})
	}
}
