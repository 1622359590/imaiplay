package api

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type CourseService interface {
	Create(context.Context, string, string, string) (*domain.Course, error)
	CreateWithType(context.Context, string, string, string, string) (*domain.Course, error)
	CreateWithCategory(context.Context, string, string, string, *string) (*domain.Course, error)
	CreateWithCategoryAndType(context.Context, string, string, string, *string, string) (*domain.Course, error)
	CreateOfficial(
		context.Context, string, string, string, int,
	) (*domain.Course, error)
	CreateOfficialWithCategory(
		context.Context, string, string, string, int, *string,
	) (*domain.Course, error)
	CreateOfficialWithType(
		context.Context, string, string, string, int, string,
	) (*domain.Course, error)
	CreateOfficialWithCategoryAndType(
		context.Context, string, string, string, int, *string, string,
	) (*domain.Course, error)
	List(context.Context, int, int) ([]domain.Course, int64, error)
	Get(context.Context, string) (*domain.Course, error)
	Update(context.Context, string, string, string, string, int) (*domain.Course, error)
	UpdateWithType(context.Context, string, string, string, string, int, string) (*domain.Course, error)
	UpdateWithCategory(
		context.Context, string, string, string, string, int, *string,
	) (*domain.Course, error)
	UpdateWithCategoryAndType(
		context.Context, string, string, string, string, int, *string, string,
	) (*domain.Course, error)
	Delete(context.Context, string) error
	GetDetail(context.Context, string) (*service.CourseDetail, error)
	ListPublished(context.Context, int, int) ([]domain.Course, int64, error)
	GetPublishedDetail(context.Context, string) (*service.LearnerCourseDetail, error)
	OfficialList(context.Context, int, int) ([]domain.Course, int64, error)
	EnableOfficial(context.Context, string, bool) error
}

type optionalCourseCategoryID struct {
	present bool
	value   *string
}

func (value *optionalCourseCategoryID) UnmarshalJSON(data []byte) error {
	value.present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		value.value = nil
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		value.value = nil
		return nil
	}
	value.value = &raw
	return nil
}

func (handler *CourseHandler) CreateOfficial(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		Title       string                   `json:"title" binding:"required"`
		Description string                   `json:"description"`
		CoverImage  string                   `json:"cover_image"`
		Status      *int                     `json:"status" binding:"required"`
		CategoryID  optionalCourseCategoryID `json:"category_id"`
		CourseType  string                   `json:"course_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	var course *domain.Course
	var err error
	if request.CategoryID.present {
		course, err = handler.service.CreateOfficialWithCategoryAndType(
			c.Request.Context(), request.Title, request.Description,
			request.CoverImage, *request.Status, request.CategoryID.value, request.CourseType,
		)
	} else {
		course, err = handler.service.CreateOfficialWithType(
			c.Request.Context(), request.Title, request.Description,
			request.CoverImage, *request.Status, request.CourseType,
		)
	}
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
		Title       string                   `json:"title" binding:"required"`
		Description string                   `json:"description"`
		CoverImage  string                   `json:"cover_image"`
		IsOfficial  bool                     `json:"is_official"`
		Status      *int                     `json:"status"`
		CategoryID  optionalCourseCategoryID `json:"category_id"`
		CourseType  string                   `json:"course_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	var course *domain.Course
	var err error
	if request.IsOfficial {
		status := 1
		if request.Status != nil {
			status = *request.Status
		}
		if request.CategoryID.present {
			course, err = handler.service.CreateOfficialWithCategoryAndType(
				c.Request.Context(), request.Title, request.Description,
				request.CoverImage, status, request.CategoryID.value, request.CourseType,
			)
		} else {
			course, err = handler.service.CreateOfficialWithType(
				c.Request.Context(), request.Title, request.Description,
				request.CoverImage, status, request.CourseType,
			)
		}
	} else {
		if request.CategoryID.present {
			course, err = handler.service.CreateWithCategoryAndType(
				c.Request.Context(), request.Title, request.Description,
				request.CoverImage, request.CategoryID.value, request.CourseType,
			)
		} else {
			course, err = handler.service.CreateWithType(
				c.Request.Context(), request.Title, request.Description, request.CoverImage,
				request.CourseType,
			)
		}
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
		Title       string                   `json:"title" binding:"required"`
		Description string                   `json:"description"`
		CoverImage  string                   `json:"cover_image"`
		Status      *int                     `json:"status" binding:"required"`
		CategoryID  optionalCourseCategoryID `json:"category_id"`
		CourseType  string                   `json:"course_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	var course *domain.Course
	var err error
	if request.CategoryID.present {
		course, err = handler.service.UpdateWithCategoryAndType(
			c.Request.Context(), c.Param("id"), request.Title,
			request.Description, request.CoverImage, *request.Status,
			request.CategoryID.value, request.CourseType,
		)
	} else {
		course, err = handler.service.UpdateWithType(
			c.Request.Context(), c.Param("id"), request.Title,
			request.Description, request.CoverImage, *request.Status, request.CourseType,
		)
	}
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
