package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type courseEnrollmentGORMRepository struct{ database *gorm.DB }

func NewCourseEnrollmentRepository(database *gorm.DB) CourseEnrollmentRepository {
	return &courseEnrollmentGORMRepository{database: database}
}

func (repo *courseEnrollmentGORMRepository) Create(
	ctx context.Context, enrollment *domain.CourseEnrollment,
) error {
	return repo.database.WithContext(ctx).Create(enrollment).Error
}

func (repo *courseEnrollmentGORMRepository) FindByID(
	ctx context.Context, id string,
) (*domain.CourseEnrollment, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var enrollment domain.CourseEnrollment
	err = repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&enrollment).Error
	return &enrollment, err
}

func (repo *courseEnrollmentGORMRepository) FindByCourse(
	ctx context.Context, courseID string,
) ([]domain.CourseEnrollment, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var enrollments []domain.CourseEnrollment
	err = repo.database.WithContext(ctx).
		Where("course_id = ? AND tenant_id = ?", courseID, tenantID).
		Order("created_at ASC").Find(&enrollments).Error
	return enrollments, err
}

func (repo *courseEnrollmentGORMRepository) FindByUser(
	ctx context.Context, userID string,
) ([]domain.CourseEnrollment, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var enrollments []domain.CourseEnrollment
	err = repo.database.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("created_at DESC").Find(&enrollments).Error
	return enrollments, err
}

func (repo *courseEnrollmentGORMRepository) FindByCourseAndUser(
	ctx context.Context, courseID, userID string,
) (*domain.CourseEnrollment, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var enrollment domain.CourseEnrollment
	err = repo.database.WithContext(ctx).
		Where(
			"course_id = ? AND user_id = ? AND tenant_id = ?",
			courseID, userID, tenantID,
		).First(&enrollment).Error
	return &enrollment, err
}

func (repo *courseEnrollmentGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.CourseEnrollment{}))
}
