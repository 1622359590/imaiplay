package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

type TenantThemeService interface {
	Get(context.Context) (*domain.Tenant, error)
	Update(context.Context, string, string, string, string) (*domain.Tenant, error)
}

type ThemeHandler struct{ service TenantThemeService }

func NewThemeHandler(service TenantThemeService) *ThemeHandler {
	return &ThemeHandler{service: service}
}

func (handler *ThemeHandler) Get(c *gin.Context) {
	theme, err := handler.service.Get(c.Request.Context())
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"primary_color": theme.PrimaryColor, "logo_url": theme.LogoURL, "welcome_text": theme.WelcomeText, "browser_title": theme.BrowserTitle})
}

func (handler *ThemeHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		PrimaryColor string `json:"primary_color"`
		LogoURL      string `json:"logo_url"`
		WelcomeText  string `json:"welcome_text"`
		BrowserTitle string `json:"browser_title"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	theme, err := handler.service.Update(c.Request.Context(), request.PrimaryColor, request.LogoURL, request.WelcomeText, request.BrowserTitle)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"primary_color": theme.PrimaryColor, "logo_url": theme.LogoURL, "welcome_text": theme.WelcomeText, "browser_title": theme.BrowserTitle})
}
