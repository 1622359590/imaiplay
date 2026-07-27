package middleware

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

type auditRecorderStub struct{ events []domain.AuditEvent }

func (stub *auditRecorderStub) Record(_ context.Context, event domain.AuditEvent) error {
	stub.events = append(stub.events, event)
	return nil
}

func TestAuditRecordsWriteAndRedactsSensitiveFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(RequestID(), Audit(recorder))
	router.POST("/api/v1/auth/login", func(c *gin.Context) { c.Status(401) })
	request := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"do-not-store","access_token":"do-not-store"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "request-22")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if len(recorder.events) != 1 {
		t.Fatalf("events = %d, want 1", len(recorder.events))
	}
	event := recorder.events[0]
	if event.Action != "auth.login_failed" || event.RequestID != "request-22" || event.UserEmail != "admin@example.com" {
		t.Fatalf("unexpected event: %+v", event)
	}
	if strings.Contains(event.Detail, "do-not-store") || strings.Contains(event.Detail, "password") {
		t.Fatalf("sensitive detail was stored: %s", event.Detail)
	}
}

func TestAuditSkipsReads(t *testing.T) {
	recorder := &auditRecorderStub{}
	router := gin.New()
	router.Use(Audit(recorder))
	router.GET("/backend/v1/courses", func(c *gin.Context) { c.Status(200) })
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/backend/v1/courses", nil))
	if len(recorder.events) != 0 {
		t.Fatalf("read request created %d audit events", len(recorder.events))
	}
}
