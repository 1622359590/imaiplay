package service

import (
	"context"
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type ResourceCategoryService struct {
	categories repository.ResourceCategoryRepository
}

func NewResourceCategoryService(
	categories repository.ResourceCategoryRepository,
) *ResourceCategoryService {
	return &ResourceCategoryService{categories: categories}
}

func (service *ResourceCategoryService) Create(
	ctx context.Context, name string, parentID *string,
) (*domain.ResourceCategory, error) {
	_, tenantID, err := resourceManager(ctx)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errorsx.BadRequest("category name is required")
	}
	if err := service.validateParent(ctx, "", parentID); err != nil {
		return nil, err
	}
	category := &domain.ResourceCategory{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Name:      name, ParentID: parentID,
	}
	if err := service.categories.Create(ctx, category); err != nil {
		return nil, errorsx.Internal("create resource category failed")
	}
	return category, nil
}

func (service *ResourceCategoryService) List(
	ctx context.Context,
) ([]domain.ResourceCategory, error) {
	_, tenantID, err := resourceManager(ctx)
	if err != nil {
		return nil, err
	}
	items, err := service.categories.FindByTenant(ctx, tenantID)
	if err != nil {
		return nil, errorsx.Internal("list resource categories failed")
	}
	return items, nil
}

func (service *ResourceCategoryService) Update(
	ctx context.Context, id, name string, parentID *string,
) (*domain.ResourceCategory, error) {
	if _, _, err := resourceManager(ctx); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errorsx.BadRequest("category name is required")
	}
	category, err := service.categories.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "resource category not found")
	}
	if err := service.validateParent(ctx, id, parentID); err != nil {
		return nil, err
	}
	category.Name, category.ParentID = name, parentID
	if err := service.categories.Update(ctx, category); err != nil {
		return nil, mapNotFound(err, "resource category not found")
	}
	return category, nil
}

func (service *ResourceCategoryService) Delete(
	ctx context.Context, id string,
) error {
	if _, _, err := resourceManager(ctx); err != nil {
		return err
	}
	if _, err := service.categories.FindByID(ctx, id); err != nil {
		return mapNotFound(err, "resource category not found")
	}
	items, err := service.List(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.ParentID != nil && *item.ParentID == id {
			return errorsx.Conflict("resource category has children")
		}
	}
	return mapNotFound(
		service.categories.Delete(ctx, id), "resource category not found",
	)
}

func (service *ResourceCategoryService) validateParent(
	ctx context.Context, id string, parentID *string,
) error {
	if parentID == nil {
		return nil
	}
	if *parentID == "" || *parentID == id {
		return errorsx.BadRequest("invalid parent category")
	}
	if _, err := service.categories.FindByID(ctx, *parentID); err != nil {
		return mapNotFound(err, "parent category not found")
	}
	if id == "" {
		return nil
	}
	items, err := service.List(ctx)
	if err != nil {
		return err
	}
	parents := make(map[string]*string, len(items))
	for _, item := range items {
		parents[item.ID] = item.ParentID
	}
	visited := map[string]bool{}
	for current := parentID; current != nil; current = parents[*current] {
		if *current == id || visited[*current] {
			return errorsx.BadRequest("invalid parent category")
		}
		visited[*current] = true
	}
	return nil
}
