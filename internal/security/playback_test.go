package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPlaybackTokenRoundTripAndExpiry(t *testing.T) {
	token, err := GeneratePlaybackToken(
		"resource-1", "course-1", "learner-1", "tenant-1", "learner@example.com",
		"learner", "secret", time.Minute,
	)
	if err != nil {
		t.Fatalf("GeneratePlaybackToken() error = %v", err)
	}
	claims, err := ValidatePlaybackToken(token, "secret")
	if err != nil {
		t.Fatalf("ValidatePlaybackToken() error = %v", err)
	}
	if claims.ResourceID != "resource-1" || claims.CourseID != "course-1" || claims.UserID != "learner-1" ||
		claims.TenantID != "tenant-1" || claims.Role != "learner" {
		t.Fatalf("claims = %#v", claims)
	}

	expired, err := GeneratePlaybackToken(
		"resource-1", "course-1", "learner-1", "tenant-1", "learner@example.com",
		"learner", "secret", -time.Minute,
	)
	if err != nil {
		t.Fatalf("GeneratePlaybackToken(expired) error = %v", err)
	}
	if _, err := ValidatePlaybackToken(expired, "secret"); err == nil {
		t.Fatal("ValidatePlaybackToken(expired) error = nil")
	}
}

func TestPlaybackTokenRejectsMissingRequiredIdentityClaims(t *testing.T) {
	now := time.Now().UTC()
	legacy := PlaybackClaims{
		ResourceID: "resource-1", UserID: "learner-1", TenantID: "tenant-1",
		Role: "learner",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "imaiplay-playback", ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, legacy).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign legacy token: %v", err)
	}
	if _, err := ValidatePlaybackToken(token, "secret"); err == nil {
		t.Fatal("ValidatePlaybackToken(legacy missing course_id) error = nil")
	}

	for _, test := range []struct {
		name       string
		resourceID string
		courseID   string
		userID     string
		tenantID   string
		role       string
	}{
		{"resource", "", "course-1", "learner-1", "tenant-1", "learner"},
		{"course", "resource-1", "", "learner-1", "tenant-1", "learner"},
		{"user", "resource-1", "course-1", "", "tenant-1", "learner"},
		{"tenant", "resource-1", "course-1", "learner-1", "", "learner"},
		{"role", "resource-1", "course-1", "learner-1", "tenant-1", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := GeneratePlaybackToken(
				test.resourceID, test.courseID, test.userID, test.tenantID,
				"learner@example.com", test.role, "secret", time.Minute,
			); err == nil {
				t.Fatal("GeneratePlaybackToken() error = nil")
			}
		})
	}
}
