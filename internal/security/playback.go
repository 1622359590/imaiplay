package security

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type PlaybackClaims struct {
	ResourceID string `json:"resource_id"`
	CourseID   string `json:"course_id"`
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func GeneratePlaybackToken(
	resourceID, courseID, userID, tenantID, email, role, secret string,
	ttl time.Duration,
) (string, error) {
	if strings.TrimSpace(resourceID) == "" || strings.TrimSpace(courseID) == "" ||
		strings.TrimSpace(userID) == "" || strings.TrimSpace(tenantID) == "" ||
		strings.TrimSpace(role) == "" {
		return "", errors.New("invalid playback identity")
	}
	now := time.Now().UTC()
	claims := PlaybackClaims{
		ResourceID: resourceID,
		CourseID:   courseID,
		UserID:     userID, TenantID: tenantID, Email: email, Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "imaiplay-playback",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(secret))
}

func ValidatePlaybackToken(tokenString, secret string) (*PlaybackClaims, error) {
	claims := &PlaybackClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *jwt.Token) (interface{}, error) { return []byte(secret), nil },
		jwt.WithIssuer("imaiplay-playback"),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid playback token")
	}
	if strings.TrimSpace(claims.ResourceID) == "" || strings.TrimSpace(claims.CourseID) == "" ||
		strings.TrimSpace(claims.UserID) == "" || strings.TrimSpace(claims.TenantID) == "" ||
		strings.TrimSpace(claims.Role) == "" || claims.ExpiresAt == nil {
		return nil, errors.New("invalid playback claims")
	}
	return claims, nil
}
