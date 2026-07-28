package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type PlanService interface {
	List(context.Context, int, int) ([]domain.Plan, int64, error)
	Create(context.Context, *domain.Plan) (*domain.Plan, error)
	Update(context.Context, *domain.Plan) (*domain.Plan, error)
	Delete(context.Context, string) error
	Assign(context.Context, string, string) (*domain.Tenant, error)
	Current(context.Context) (map[string]interface{}, error)
}

type PlanHandler struct{ service PlanService }

func NewPlanHandler(service PlanService) *PlanHandler { return &PlanHandler{service: service} }

func (handler *PlanHandler) List(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
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
func (handler *PlanHandler) Create(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var plan domain.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	item, err := handler.service.Create(c.Request.Context(), &plan)
	respond(c, item, err)
}
func (handler *PlanHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var plan domain.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	plan.ID = c.Param("id")
	item, err := handler.service.Update(c.Request.Context(), &plan)
	respond(c, item, err)
}
func (handler *PlanHandler) Delete(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
func (handler *PlanHandler) Assign(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		PlanID string `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	item, err := handler.service.Assign(c.Request.Context(), c.Param("tenantID"), request.PlanID)
	respond(c, item, err)
}
func (handler *PlanHandler) Current(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	item, err := handler.service.Current(c.Request.Context())
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, item)
}
