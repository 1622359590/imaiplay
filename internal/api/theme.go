package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type TenantThemeService interface {
	Get(context.Context) (*domain.Tenant, error)
	Update(context.Context, service.ThemeUpdate) (*domain.Tenant, error)
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
	success(c, themeResponse(theme))
}

func (handler *ThemeHandler) Update(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	var request struct {
		PrimaryColor            string `json:"primary_color"`
		SelectedBackgroundColor string `json:"selected_background_color"`
		SelectedTextColor       string `json:"selected_text_color"`
		SelectedIconColor       string `json:"selected_icon_color"`
		LogoURL                 string `json:"logo_url"`
		WelcomeText             string `json:"welcome_text"`
		BrowserTitle            string `json:"browser_title"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	theme, err := handler.service.Update(c.Request.Context(), service.ThemeUpdate{
		PrimaryColor:            request.PrimaryColor,
		SelectedBackgroundColor: request.SelectedBackgroundColor,
		SelectedTextColor:       request.SelectedTextColor,
		SelectedIconColor:       request.SelectedIconColor,
		LogoURL:                 request.LogoURL,
		WelcomeText:             request.WelcomeText,
		BrowserTitle:            request.BrowserTitle,
	})
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, themeResponse(theme))
}

func themeResponse(theme *domain.Tenant) gin.H {
	return gin.H{
		"primary_color":             theme.PrimaryColor,
		"selected_background_color": theme.SelectedBackgroundColor,
		"selected_text_color":       theme.SelectedTextColor,
		"selected_icon_color":       theme.SelectedIconColor,
		"logo_url":                  theme.LogoURL,
		"welcome_text":              theme.WelcomeText,
		"browser_title":             theme.BrowserTitle,
	}
}
