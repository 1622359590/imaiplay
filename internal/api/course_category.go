package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type CourseCategoryService interface {
	Create(ctx context.Context, name string, sortOrder, status int) (*domain.CourseCategory, error)
	List(ctx context.Context) ([]domain.CourseCategory, error)
	Update(ctx context.Context, id, name string, sortOrder, status int) (*domain.CourseCategory, error)
	Delete(ctx context.Context, id string) error
	CreatePlatform(ctx context.Context, name string, sortOrder, status int) (*domain.CourseCategory, error)
	ListPlatform(ctx context.Context) ([]domain.CourseCategory, error)
	UpdatePlatform(ctx context.Context, id, name string, sortOrder, status int) (*domain.CourseCategory, error)
	DeletePlatform(ctx context.Context, id string) error
}

type CourseCategoryHandler struct{ service CourseCategoryService }

func NewCourseCategoryHandler(service CourseCategoryService) *CourseCategoryHandler {
	return &CourseCategoryHandler{service: service}
}

type courseCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Status    *int   `json:"status"`
}

func (handler *CourseCategoryHandler) Create(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	handler.create(c, false)
}

func (handler *CourseCategoryHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	items, err := handler.service.List(c.Request.Context())
	respond(c, items, err)
}

func (handler *CourseCategoryHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	handler.update(c, false)
}

func (handler *CourseCategoryHandler) Delete(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	handler.delete(c, false)
}

func (handler *CourseCategoryHandler) CreatePlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.create(c, true)
}

func (handler *CourseCategoryHandler) ListPlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	items, err := handler.service.ListPlatform(c.Request.Context())
	respond(c, items, err)
}

func (handler *CourseCategoryHandler) UpdatePlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.update(c, true)
}

func (handler *CourseCategoryHandler) DeletePlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.delete(c, true)
}

func (handler *CourseCategoryHandler) create(c *gin.Context, platform bool) {
	var request courseCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	status := categoryRequestStatus(request.Status)
	var category *domain.CourseCategory
	var err error
	if platform {
		category, err = handler.service.CreatePlatform(c.Request.Context(), request.Name, request.SortOrder, status)
	} else {
		category, err = handler.service.Create(c.Request.Context(), request.Name, request.SortOrder, status)
	}
	respond(c, category, err)
}

func (handler *CourseCategoryHandler) update(c *gin.Context, platform bool) {
	var request courseCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	status := categoryRequestStatus(request.Status)
	var category *domain.CourseCategory
	var err error
	if platform {
		category, err = handler.service.UpdatePlatform(c.Request.Context(), c.Param("id"), request.Name, request.SortOrder, status)
	} else {
		category, err = handler.service.Update(c.Request.Context(), c.Param("id"), request.Name, request.SortOrder, status)
	}
	respond(c, category, err)
}

func (handler *CourseCategoryHandler) delete(c *gin.Context, platform bool) {
	var err error
	if platform {
		err = handler.service.DeletePlatform(c.Request.Context(), c.Param("id"))
	} else {
		err = handler.service.Delete(c.Request.Context(), c.Param("id"))
	}
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}

func categoryRequestStatus(status *int) int {
	if status == nil {
		return 1
	}
	return *status
}
