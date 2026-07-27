package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type ProgressService interface {
	Report(
		ctx context.Context, lessonID string, positionSeconds, percent int,
	) (*domain.LessonProgress, error)
	Get(ctx context.Context, lessonID string) (*domain.LessonProgress, error)
	GetRecent(
		ctx context.Context, offset, limit int,
	) ([]service.RecentLearnItem, int64, error)
}

type ProgressHandler struct {
	service ProgressService
}

func NewProgressHandler(service ProgressService) *ProgressHandler {
	return &ProgressHandler{service: service}
}

func (handler *ProgressHandler) Report(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	var request struct {
		PositionSeconds int `json:"position_seconds"`
		ProgressPercent int `json:"progress_percent"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	progress, err := handler.service.Report(
		c.Request.Context(), c.Param("id"),
		request.PositionSeconds, request.ProgressPercent,
	)
	respond(c, progress, err)
}

func (handler *ProgressHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	progress, err := handler.service.Get(c.Request.Context(), c.Param("id"))
	respond(c, progress, err)
}

func (handler *ProgressHandler) Recent(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.GetRecent(
		c.Request.Context(), offset, limit,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}
