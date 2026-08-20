package repository

import (
	"context"
	"strings"
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
	return repository.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if user.Role != "learner" {
			return nil
		}
		return tx.Create(&domain.LearnerEngagementState{
			BaseModel: domain.BaseModel{TenantID: user.TenantID},
			UserID:    user.ID,
		}).Error
	})
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

func (repository *userGORMRepository) FindByIDAcrossTenants(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	var user domain.User
	if err := repository.database.WithContext(ctx).
		Where("id = ? AND tenant_id IS NOT NULL", id).
		First(&user).Error; err != nil {
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

func (repository *userGORMRepository) FindByCredentialAcrossTenants(
	ctx context.Context,
	identifier string,
) ([]domain.User, error) {
	identifier = strings.TrimSpace(identifier)
	query := repository.database.WithContext(ctx).
		Where("tenant_id IS NOT NULL")
	if strings.Contains(identifier, "@") {
		query = query.Where("LOWER(email) = ?", strings.ToLower(identifier))
	} else {
		query = query.Where("phone = ?", identifier)
	}
	var users []domain.User
	if err := query.Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
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

func (repository *userGORMRepository) FindAll(
	ctx context.Context,
	offset, limit int,
) ([]domain.User, int64, error) {
	query := repository.database.WithContext(ctx).Model(&domain.User{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []domain.User
	if err := query.Preload("Tenant").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	for index := range users {
		users[index].TenantName = users[index].Tenant.Name
		users[index].TenantCode = users[index].Tenant.Code
	}
	return users, total, nil
}

func (repository *userGORMRepository) UpdatePassword(
	ctx context.Context,
	id, password string,
) error {
	result := repository.database.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND role = ?", id, "tenant_admin").
		Updates(map[string]interface{}{"password": password})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

func (repository *userGORMRepository) Deactivate(
	ctx context.Context,
	id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	result := repository.database.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{"status": 0})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func tenantIDFromContext(ctx context.Context) (string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (tenantID == "" && role != "superadmin") {
		return "", gorm.ErrRecordNotFound
	}
	return tenantID, nil
}
