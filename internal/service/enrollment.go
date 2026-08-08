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
	ctx context.Context, courseID, userID, assignmentType string,
) (*domain.CourseEnrollment, error) {
	if _, err := validateAssignmentType(assignmentType, true); err != nil {
		return nil, err
	}
	tenantID, err := tenantAdminID(ctx)
	if err != nil {
		return nil, err
	}
	course, err := service.courses.FindByID(ctx, courseID)
	if err != nil {
		return nil, mapNotFound(err, "course not found")
	}
	assignmentType = course.CourseType
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
		AssignmentType: assignmentType,
	}
	if err := service.enrollments.Create(ctx, enrollment); err != nil {
		return nil, mapCreateError(
			err, "learner already enrolled", "create enrollment failed",
		)
	}
	return enrollment, nil
}

func (service *EnrollmentService) UpdateAssignment(
	ctx context.Context, id, assignmentType string,
) (*domain.CourseEnrollment, error) {
	if _, err := tenantAdminID(ctx); err != nil {
		return nil, err
	}
	assignmentType, err := validateAssignmentType(assignmentType, false)
	if err != nil {
		return nil, err
	}
	enrollment, err := service.enrollments.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "enrollment not found")
	}
	if err := service.enrollments.UpdateAssignment(ctx, id, assignmentType); err != nil {
		return nil, mapNotFound(err, "enrollment not found")
	}
	enrollment.AssignmentType = assignmentType
	return enrollment, nil
}

func validateAssignmentType(value string, defaultRequired bool) (string, error) {
	if value == "" && defaultRequired {
		return domain.AssignmentRequired, nil
	}
	if value != domain.AssignmentRequired && value != domain.AssignmentOptional {
		return "", errorsx.BadRequest("invalid assignment type")
	}
	return value, nil
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
