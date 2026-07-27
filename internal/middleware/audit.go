package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

type AuditRecorder interface {
	Record(context.Context, domain.AuditEvent) error
}

func Audit(recorder AuditRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload map[string]interface{}
		if strings.Contains(c.GetHeader("Content-Type"), "application/json") && c.Request.ContentLength <= 64*1024 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewReader(body))
				_ = json.Unmarshal(body, &payload)
			}
		}
		c.Next()
		if recorder == nil || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			return
		}
		action, resourceType := auditAction(c.FullPath(), c.Request.Method, c.Writer.Status())
		if action == "" {
			return
		}
		userID, tenantID, email, role, _ := usercontext.UserFromContext(c.Request.Context())
		if email == "" && payload != nil {
			if value, ok := payload["email"].(string); ok {
				email = value
			}
		}
		event := domain.AuditEvent{TenantID: tenantID, UserID: userID, UserEmail: email, UserRole: role, Action: action, ResourceType: resourceType, ResourceID: c.Param("id"), Detail: auditDetail(payload), IP: c.ClientIP(), RequestID: RequestIDFromContext(c.Request.Context())}
		if err := recorder.Record(c.Request.Context(), event); err != nil {
			slog.Default().Error("write audit log", "error", err, "action", action)
		}
	}
}

func auditDetail(payload map[string]interface{}) string {
	clean := make(map[string]interface{}, len(payload))
	for key, value := range payload {
		if !sensitiveAuditKey(key) {
			clean[key] = sanitizeAuditValue(value)
		}
	}
	data, err := json.Marshal(clean)
	if err != nil {
		return "{}"
	}
	return string(data)
}
func sanitizeAuditValue(value interface{}) interface{} {
	switch current := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(current))
		for key, item := range current {
			if !sensitiveAuditKey(key) {
				result[key] = sanitizeAuditValue(item)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(current))
		for i, item := range current {
			result[i] = sanitizeAuditValue(item)
		}
		return result
	default:
		return value
	}
}
func sensitiveAuditKey(key string) bool {
	key = strings.ToLower(key)
	for _, part := range []string{"password", "secret", "token", "authorization", "accesskey", "refresh", "captcha", "verificationcode"} {
		if strings.Contains(key, part) {
			return true
		}
	}
	return false
}

func auditAction(path, method string, status int) (string, string) {
	if method == http.MethodOptions {
		return "", ""
	}
	if path == "/api/v1/auth/login" && method == http.MethodPost {
		if status >= 200 && status < 300 {
			return "auth.login_success", "auth"
		}
		return "auth.login_failed", "auth"
	}
	if path == "/api/v1/auth/logout" && method == http.MethodPost {
		return "auth.logout", "auth"
	}
	if path == "/api/v1/auth/register" && method == http.MethodPost {
		return "auth.register", "user"
	}
	if path == "/api/v1/auth/forgot-password" && method == http.MethodPost {
		return "auth.password_reset_request", "auth"
	}
	if path == "/api/v1/auth/reset-password" && method == http.MethodPost {
		return "auth.password_reset", "auth"
	}
	if path == "/api/v1/tenants/register" && method == http.MethodPost {
		return "tenant.register", "tenant"
	}
	if path == "/api/v1/bootstrap/superadmin" && method == http.MethodPost {
		return "auth.superadmin_bootstrap", "user"
	}
	if strings.HasPrefix(path, "/backend/v1/") {
		parts := strings.Split(strings.TrimPrefix(path, "/backend/v1/"), "/")
		resource := map[string]string{"tenants": "tenant", "users": "user", "courses": "course", "chapters": "course_chapter", "lessons": "course_lesson", "resources": "resource", "resource-categories": "resource_category", "sms-config": "config"}[parts[0]]
		if len(parts) >= 2 && parts[0] == "admin" && parts[1] == "tenants" {
			resource = "tenant"
		}
		if resource == "" {
			return "", ""
		}
		verb := map[string]string{http.MethodPost: "create", http.MethodPut: "update", http.MethodDelete: "delete"}[method]
		if verb == "" {
			return "", ""
		}
		if parts[0] == "sms-config" {
			verb = "update"
		}
		return resource + "." + verb, resource
	}
	return "", ""
}
