package repository

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type courseMaterialGORMRepository struct{ database *gorm.DB }

func NewCourseMaterialRepository(database *gorm.DB) CourseMaterialRepository {
	return &courseMaterialGORMRepository{database: database}
}

func (repo *courseMaterialGORMRepository) Create(ctx context.Context, material *domain.CourseMaterial) error {
	return repo.database.WithContext(ctx).Create(material).Error
}

func (repo *courseMaterialGORMRepository) FindByID(ctx context.Context, id string) (*domain.CourseMaterial, error) {
	var material domain.CourseMaterial
	err := repo.readScope(ctx).Preload("Resource").Where("course_materials.id = ?", id).First(&material).Error
	return &material, err
}

func (repo *courseMaterialGORMRepository) FindByCourse(ctx context.Context, courseID string) ([]domain.CourseMaterial, error) {
	items := make([]domain.CourseMaterial, 0)
	err := repo.readScope(ctx).Preload("Resource").
		Where("course_materials.course_id = ?", courseID).
		Order("course_materials.sort_order ASC, course_materials.created_at ASC").
		Find(&items).Error
	return items, err
}

func (repo *courseMaterialGORMRepository) Update(ctx context.Context, material *domain.CourseMaterial) error {
	return affected(repo.writeScope(ctx).
		Where("id = ?", material.ID).
		Updates(map[string]interface{}{
			"display_name": material.DisplayName,
			"sort_order":   material.SortOrder,
			"resource_id":  material.ResourceID,
		}))
}

func (repo *courseMaterialGORMRepository) Delete(ctx context.Context, id string) error {
	return affected(repo.writeScope(ctx).Where("id = ?", id).Delete(&domain.CourseMaterial{}))
}

func (repo *courseMaterialGORMRepository) DeleteByCourse(ctx context.Context, courseID string) error {
	return repo.writeScope(ctx).Where("course_id = ?", courseID).Delete(&domain.CourseMaterial{}).Error
}

func (repo *courseMaterialGORMRepository) readScope(ctx context.Context) *gorm.DB {
	query := repo.database.WithContext(ctx).Model(&domain.CourseMaterial{})
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if ok && role == "superadmin" {
		return query
	}
	if !ok || tenantID == "" {
		return query.Where("1 = 0")
	}
	return query.Where(
		"course_materials.tenant_id = ? OR (course_materials.tenant_id = ? AND course_materials.course_id IN (SELECT courses.id FROM courses JOIN tenant_official_courses ON tenant_official_courses.course_id = courses.id WHERE courses.is_official = ? AND courses.tenant_id = ? AND courses.status = ? AND tenant_official_courses.tenant_id = ? AND tenant_official_courses.enabled = ?))",
		tenantID, "", true, "", 1, tenantID, true,
	)
}

func (repo *courseMaterialGORMRepository) writeScope(ctx context.Context) *gorm.DB {
	query := repo.database.WithContext(ctx).Model(&domain.CourseMaterial{})
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if ok && role == "superadmin" {
		return query.Where("tenant_id = ?", "")
	}
	if !ok || role != "tenant_admin" || tenantID == "" {
		return query.Where("1 = 0")
	}
	return query.Where("tenant_id = ?", tenantID)
}
