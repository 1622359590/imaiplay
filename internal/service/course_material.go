package service

import (
	"context"
	"errors"
	"io"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type CourseMaterialInput struct {
	ResourceID  string `json:"resource_id"`
	DisplayName string `json:"display_name"`
	SortOrder   int    `json:"sort_order"`
}

type CourseMaterialService struct {
	courses   repository.CourseRepository
	materials repository.CourseMaterialRepository
	resources repository.ResourceRepository
	opener    interface {
		Open(context.Context, string) (io.ReadCloser, string, string, error)
	}
}

func NewCourseMaterialService(
	courses repository.CourseRepository,
	materials repository.CourseMaterialRepository,
	resources repository.ResourceRepository,
	opener interface {
		Open(context.Context, string) (io.ReadCloser, string, string, error)
	},
) *CourseMaterialService {
	return &CourseMaterialService{courses: courses, materials: materials, resources: resources, opener: opener}
}

func (service *CourseMaterialService) OpenForLearner(
	ctx context.Context, materialID string,
) (io.ReadCloser, string, string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || tenantID == "" {
		return nil, "", "", errorsx.Forbidden("permission denied")
	}
	material, err := service.materials.FindByID(ctx, materialID)
	if err != nil {
		return nil, "", "", errorsx.NotFound("course material not found")
	}
	if _, err := service.courses.FindPublishedByID(ctx, tenantID, material.CourseID); err != nil {
		return nil, "", "", errorsx.NotFound("course material not found")
	}
	if service.opener == nil {
		return nil, "", "", errorsx.Internal("resource service unavailable")
	}
	body, contentType, _, err := service.opener.Open(ctx, material.ResourceID)
	if err != nil {
		return nil, "", "", errorsx.NotFound("course material not found")
	}
	return body, contentType, material.DisplayName, nil
}

func (service *CourseMaterialService) Add(
	ctx context.Context, courseID string, input CourseMaterialInput,
) (*domain.CourseMaterial, error) {
	course, err := service.manageableCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	name, err := validateMaterialInput(input)
	if err != nil {
		return nil, err
	}
	resource, err := service.findManagedResource(ctx, course, input.ResourceID)
	if err != nil {
		return nil, err
	}
	if resource.ResourceType != "attachment" {
		return nil, errorsx.BadRequest("course material must be an attachment")
	}
	existing, err := service.materials.FindByCourse(ctx, courseID)
	if err != nil {
		return nil, errorsx.Internal("list course materials failed")
	}
	for _, item := range existing {
		if item.ResourceID == resource.ID {
			return nil, errorsx.Conflict("course material already exists")
		}
	}
	userID, _, _, _, _ := usercontext.UserFromContext(ctx)
	material := &domain.CourseMaterial{
		BaseModel: domain.BaseModel{TenantID: course.TenantID},
		CourseID:  course.ID, ResourceID: resource.ID, DisplayName: name,
		SortOrder: input.SortOrder, CreatedBy: userID, Resource: *resource,
	}
	if err := service.materials.Create(ctx, material); err != nil {
		return nil, errorsx.Internal("create course material failed")
	}
	return material, nil
}

func (service *CourseMaterialService) Update(
	ctx context.Context, courseID, materialID string, input CourseMaterialInput,
) (*domain.CourseMaterial, error) {
	course, err := service.manageableCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	name, err := validateMaterialInput(input)
	if err != nil {
		return nil, err
	}
	material, err := service.materials.FindByID(ctx, materialID)
	if err != nil || material.CourseID != courseID {
		return nil, errorsx.NotFound("course material not found")
	}
	resource, err := service.findManagedResource(ctx, course, input.ResourceID)
	if err != nil {
		return nil, err
	}
	if resource.ResourceType != "attachment" {
		return nil, errorsx.BadRequest("course material must be an attachment")
	}
	if resource.ID != material.ResourceID {
		existing, err := service.materials.FindByCourse(ctx, courseID)
		if err != nil {
			return nil, errorsx.Internal("list course materials failed")
		}
		for _, item := range existing {
			if item.ID != material.ID && item.ResourceID == resource.ID {
				return nil, errorsx.Conflict("course material already exists")
			}
		}
	}
	material.DisplayName, material.SortOrder, material.ResourceID = name, input.SortOrder, resource.ID
	material.Resource = *resource
	if err := service.materials.Update(ctx, material); err != nil {
		return nil, mapNotFound(err, "course material not found")
	}
	return material, nil
}

func (service *CourseMaterialService) Remove(ctx context.Context, courseID, materialID string) error {
	if _, err := service.manageableCourse(ctx, courseID); err != nil {
		return err
	}
	material, err := service.materials.FindByID(ctx, materialID)
	if err != nil || material.CourseID != courseID {
		return errorsx.NotFound("course material not found")
	}
	return mapNotFound(service.materials.Delete(ctx, materialID), "course material not found")
}

func (service *CourseMaterialService) ListForManager(ctx context.Context, courseID string) ([]domain.CourseMaterial, error) {
	if _, err := service.manageableCourse(ctx, courseID); err != nil {
		return nil, err
	}
	items, err := service.materials.FindByCourse(ctx, courseID)
	if err != nil {
		return nil, errorsx.Internal("list course materials failed")
	}
	return items, nil
}

func (service *CourseMaterialService) manageableCourse(ctx context.Context, courseID string) (*domain.Course, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (role != "tenant_admin" && role != "superadmin") {
		return nil, errorsx.Forbidden("permission denied")
	}
	course, err := service.courses.FindByID(ctx, courseID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.NotFound("course not found")
	}
	if err != nil {
		return nil, errorsx.Internal("find course failed")
	}
	if role == "superadmin" {
		if !course.IsOfficial || course.TenantID != "" {
			return nil, errorsx.Forbidden("permission denied")
		}
		return course, nil
	}
	if course.IsOfficial || course.TenantID != tenantID {
		return nil, errorsx.Forbidden("permission denied")
	}
	return course, nil
}

func (service *CourseMaterialService) findManagedResource(
	ctx context.Context, course *domain.Course, resourceID string,
) (*domain.Resource, error) {
	var resource *domain.Resource
	var err error
	if course.IsOfficial {
		resource, err = service.resources.FindPlatformByID(ctx, resourceID)
	} else {
		resource, err = service.resources.FindByID(ctx, resourceID)
	}
	if err != nil {
		return nil, errorsx.NotFound("resource not found")
	}
	if resource.TenantID != course.TenantID {
		return nil, errorsx.NotFound("resource not found")
	}
	return resource, nil
}

func validateMaterialInput(input CourseMaterialInput) (string, error) {
	name := strings.TrimSpace(input.DisplayName)
	if name == "" || len([]rune(name)) > 255 || strings.TrimSpace(input.ResourceID) == "" {
		return "", errorsx.BadRequest("invalid course material")
	}
	return name, nil
}
