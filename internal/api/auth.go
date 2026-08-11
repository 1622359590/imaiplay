package api

import (
	"context"
	"crypto/subtle"
	"strings"

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
	RegisterWithPhone(ctx context.Context, email, phone, password, name, role string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	LoginWithRefresh(ctx context.Context, email, password string) (*service.TokenPair, error)
	BeginLogin(ctx context.Context, identifier, password string) (*service.LoginOutcome, error)
	SelectTenant(ctx context.Context, selectionToken, tenantCode string) (*service.LoginOutcome, error)
	IssueTokens(ctx context.Context, user *domain.User) (*service.TokenPair, error)
	Refresh(ctx context.Context, token string) (*service.TokenPair, error)
	Logout(ctx context.Context, token string) error
	BootstrapSuperadmin(ctx context.Context, email, name, password string) (*domain.User, *service.TokenPair, error)
	ForgotPassword(ctx context.Context, phone string) error
	ResetPassword(ctx context.Context, phone, code, newPassword string) error
	SendLoginCode(context.Context, string) error
	LoginWithCode(context.Context, string, string) (*service.LoginOutcome, error)
	CurrentUser(context.Context) (service.AuthUser, error)
}

type AuthHandler struct {
	service         AuthService
	bootstrapSecret string
}

func (handler *AuthHandler) BootstrapSuperadmin(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
		Secret   string `json:"bootstrap_secret" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	configuredSecret := strings.TrimSpace(handler.bootstrapSecret)
	if configuredSecret == "" || subtle.ConstantTimeCompare(
		[]byte(configuredSecret), []byte(request.Secret),
	) != 1 {
		errorsx.GinResponse(c, errorsx.Forbidden("超级管理员初始化未启用或密钥无效"))
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

func (handler *AuthHandler) WithBootstrapSecret(secret string) *AuthHandler {
	handler.bootstrapSecret = secret
	return handler
}

func (handler *AuthHandler) Register(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if request.Role != "learner" {
		errorsx.GinResponse(c, errorsx.Forbidden("公开注册仅支持学员账号"))
		return
	}
	user, err := handler.service.RegisterWithPhone(
		c.Request.Context(),
		request.Email,
		request.Phone, request.Password,
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
		Identifier string `json:"identifier"`
		Email      string `json:"email"`
		Password   string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	identifier := request.Identifier
	if identifier == "" {
		identifier = request.Email
	}
	if identifier == "" || request.Password == "" {
		errorsx.GinResponse(c, errorsx.BadRequest("identifier and password are required"))
		return
	}
	outcome, err := handler.service.BeginLogin(
		c.Request.Context(), identifier, request.Password,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, loginOutcomeResponse(outcome))
}

func (handler *AuthHandler) SelectTenant(c *gin.Context) {
	var request struct {
		SelectionToken string `json:"selection_token" binding:"required"`
		TenantCode     string `json:"tenant_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	outcome, err := handler.service.SelectTenant(
		c.Request.Context(),
		request.SelectionToken,
		request.TenantCode,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, loginOutcomeResponse(outcome))
}

func (handler *AuthHandler) SendLoginCode(c *gin.Context) {
	var request struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.SendLoginCode(c.Request.Context(), request.Phone); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"message": "if the phone exists, a login code has been sent"})
}

func (handler *AuthHandler) LoginWithCode(c *gin.Context) {
	var request struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	outcome, err := handler.service.LoginWithCode(c.Request.Context(), request.Phone, request.Code)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, loginOutcomeResponse(outcome))
}

func (handler *AuthHandler) Me(c *gin.Context) {
	user, err := handler.service.CurrentUser(c.Request.Context())
	respond(c, user, err)
}

func (handler *AuthHandler) ForgotPassword(c *gin.Context) {
	var request struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.ForgotPassword(c.Request.Context(), request.Phone); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"message": "if the phone exists, a verification code has been sent"})
}

func (handler *AuthHandler) ResetPassword(c *gin.Context) {
	var request struct {
		Phone       string `json:"phone" binding:"required"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.ResetPassword(c.Request.Context(), request.Phone, request.Code, request.NewPassword); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
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

func loginOutcomeResponse(outcome *service.LoginOutcome) gin.H {
	response := gin.H{
		"requires_tenant_selection": outcome.RequiresTenantSelection,
	}
	if outcome.RequiresTenantSelection {
		response["selection_token"] = outcome.SelectionToken
		response["organizations"] = outcome.Organizations
		return response
	}
	response["user"] = outcome.User
	if outcome.Tenant != nil {
		response["tenant"] = outcome.Tenant
	}
	if outcome.Pair != nil {
		response["token"] = outcome.Pair.AccessToken
		response["refresh_token"] = outcome.Pair.RefreshToken
		response["expires_at"] = outcome.Pair.ExpiresAt
	}
	return response
}
