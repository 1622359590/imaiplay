package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type resourceGORMRepository struct{ database *gorm.DB }

func NewResourceRepository(database *gorm.DB) ResourceRepository {
	return &resourceGORMRepository{database: database}
}

func (repo *resourceGORMRepository) Create(
	ctx context.Context, resource *domain.Resource,
) error {
	return repo.database.WithContext(ctx).Create(resource).Error
}

func (repo *resourceGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.Resource, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var resource domain.Resource
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&resource).Error
	return &resource, err
}

func (repo *resourceGORMRepository) FindByTenant(
	ctx context.Context, tenantID string, offset, limit int,
) ([]domain.Resource, int64, error) {
	query := repo.database.WithContext(ctx).Model(&domain.Resource{}).
		Where("tenant_id = ?", tenantID)
	return repo.find(ctx, query, offset, limit)
}

func (repo *resourceGORMRepository) FindPlatformByID(
	ctx context.Context, id string,
) (*domain.Resource, error) {
	var resource domain.Resource
	err := repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, "").
		First(&resource).Error
	return &resource, err
}

func (repo *resourceGORMRepository) FindPlatform(
	ctx context.Context, offset, limit int,
) ([]domain.Resource, int64, error) {
	query := repo.database.WithContext(ctx).Model(&domain.Resource{}).
		Where("tenant_id = ?", "")
	return repo.find(ctx, query, offset, limit)
}

func (repo *resourceGORMRepository) DeletePlatform(
	ctx context.Context, id string,
) error {
	return affected(repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, "").
		Delete(&domain.Resource{}))
}

func (repo *resourceGORMRepository) IsPlatformReferenced(
	ctx context.Context, id string, coverURLs []string,
) (bool, error) {
	var count int64
	if err := repo.database.WithContext(ctx).Model(&domain.CourseLesson{}).
		Where("tenant_id = ? AND resource_id = ?", "", id).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	if len(coverURLs) == 0 {
		return false, nil
	}
	if err := repo.database.WithContext(ctx).Model(&domain.Course{}).
		Where(
			"tenant_id = ? AND is_official = ? AND cover_image IN ?",
			"", true, coverURLs,
		).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *resourceGORMRepository) CanAccessPlatformResource(
	ctx context.Context,
	resourceID, tenantID, userID, role string,
) (bool, error) {
	if role == "superadmin" {
		var count int64
		err := repo.database.WithContext(ctx).Model(&domain.Resource{}).
			Where("id = ? AND tenant_id = ?", resourceID, "").
			Count(&count).Error
		return count > 0, err
	}
	if tenantID == "" ||
		(role != "tenant_admin" && role != "instructor" && role != "learner") {
		return false, nil
	}
	query := repo.database.WithContext(ctx).
		Table("course_lessons AS lessons").
		Joins(
			"JOIN course_chapters AS chapters ON chapters.id = lessons.chapter_id AND chapters.tenant_id = ?",
			"",
		).
		Joins(
			"JOIN courses ON courses.id = chapters.course_id AND courses.tenant_id = ? AND courses.is_official = ? AND courses.status = ?",
			"", true, 1,
		).
		Joins(
			"JOIN tenant_official_courses AS enabled_courses ON enabled_courses.course_id = courses.id AND enabled_courses.tenant_id = ? AND enabled_courses.enabled = ?",
			tenantID, true,
		).
		Joins(
			"JOIN resources ON resources.id = lessons.resource_id AND resources.tenant_id = ?",
			"",
		).
		Where("lessons.resource_id = ?", resourceID)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *resourceGORMRepository) find(
	ctx context.Context, query *gorm.DB, offset, limit int,
) ([]domain.Resource, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.Resource
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (repo *resourceGORMRepository) Update(
	ctx context.Context, resource *domain.Resource,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).Model(&domain.Resource{}).
		Where("id = ? AND tenant_id = ?", resource.ID, tenantID).
		Updates(map[string]interface{}{
			"category_id": resource.CategoryID,
			"name":        resource.Name, "resource_type": resource.ResourceType,
			"url": resource.URL, "size_bytes": resource.SizeBytes,
		}))
}

func (repo *resourceGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.Resource{}))
}

func (repo *resourceGORMRepository) TotalSizeByTenant(ctx context.Context, tenantID string) (int64, error) {
	var total int64
	err := repo.database.WithContext(ctx).Model(&domain.Resource{}).Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(size_bytes), 0)").Scan(&total).Error
	return total, err
}
