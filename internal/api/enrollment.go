package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type EnrollmentService interface {
	Enroll(
		ctx context.Context, courseID, userID string,
	) (*domain.CourseEnrollment, error)
	ListByCourse(
		ctx context.Context, courseID string,
	) ([]domain.CourseEnrollment, error)
	Remove(ctx context.Context, id string) error
}

type EnrollmentHandler struct {
	service EnrollmentService
}

func NewEnrollmentHandler(service EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{service: service}
}

func (handler *EnrollmentHandler) Enroll(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	enrollment, err := handler.service.Enroll(
		c.Request.Context(), c.Param("id"), request.UserID,
	)
	respond(c, enrollment, err)
}

func (handler *EnrollmentHandler) ListByCourse(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	items, err := handler.service.ListByCourse(
		c.Request.Context(), c.Param("id"),
	)
	respond(c, items, err)
}

func (handler *EnrollmentHandler) Remove(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	if err := handler.service.Remove(
		c.Request.Context(), c.Param("id"),
	); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
