package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	Create(
		ctx context.Context,
		email, password, name, role string,
	) (*domain.User, error)
	List(ctx context.Context, offset, limit int) ([]domain.User, int64, error)
	Get(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id, name string, status int) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (handler *UserHandler) Create(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	user, err := handler.service.Create(
		c.Request.Context(), request.Email, request.Password, request.Name, request.Role,
	)
	respond(c, user, err)
}

func (handler *UserHandler) List(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
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

func (handler *UserHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	user, err := handler.service.Get(c.Request.Context(), c.Param("id"))
	respond(c, user, err)
}

func (handler *UserHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		Name   string `json:"name" binding:"required"`
		Status *int   `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	user, err := handler.service.Update(
		c.Request.Context(), c.Param("id"), request.Name, *request.Status,
	)
	respond(c, user, err)
}

func (handler *UserHandler) Delete(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	if err := handler.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
