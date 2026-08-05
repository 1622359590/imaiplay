package repository

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
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
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" || enrollment.TenantID != tenantID || enrollment.Status != 1 {
		return gorm.ErrRecordNotFound
	}
	if role != "tenant_admin" && (role != "learner" || enrollment.UserID != userID) {
		return gorm.ErrRecordNotFound
	}
	assignmentType, err := repositoryAssignmentType(enrollment.AssignmentType, true)
	if err != nil {
		return err
	}
	return repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var learner domain.User
		if err := tx.Select("id").Where(
			"id = ? AND tenant_id = ? AND role = ? AND status = ?",
			enrollment.UserID, tenantID, "learner", 1,
		).First(&learner).Error; err != nil {
			return err
		}
		courseQuery := tx.Select("id").Where(
			"id = ? AND ((tenant_id = ? AND is_official = ?) OR (tenant_id = ? AND is_official = ? AND status = ? AND id IN (SELECT course_id FROM tenant_official_courses WHERE tenant_id = ? AND enabled = ?)))",
			enrollment.CourseID, tenantID, false, "", true, 1, tenantID, true,
		)
		if role == "learner" {
			courseQuery = courseQuery.Where("status = ?", 1)
		}
		var course domain.Course
		if err := courseQuery.First(&course).Error; err != nil {
			return err
		}
		enrollment.AssignmentType = assignmentType
		return tx.Create(enrollment).Error
	})
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

func (repo *courseEnrollmentGORMRepository) UpdateAssignment(
	ctx context.Context, id, assignmentType string,
) error {
	tenantID, err := enrollmentTenantAdminID(ctx)
	if err != nil {
		return err
	}
	assignmentType, err = repositoryAssignmentType(assignmentType, false)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).Model(&domain.CourseEnrollment{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("assignment_type", assignmentType))
}

func (repo *courseEnrollmentGORMRepository) Delete(
	ctx context.Context, id string,
) error {
	tenantID, err := enrollmentTenantAdminID(ctx)
	if err != nil {
		return err
	}
	return affected(repo.database.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.CourseEnrollment{}))
}

func enrollmentTenantAdminID(ctx context.Context) (string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return "", gorm.ErrRecordNotFound
	}
	return tenantID, nil
}

func repositoryAssignmentType(value string, defaultRequired bool) (string, error) {
	if value == "" && defaultRequired {
		return domain.AssignmentRequired, nil
	}
	if value != domain.AssignmentRequired && value != domain.AssignmentOptional {
		return "", ErrInvalidAssignmentType
	}
	return value, nil
}
