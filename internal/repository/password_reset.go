package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type PasswordResetRepository interface {
	Create(context.Context, *domain.PasswordReset) error
	FindLatest(context.Context, string, string) (*domain.PasswordReset, error)
	FindLatestForPurpose(context.Context, string, string, string) (*domain.PasswordReset, error)
	IncrementAttempts(context.Context, string) error
	MarkUsed(context.Context, string) error
}
