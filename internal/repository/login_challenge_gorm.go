package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type loginChallengeGORMRepository struct {
	database *gorm.DB
}

func NewLoginChallengeRepository(
	database *gorm.DB,
) LoginChallengeRepository {
	return &loginChallengeGORMRepository{database: database}
}

func (repository *loginChallengeGORMRepository) Create(
	ctx context.Context,
	challenge *domain.LoginChallenge,
) error {
	return repository.database.WithContext(ctx).Create(challenge).Error
}

func (repository *loginChallengeGORMRepository) Consume(
	ctx context.Context,
	tokenHash string,
	now time.Time,
) (*domain.LoginChallenge, error) {
	var challenge domain.LoginChallenge
	err := repository.database.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			result := transaction.Model(&domain.LoginChallenge{}).
				Where(
					"token_hash = ? AND consumed_at IS NULL AND expires_at > ?",
					tokenHash,
					now,
				).
				Update("consumed_at", now)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			return transaction.
				Where("token_hash = ?", tokenHash).
				First(&challenge).Error
		},
	)
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}
