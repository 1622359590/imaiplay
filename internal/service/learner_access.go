package service

import (
	"context"
	"errors"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type LearnerAccess struct {
	courses     repository.CourseRepository
	enrollments repository.CourseEnrollmentRepository
	materials   repository.CourseMaterialRepository
}

func NewLearnerAccess(
	courses repository.CourseRepository,
	enrollments repository.CourseEnrollmentRepository,
	materials repository.CourseMaterialRepository,
) *LearnerAccess {
	return &LearnerAccess{courses: courses, enrollments: enrollments, materials: materials}
}

func (access *LearnerAccess) AuthorizeCourse(
	ctx context.Context, courseID string,
) (*domain.Course, error) {
	userID, tenantID, ok := learnerAccessIdentity(ctx)
	if !ok || access == nil || access.courses == nil || access.enrollments == nil {
		return nil, learnerAccessNotFound()
	}
	course, err := access.courses.FindPublishedByID(ctx, tenantID, courseID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, learnerAccessNotFound()
	}
	if err != nil {
		return nil, errorsx.Internal("check learner course access failed")
	}
	if course == nil {
		return nil, learnerAccessNotFound()
	}
	if course.IsOfficial {
		return course, nil
	}
	enrollment, err := access.enrollments.FindByCourseAndUser(ctx, courseID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && (enrollment == nil || enrollment.Status != 1)) {
		return nil, learnerAccessNotFound()
	}
	if err != nil {
		return nil, errorsx.Internal("check learner enrollment failed")
	}
	return course, nil
}

func (access *LearnerAccess) AuthorizeMaterial(
	ctx context.Context, materialID string,
) (*domain.CourseMaterial, *domain.Course, error) {
	if _, _, ok := learnerAccessIdentity(ctx); !ok || access == nil || access.materials == nil {
		return nil, nil, learnerAccessNotFound()
	}
	material, err := access.materials.FindByID(ctx, materialID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, learnerAccessNotFound()
	}
	if err != nil {
		return nil, nil, errorsx.Internal("find course material failed")
	}
	if material == nil {
		return nil, nil, learnerAccessNotFound()
	}
	course, err := access.AuthorizeCourse(ctx, material.CourseID)
	if err != nil {
		return nil, nil, err
	}
	if material.CourseID != course.ID || material.TenantID != course.TenantID ||
		material.Resource.ID == "" || material.ResourceID != material.Resource.ID ||
		material.Resource.TenantID != course.TenantID {
		return nil, nil, learnerAccessNotFound()
	}
	return material, course, nil
}

func (access *LearnerAccess) AuthorizeLessonResource(
	ctx context.Context, resourceID string,
) (*domain.Course, error) {
	_, tenantID, ok := learnerAccessIdentity(ctx)
	if !ok || access == nil || access.courses == nil {
		return nil, learnerAccessNotFound()
	}
	candidates, err := access.courses.FindPublishedByLessonResource(ctx, tenantID, resourceID)
	if err != nil {
		return nil, errorsx.Internal("check lesson resource access failed")
	}
	for index := range candidates {
		course, authorizeErr := access.AuthorizeCourse(ctx, candidates[index].ID)
		if authorizeErr == nil {
			return course, nil
		}
		if errorCodeForLearnerAccess(authorizeErr) != 40400 {
			return nil, authorizeErr
		}
	}
	return nil, learnerAccessNotFound()
}

func learnerAccessIdentity(ctx context.Context) (string, string, bool) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	return userID, tenantID, ok && role == "learner" && userID != "" && tenantID != ""
}

func learnerAccessNotFound() error {
	return errorsx.NotFound("learner content not found")
}

func errorCodeForLearnerAccess(err error) int {
	var appErr *errorsx.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return 0
}
