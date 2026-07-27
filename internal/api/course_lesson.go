package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type CourseLessonService interface {
	Create(context.Context, string, string, string, string, int, int) (*domain.CourseLesson, error)
	CreateWithResource(context.Context, string, string, string, string, string, int, int) (*domain.CourseLesson, error)
	List(context.Context, string) ([]domain.CourseLesson, error)
	Update(context.Context, string, string, string, string, int, int) (*domain.CourseLesson, error)
	UpdateWithResource(context.Context, string, string, string, string, string, int, int) (*domain.CourseLesson, error)
	Delete(context.Context, string) error
}

type CourseLessonHandler struct {
	service CourseLessonService
}

func NewCourseLessonHandler(service CourseLessonService) *CourseLessonHandler {
	return &CourseLessonHandler{service: service}
}

type lessonRequest struct {
	Title           string `json:"title" binding:"required"`
	ContentType     string `json:"content_type" binding:"required"`
	ResourceID      string `json:"resource_id"`
	ContentURL      string `json:"content_url"`
	DurationSeconds int    `json:"duration_seconds"`
	SortOrder       int    `json:"sort_order"`
}

func (handler *CourseLessonHandler) Create(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	var request lessonRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	lesson, err := handler.service.CreateWithResource(
		c.Request.Context(), c.Param("id"), request.Title, request.ContentType,
		request.ResourceID, request.ContentURL, request.DurationSeconds, request.SortOrder,
	)
	respond(c, lesson, err)
}

func (handler *CourseLessonHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	items, err := handler.service.List(c.Request.Context(), c.Param("id"))
	respond(c, items, err)
}

func (handler *CourseLessonHandler) Update(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	var request lessonRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	lesson, err := handler.service.UpdateWithResource(
		c.Request.Context(), c.Param("id"), request.Title, request.ContentType,
		request.ResourceID, request.ContentURL, request.DurationSeconds, request.SortOrder,
	)
	respond(c, lesson, err)
}

func (handler *CourseLessonHandler) Delete(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
