package service

import (
	"context"
	"errors"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type CourseCategoryService struct {
	categories repository.CourseCategoryRepository
}

func NewCourseCategoryService(
	categories repository.CourseCategoryRepository,
) *CourseCategoryService {
	return &CourseCategoryService{categories: categories}
}

func NormalizeCourseCategoryName(value string) (string, string, error) {
	display := strings.Join(strings.Fields(norm.NFKC.String(value)), " ")
	if display == "" || len([]rune(display)) > 64 {
		return "", "", errorsx.BadRequest("invalid course category name")
	}
	return display, cases.Fold().String(display), nil
}

func (service *CourseCategoryService) Create(
	ctx context.Context, name string, sortOrder, status int,
) (*domain.CourseCategory, error) {
	tenantID, err := tenantCourseCategoryScope(ctx, true)
	if err != nil {
		return nil, err
	}
	return service.create(ctx, tenantID, name, sortOrder, status)
}

func (service *CourseCategoryService) List(
	ctx context.Context,
) ([]domain.CourseCategory, error) {
	tenantID, err := tenantCourseCategoryScope(ctx, false)
	if err != nil {
		return nil, err
	}
	return service.list(ctx, tenantID)
}

func (service *CourseCategoryService) Update(
	ctx context.Context, id, name string, sortOrder, status int,
) (*domain.CourseCategory, error) {
	tenantID, err := tenantCourseCategoryScope(ctx, true)
	if err != nil {
		return nil, err
	}
	return service.update(ctx, tenantID, id, name, sortOrder, status)
}

func (service *CourseCategoryService) Delete(ctx context.Context, id string) error {
	tenantID, err := tenantCourseCategoryScope(ctx, true)
	if err != nil {
		return err
	}
	return service.delete(ctx, tenantID, id)
}

func (service *CourseCategoryService) CreatePlatform(
	ctx context.Context, name string, sortOrder, status int,
) (*domain.CourseCategory, error) {
	if err := platformCourseCategoryScope(ctx); err != nil {
		return nil, err
	}
	return service.create(ctx, "", name, sortOrder, status)
}

func (service *CourseCategoryService) ListPlatform(
	ctx context.Context,
) ([]domain.CourseCategory, error) {
	if err := platformCourseCategoryScope(ctx); err != nil {
		return nil, err
	}
	return service.list(ctx, "")
}

func (service *CourseCategoryService) UpdatePlatform(
	ctx context.Context, id, name string, sortOrder, status int,
) (*domain.CourseCategory, error) {
	if err := platformCourseCategoryScope(ctx); err != nil {
		return nil, err
	}
	return service.update(ctx, "", id, name, sortOrder, status)
}

func (service *CourseCategoryService) DeletePlatform(ctx context.Context, id string) error {
	if err := platformCourseCategoryScope(ctx); err != nil {
		return err
	}
	return service.delete(ctx, "", id)
}

func (service *CourseCategoryService) create(
	ctx context.Context, tenantID, name string, sortOrder, status int,
) (*domain.CourseCategory, error) {
	display, normalized, err := NormalizeCourseCategoryName(name)
	if err != nil {
		return nil, err
	}
	if err := validateCourseCategoryStatus(status); err != nil {
		return nil, err
	}
	category := &domain.CourseCategory{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Name:      display, NormalizedName: normalized, SortOrder: sortOrder, Status: status,
	}
	if err := service.categories.Create(ctx, category); err != nil {
		if errors.Is(err, repository.ErrCourseCategoryNameConflict) {
			return nil, errorsx.Conflict("course category already exists")
		}
		return nil, errorsx.Internal("create course category failed")
	}
	return category, nil
}

func (service *CourseCategoryService) list(
	ctx context.Context, tenantID string,
) ([]domain.CourseCategory, error) {
	items, err := service.categories.FindByTenant(ctx, tenantID)
	if err != nil {
		return nil, errorsx.Internal("list course categories failed")
	}
	return items, nil
}

func (service *CourseCategoryService) update(
	ctx context.Context, tenantID, id, name string, sortOrder, status int,
) (*domain.CourseCategory, error) {
	display, normalized, err := NormalizeCourseCategoryName(name)
	if err != nil {
		return nil, err
	}
	if err := validateCourseCategoryStatus(status); err != nil {
		return nil, err
	}
	category, err := service.categories.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, mapNotFound(err, "course category not found")
	}
	category.Name, category.NormalizedName = display, normalized
	category.SortOrder, category.Status = sortOrder, status
	if err := service.categories.Update(ctx, tenantID, category); err != nil {
		if errors.Is(err, repository.ErrCourseCategoryNameConflict) {
			return nil, errorsx.Conflict("course category already exists")
		}
		return nil, mapNotFound(err, "course category not found")
	}
	return category, nil
}

func (service *CourseCategoryService) delete(
	ctx context.Context, tenantID, id string,
) error {
	err := service.categories.Delete(ctx, tenantID, id)
	if errors.Is(err, repository.ErrCourseCategoryInUse) {
		return errorsx.Conflict("course category is referenced")
	}
	return mapNotFound(err, "course category not found")
}

func tenantCourseCategoryScope(ctx context.Context, write bool) (string, error) {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	allowed := role == "tenant_admin" || (!write && role == "instructor")
	if !ok || tenantID == "" || !allowed {
		return "", errorsx.Forbidden("permission denied")
	}
	return tenantID, nil
}

func platformCourseCategoryScope(ctx context.Context) error {
	_, _, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "superadmin" {
		return errorsx.Forbidden("permission denied")
	}
	return nil
}

func validateCourseCategoryStatus(status int) error {
	if status != 0 && status != 1 {
		return errorsx.BadRequest("invalid course category status")
	}
	return nil
}
