package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type CourseEnrollmentRepository interface {
	Create(ctx context.Context, enrollment *domain.CourseEnrollment) error
	FindByID(ctx context.Context, id string) (*domain.CourseEnrollment, error)
	FindByCourse(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	FindByUser(ctx context.Context, userID string) ([]domain.CourseEnrollment, error)
	FindByCourseAndUser(
		ctx context.Context, courseID, userID string,
	) (*domain.CourseEnrollment, error)
	UpdateAssignment(ctx context.Context, id, assignmentType string) error
	Delete(ctx context.Context, id string) error
}
