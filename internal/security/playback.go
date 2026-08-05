package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type PlaybackClaims struct {
	ResourceID string `json:"resource_id"`
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func GeneratePlaybackToken(
	resourceID, userID, tenantID, email, role, secret string,
	ttl time.Duration,
) (string, error) {
	now := time.Now().UTC()
	claims := PlaybackClaims{
		ResourceID: resourceID,
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
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
