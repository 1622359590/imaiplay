package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type LearnerOverviewService interface {
	Get(ctx context.Context) (service.LearnerOverview, error)
	GetRecent(
		ctx context.Context,
		offset, limit int,
	) ([]service.RecentLearningItem, int64, error)
}

type LearnerOverviewHandler struct {
	service LearnerOverviewService
}

func NewLearnerOverviewHandler(
	service LearnerOverviewService,
) *LearnerOverviewHandler {
	return &LearnerOverviewHandler{service: service}
}

func (handler *LearnerOverviewHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	overview, err := handler.service.Get(c.Request.Context())
	respond(c, overview, err)
}

func (handler *LearnerOverviewHandler) Recent(c *gin.Context) {
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
