package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type CourseService interface {
	Create(context.Context, string, string, string) (*domain.Course, error)
	List(context.Context, int, int) ([]domain.Course, int64, error)
	Get(context.Context, string) (*domain.Course, error)
	Update(context.Context, string, string, string, string, int) (*domain.Course, error)
	Delete(context.Context, string) error
	GetDetail(context.Context, string) (*service.CourseDetail, error)
	ListPublished(context.Context, int, int) ([]domain.Course, int64, error)
	GetPublishedDetail(context.Context, string) (*service.CourseDetail, error)
}

type CourseHandler struct {
	service CourseService
}

func NewCourseHandler(courseService CourseService) *CourseHandler {
	return &CourseHandler{service: courseService}
}

func (handler *CourseHandler) Create(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	var request struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	course, err := handler.service.Create(
		c.Request.Context(), request.Title, request.Description, request.CoverImage,
	)
	respond(c, course, err)
}

func (handler *CourseHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.List(c.Request.Context(), offset, limit)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}

func (handler *CourseHandler) Get(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	course, err := handler.service.Get(c.Request.Context(), c.Param("id"))
	respond(c, course, err)
}

func (handler *CourseHandler) Update(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	var request struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
		Status      *int   `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	course, err := handler.service.Update(
		c.Request.Context(), c.Param("id"), request.Title,
		request.Description, request.CoverImage, *request.Status,
	)
	respond(c, course, err)
}

func (handler *CourseHandler) Delete(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}

func (handler *CourseHandler) Detail(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	detail, err := handler.service.GetDetail(c.Request.Context(), c.Param("id"))
	respond(c, detail, err)
}

func (handler *CourseHandler) PublishedList(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.ListPublished(c.Request.Context(), offset, limit)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}

func (handler *CourseHandler) PublishedDetail(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	detail, err := handler.service.GetPublishedDetail(
		c.Request.Context(), c.Param("id"),
	)
	respond(c, detail, err)
}
