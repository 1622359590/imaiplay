package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(
		ctx context.Context,
		email, password, name, role string,
	) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	LoginWithRefresh(ctx context.Context, email, password string) (*service.TokenPair, error)
	IssueTokens(ctx context.Context, user *domain.User) (*service.TokenPair, error)
	Refresh(ctx context.Context, token string) (*service.TokenPair, error)
	Logout(ctx context.Context, token string) error
	BootstrapSuperadmin(ctx context.Context, email, name, password string) (*domain.User, *service.TokenPair, error)
}

type AuthHandler struct {
	service AuthService
}

func (handler *AuthHandler) BootstrapSuperadmin(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	user, pair, err := handler.service.BootstrapSuperadmin(c.Request.Context(), request.Email, request.Name, request.Password)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"user": user, "token": pair.AccessToken, "refresh_token": pair.RefreshToken, "expires_at": pair.ExpiresAt})
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
	pair, err := handler.service.IssueTokens(c.Request.Context(), user)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"user": user, "token": pair.AccessToken, "refresh_token": pair.RefreshToken, "expires_at": pair.ExpiresAt})
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
	pair, err := handler.service.LoginWithRefresh(
		c.Request.Context(), request.Email, request.Password,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{
		"token": pair.AccessToken, "refresh_token": pair.RefreshToken, "expires_at": pair.ExpiresAt,
	})
}

func (handler *AuthHandler) Refresh(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	pair, err := handler.service.Refresh(c.Request.Context(), request.RefreshToken)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, pair)
}

func (handler *AuthHandler) Logout(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil && c.Request.ContentLength != 0 {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
