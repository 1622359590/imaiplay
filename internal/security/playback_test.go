package security

import (
	"testing"
	"time"
)

func TestPlaybackTokenRoundTripAndExpiry(t *testing.T) {
	token, err := GeneratePlaybackToken(
		"resource-1", "learner-1", "tenant-1", "learner@example.com",
		"learner", "secret", time.Minute,
	)
	if err != nil {
		t.Fatalf("GeneratePlaybackToken() error = %v", err)
	}
	claims, err := ValidatePlaybackToken(token, "secret")
	if err != nil {
		t.Fatalf("ValidatePlaybackToken() error = %v", err)
	}
	if claims.ResourceID != "resource-1" || claims.UserID != "learner-1" ||
		claims.TenantID != "tenant-1" || claims.Role != "learner" {
		t.Fatalf("claims = %#v", claims)
	}

	expired, err := GeneratePlaybackToken(
		"resource-1", "learner-1", "tenant-1", "learner@example.com",
		"learner", "secret", -time.Minute,
	)
	if err != nil {
		t.Fatalf("GeneratePlaybackToken(expired) error = %v", err)
	}
	if _, err := ValidatePlaybackToken(expired, "secret"); err == nil {
		t.Fatal("ValidatePlaybackToken(expired) error = nil")
	}
}
