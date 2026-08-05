package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type ProgressService interface {
	Report(
		ctx context.Context, lessonID string, positionSeconds, percent int,
		watchedSecondsDelta int, reportID string,
	) (*domain.LessonProgress, error)
	Get(ctx context.Context, lessonID string) (*domain.LessonProgress, error)
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
		PositionSeconds     int    `json:"position_seconds"`
		ProgressPercent     int    `json:"progress_percent"`
		WatchedSecondsDelta int    `json:"watched_seconds_delta"`
		ReportID            string `json:"report_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	progress, err := handler.service.Report(
		c.Request.Context(), c.Param("id"),
		request.PositionSeconds, request.ProgressPercent,
		request.WatchedSecondsDelta, request.ReportID,
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
