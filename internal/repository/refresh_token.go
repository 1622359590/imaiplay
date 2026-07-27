package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type RefreshTokenRepository interface {
	Create(context.Context, *domain.RefreshToken) error
	FindValidByHash(context.Context, string) (*domain.RefreshToken, error)
	Revoke(context.Context, string) error
	RevokeAllForUser(context.Context, string) error
}
