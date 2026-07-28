package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type refreshTokenGORMRepository struct{ database *gorm.DB }

func NewRefreshTokenRepository(database *gorm.DB) RefreshTokenRepository {
	return &refreshTokenGORMRepository{database: database}
}

func (repo *refreshTokenGORMRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return repo.database.WithContext(ctx).Create(token).Error
}

func (repo *refreshTokenGORMRepository) FindValidByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := repo.database.WithContext(ctx).
		Where("token_hash = ? AND revoked = ?", hash, false).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (repo *refreshTokenGORMRepository) Revoke(ctx context.Context, hash string) error {
	result := repo.database.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("token_hash = ? AND revoked = ?", hash, false).Update("revoked", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repo *refreshTokenGORMRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	return repo.database.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}
