package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type CourseChapterService interface {
	Create(context.Context, string, string, int) (*domain.CourseChapter, error)
	List(context.Context, string) ([]domain.CourseChapter, error)
	Update(context.Context, string, string, int) (*domain.CourseChapter, error)
	Delete(context.Context, string) error
}

type CourseChapterHandler struct {
	service CourseChapterService
}

func NewCourseChapterHandler(service CourseChapterService) *CourseChapterHandler {
	return &CourseChapterHandler{service: service}
}

func (handler *CourseChapterHandler) Create(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	var request struct {
		Title     string `json:"title" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	chapter, err := handler.service.Create(
		c.Request.Context(), c.Param("id"), request.Title, request.SortOrder,
	)
	respond(c, chapter, err)
}

func (handler *CourseChapterHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	items, err := handler.service.List(c.Request.Context(), c.Param("id"))
	respond(c, items, err)
}

func (handler *CourseChapterHandler) Update(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	var request struct {
		Title     string `json:"title" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	chapter, err := handler.service.Update(
		c.Request.Context(), c.Param("id"), request.Title, request.SortOrder,
	)
	respond(c, chapter, err)
}

func (handler *CourseChapterHandler) Delete(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
