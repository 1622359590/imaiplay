package service

import (
	"context"
	"errors"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type EnrollmentService struct {
	enrollments repository.CourseEnrollmentRepository
	courses     repository.CourseRepository
	users       repository.UserRepository
}

func NewEnrollmentService(
	enrollments repository.CourseEnrollmentRepository,
	courses repository.CourseRepository,
	users repository.UserRepository,
) *EnrollmentService {
	return &EnrollmentService{
		enrollments: enrollments, courses: courses, users: users,
	}
}

func (service *EnrollmentService) Enroll(
	ctx context.Context, courseID, userID string,
) (*domain.CourseEnrollment, error) {
	tenantID, err := tenantAdminID(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := service.courses.FindByID(ctx, courseID); err != nil {
		return nil, mapNotFound(err, "course not found")
	}
	user, err := service.users.FindByID(ctx, userID)
	if err != nil {
		return nil, mapNotFound(err, "user not found")
	}
	if user.Role != "learner" {
		return nil, errorsx.BadRequest("user must be a learner")
	}
	if user.Status != 1 {
		return nil, errorsx.BadRequest("learner is disabled")
	}
	if _, err := service.enrollments.FindByCourseAndUser(
		ctx, courseID, userID,
	); err == nil {
		return nil, errorsx.Conflict("learner already enrolled")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find enrollment failed")
	}
	enrollment := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		CourseID:  courseID, UserID: userID, Status: 1,
	}
	if err := service.enrollments.Create(ctx, enrollment); err != nil {
		return nil, mapCreateError(
			err, "learner already enrolled", "create enrollment failed",
		)
	}
	return enrollment, nil
}

func (service *EnrollmentService) ListByCourse(
	ctx context.Context, courseID string,
) ([]domain.CourseEnrollment, error) {
	if _, err := tenantAdminID(ctx); err != nil {
		return nil, err
	}
	if _, err := service.courses.FindByID(ctx, courseID); err != nil {
		return nil, mapNotFound(err, "course not found")
	}
	items, err := service.enrollments.FindByCourse(ctx, courseID)
	if err != nil {
		return nil, errorsx.Internal("list enrollments failed")
	}
	return items, nil
}

func (service *EnrollmentService) Remove(ctx context.Context, id string) error {
	if _, err := tenantAdminID(ctx); err != nil {
		return err
	}
	if _, err := service.enrollments.FindByID(ctx, id); err != nil {
		return mapNotFound(err, "enrollment not found")
	}
	return mapNotFound(service.enrollments.Delete(ctx, id), "enrollment not found")
}
