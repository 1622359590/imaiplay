package security

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken(
		"user-1", "tenant-1", "admin@example.com", "tenant_admin", "secret",
	)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ValidateToken(token, "secret")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != "user-1" || claims.TenantID != "tenant-1" ||
		claims.Email != "admin@example.com" || claims.Role != "tenant_admin" {
		t.Fatalf("claims = %#v", claims)
	}
	if claims.Issuer != "imaiplay" {
		t.Fatalf("issuer = %q, want imaiplay", claims.Issuer)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) < 23*time.Hour {
		t.Fatalf("expires_at = %v, want approximately 24 hours", claims.ExpiresAt)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken("u", "t", "a@example.com", "learner", "secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if _, err := ValidateToken(token, "wrong-secret"); err == nil {
		t.Fatal("ValidateToken() accepted wrong secret")
	}
}
