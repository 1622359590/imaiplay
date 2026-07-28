package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

type AuditService struct{ logs repository.AuditLogRepository }

func NewAuditService(logs repository.AuditLogRepository) *AuditService {
	return &AuditService{logs: logs}
}

func (service *AuditService) Record(ctx context.Context, event domain.AuditEvent) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	event.Detail = sanitizeDetailString(event.Detail)
	if event.TenantID == "" {
		_, tenantID, email, role, ok := usercontext.UserFromContext(ctx)
		if ok {
			event.TenantID, event.UserEmail, event.UserRole = tenantID, first(event.UserEmail, email), first(event.UserRole, role)
		}
	}
	return service.logs.Create(ctx, &domain.AuditLog{
		BaseModel: domain.BaseModel{TenantID: event.TenantID, CreatedAt: event.CreatedAt, UpdatedAt: event.CreatedAt},
		UserID:    event.UserID, UserEmail: event.UserEmail, UserRole: event.UserRole,
		Action: event.Action, ResourceType: event.ResourceType, ResourceID: event.ResourceID,
		Detail: event.Detail, IP: event.IP, RequestID: event.RequestID,
	})
}

func sanitizeDetailString(value string) string {
	var payload map[string]interface{}
	if json.Unmarshal([]byte(value), &payload) != nil {
		return "{}"
	}
	return AuditDetail(payload)
}

func first(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func (service *AuditService) List(ctx context.Context, filter repository.AuditLogFilter) ([]domain.AuditLog, int64, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (role != "tenant_admin" && role != "superadmin") {
		return nil, 0, fmt.Errorf("permission denied")
	}
	if role == "tenant_admin" {
		filter.TenantID = tenantID
	}
	return service.logs.List(ctx, filter)
}

func AuditDetail(values map[string]interface{}) string {
	clean := sanitize(values)
	data, err := json.Marshal(clean)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func sanitize(value interface{}) interface{} {
	switch current := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(current))
		for key, item := range current {
			lower := key
			if containsSensitive(lower) {
				continue
			}
			result[key] = sanitize(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(current))
		for i, item := range current {
			result[i] = sanitize(item)
		}
		return result
	default:
		return value
	}
}

func containsSensitive(key string) bool {
	key = stringLower(key)
	for _, part := range []string{"password", "secret", "token", "authorization", "accesskey", "refresh", "captcha", "verificationcode"} {
		if contains(key, part) {
			return true
		}
	}
	return false
}

func stringLower(value string) string {
	const delta = 'a' - 'A'
	result := []byte(value)
	for i, char := range result {
		if char >= 'A' && char <= 'Z' {
			result[i] = char + delta
		}
	}
	return string(result)
}
func contains(value, part string) bool {
	if len(part) == 0 {
		return true
	}
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
