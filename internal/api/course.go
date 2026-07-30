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
	CreateOfficial(context.Context, string, string, string) (*domain.Course, error)
	List(context.Context, int, int) ([]domain.Course, int64, error)
	Get(context.Context, string) (*domain.Course, error)
	Update(context.Context, string, string, string, string, int) (*domain.Course, error)
	Delete(context.Context, string) error
	GetDetail(context.Context, string) (*service.CourseDetail, error)
	ListPublished(context.Context, int, int) ([]domain.Course, int64, error)
	GetPublishedDetail(context.Context, string) (*service.CourseDetail, error)
	OfficialList(context.Context, int, int) ([]domain.Course, int64, error)
	EnableOfficial(context.Context, string, bool) error
}

func (handler *CourseHandler) CreateOfficial(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
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
	course, err := handler.service.CreateOfficial(c.Request.Context(), request.Title, request.Description, request.CoverImage)
	respond(c, course, err)
}

func (handler *CourseHandler) OfficialList(c *gin.Context) {
	if !requireHandlerRoles(c, "superadmin", "tenant_admin") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.OfficialList(c.Request.Context(), offset, limit)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}

func (handler *CourseHandler) EnableOfficial(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.EnableOfficial(c.Request.Context(), c.Param("id"), request.Enabled); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}

type CourseHandler struct {
	service CourseService
}

func NewCourseHandler(courseService CourseService) *CourseHandler {
	return &CourseHandler{service: courseService}
}

func (handler *CourseHandler) Create(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
		return
	}
	var request struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
		IsOfficial  bool   `json:"is_official"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	var course *domain.Course
	var err error
	if request.IsOfficial {
		course, err = handler.service.CreateOfficial(
			c.Request.Context(), request.Title, request.Description, request.CoverImage,
		)
	} else {
		course, err = handler.service.Create(
			c.Request.Context(), request.Title, request.Description, request.CoverImage,
		)
	}
	respond(c, course, err)
}

func (handler *CourseHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
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
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
		return
	}
	course, err := handler.service.Get(c.Request.Context(), c.Param("id"))
	respond(c, course, err)
}

func (handler *CourseHandler) Update(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
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
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}

func (handler *CourseHandler) Detail(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
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
