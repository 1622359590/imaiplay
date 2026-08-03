package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByIDAcrossTenants(ctx context.Context, id string) (*domain.User, error)
	FindByEmailAndTenant(
		ctx context.Context,
		email, tenantID string,
	) (*domain.User, error)
	FindByPhoneAndTenant(ctx context.Context, phone, tenantID string) (*domain.User, error)
	FindByCredentialAcrossTenants(ctx context.Context, identifier string) ([]domain.User, error)
	FindByTenant(
		ctx context.Context,
		tenantID string,
		offset, limit int,
	) ([]domain.User, int64, error)
	FindAll(ctx context.Context, offset, limit int) ([]domain.User, int64, error)
	UpdatePassword(ctx context.Context, id, password string) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
