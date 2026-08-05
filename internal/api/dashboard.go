package api

import (
	"context"

	"github.com/gin-gonic/gin"
)

type DashboardService interface {
	Stats(ctx context.Context) (interface{}, error)
}

type DashboardHandler struct {
	service DashboardService
}

func NewDashboardHandler(service DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (handler *DashboardHandler) Get(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
		return
	}
	stats, err := handler.service.Stats(c.Request.Context())
	respond(c, stats, err)
}
