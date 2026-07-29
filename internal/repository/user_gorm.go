package repository

import (
	"context"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/google/uuid"
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
	if user.Role == "superadmin" && user.TenantID == "" {
		if user.ID == "" {
			user.ID = uuid.NewString()
		}
		now := time.Now().UTC()
		if user.CreatedAt.IsZero() {
			user.CreatedAt = now
		}
		if user.UpdatedAt.IsZero() {
			user.UpdatedAt = now
		}
		return repository.database.WithContext(ctx).Model(&domain.User{}).Create(map[string]interface{}{
			"id": user.ID, "tenant_id": nil, "created_at": user.CreatedAt, "updated_at": user.UpdatedAt,
			"email": user.Email, "phone": user.Phone, "password": user.Password,
			"name": user.Name, "role": user.Role, "status": user.Status,
		}).Error
	}
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
	query := repository.database.WithContext(ctx).Where("id = ?", id)
	if tenantID == "" {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err = query.First(&user).Error
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
	query := repository.database.WithContext(ctx).Where("email = ?", email)
	if tenantID == "" {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repository *userGORMRepository) FindByPhoneAndTenant(ctx context.Context, phone, tenantID string) (*domain.User, error) {
	var user domain.User
	query := repository.database.WithContext(ctx).Where("phone = ?", phone)
	if tenantID == "" {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&user).Error
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
	query := repository.database.WithContext(ctx).Model(&domain.User{})
	if tenantID == "" {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", tenantID)
	}
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
