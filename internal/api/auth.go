package api

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(
		ctx context.Context,
		email, password, name, role string,
	) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (handler *AuthHandler) Register(c *gin.Context) {
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
	user, err := handler.service.Register(
		c.Request.Context(),
		request.Email,
		request.Password,
		request.Name,
		request.Role,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, user)
}

func (handler *AuthHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	token, err := handler.service.Login(
		c.Request.Context(), request.Email, request.Password,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{
		"token": token, "expires_at": time.Now().UTC().Add(24 * time.Hour),
	})
}
