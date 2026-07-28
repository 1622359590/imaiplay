package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type passwordResetGORMRepository struct{ database *gorm.DB }

func NewPasswordResetRepository(database *gorm.DB) PasswordResetRepository {
	return &passwordResetGORMRepository{database: database}
}

func (repo *passwordResetGORMRepository) Create(ctx context.Context, reset *domain.PasswordReset) error {
	return repo.database.WithContext(ctx).Create(reset).Error
}

func (repo *passwordResetGORMRepository) FindLatest(ctx context.Context, tenantID, phone string) (*domain.PasswordReset, error) {
	return repo.FindLatestForPurpose(ctx, tenantID, phone, "password_reset")
}

func (repo *passwordResetGORMRepository) FindLatestForPurpose(ctx context.Context, tenantID, phone, purpose string) (*domain.PasswordReset, error) {
	var reset domain.PasswordReset
	err := repo.database.WithContext(ctx).Where("tenant_id = ? AND phone = ? AND purpose = ?", tenantID, phone, purpose).Order("created_at DESC").First(&reset).Error
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

func (repo *passwordResetGORMRepository) IncrementAttempts(ctx context.Context, id string) error {
	return repo.database.WithContext(ctx).Model(&domain.PasswordReset{}).Where("id = ?", id).UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error
}

func (repo *passwordResetGORMRepository) MarkUsed(ctx context.Context, id string) error {
	return repo.database.WithContext(ctx).Model(&domain.PasswordReset{}).Where("id = ?", id).Updates(map[string]interface{}{"used": true, "updated_at": time.Now().UTC()}).Error
}
