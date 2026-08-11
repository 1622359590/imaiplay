package repository

import (
	"context"
	"errors"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type domainBindJobGORMRepository struct{ database *gorm.DB }

func NewDomainBindJobRepository(database *gorm.DB) DomainBindJobRepository {
	return &domainBindJobGORMRepository{database: database}
}

func (repo *domainBindJobGORMRepository) FindByTenant(ctx context.Context, tenantID string) (*domain.DomainBindJob, error) {
	var job domain.DomainBindJob
	if err := repo.database.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (repo *domainBindJobGORMRepository) FindByDomain(ctx context.Context, domainName string) (*domain.DomainBindJob, error) {
	var job domain.DomainBindJob
	if err := repo.database.WithContext(ctx).Where("domain = ?", domainName).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (repo *domainBindJobGORMRepository) Reserve(ctx context.Context, job *domain.DomainBindJob) error {
	existing, err := repo.FindByTenant(ctx, job.TenantID)
	if err == nil {
		if existing.Domain == job.Domain {
			return nil
		}
		return errors.New("tenant already has a domain binding job")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return repo.database.WithContext(ctx).Create(job).Error
}

func (repo *domainBindJobGORMRepository) UpdateStatus(ctx context.Context, tenantID, state, message string, step int, errorMessage string) error {
	result := repo.database.WithContext(ctx).Model(&domain.DomainBindJob{}).
		Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{"state": state, "message": message, "current_step": step, "error_message": errorMessage})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repo *domainBindJobGORMRepository) IncrementAttempt(ctx context.Context, tenantID string) error {
	result := repo.database.WithContext(ctx).Model(&domain.DomainBindJob{}).
		Where("tenant_id = ?", tenantID).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repo *domainBindJobGORMRepository) Delete(ctx context.Context, tenantID string) error {
	return repo.database.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&domain.DomainBindJob{}).Error
}
