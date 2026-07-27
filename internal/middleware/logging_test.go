package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareReusesOrGeneratesID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/health", nil))
	if first.Header().Get("X-Request-ID") == "" {
		t.Fatal("request id was not generated")
	}
	secondRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	secondRequest.Header.Set("X-Request-ID", "client-request-id")
	second := httptest.NewRecorder()
	router.ServeHTTP(second, secondRequest)
	if second.Header().Get("X-Request-ID") != "client-request-id" {
		t.Fatalf("request id = %q", second.Header().Get("X-Request-ID"))
	}
}

func TestLoggingWritesStructuredRequestFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	router := gin.New()
	router.Use(RequestID(), Logging(logger))
	router.GET("/health", func(c *gin.Context) {
		c.Request = c.Request.WithContext(usercontext.WithUser(c.Request.Context(), "user-1", "tenant-1", "", "learner"))
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("X-Request-ID", "test-request")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	var record map[string]interface{}
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log: %v; output=%s", err, output.String())
	}
	for _, key := range []string{"request_id", "method", "path", "status", "duration", "client_ip", "tenant_id", "user_id"} {
		if _, ok := record[key]; !ok {
			t.Fatalf("missing log field %q: %s", key, output.String())
		}
	}
	if record["request_id"] != "test-request" || record["status"] != float64(http.StatusNoContent) {
		t.Fatalf("log fields=%v", record)
	}
}
