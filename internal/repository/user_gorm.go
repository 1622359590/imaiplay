package repository

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type userGORMRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) UserRepository {
	return &userGORMRepository{database: database}
}

func (repository *userGORMRepository) Create(
	ctx context.Context,
	user *domain.User,
) error {
	return repository.database.WithContext(ctx).Create(user).Error
}

func (repository *userGORMRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var user domain.User
	err = repository.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repository *userGORMRepository) FindByEmailAndTenant(
	ctx context.Context,
	email, tenantID string,
) (*domain.User, error) {
	var user domain.User
	err := repository.database.WithContext(ctx).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repository *userGORMRepository) FindByPhoneAndTenant(ctx context.Context, phone, tenantID string) (*domain.User, error) {
	var user domain.User
	err := repository.database.WithContext(ctx).Where("phone = ? AND tenant_id = ?", phone, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repository *userGORMRepository) FindByTenant(
	ctx context.Context,
	tenantID string,
	offset, limit int,
) ([]domain.User, int64, error) {
	query := repository.database.WithContext(ctx).
		Model(&domain.User{}).
		Where("tenant_id = ?", tenantID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []domain.User
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (repository *userGORMRepository) Update(
	ctx context.Context,
	user *domain.User,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repository.database.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", user.ID, tenantID).
		Updates(map[string]interface{}{
			"email": user.Email, "password": user.Password,
			"phone": user.Phone,
			"name":  user.Name, "role": user.Role, "status": user.Status,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repository *userGORMRepository) Delete(
	ctx context.Context,
	id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repository.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func tenantIDFromContext(ctx context.Context) (string, error) {
	_, tenantID, _, _, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" {
		return "", gorm.ErrRecordNotFound
	}
	return tenantID, nil
}
