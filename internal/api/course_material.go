package api

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type CourseMaterialService interface {
	Add(context.Context, string, service.CourseMaterialInput) (*domain.CourseMaterial, error)
	Update(context.Context, string, string, service.CourseMaterialInput) (*domain.CourseMaterial, error)
	Remove(context.Context, string, string) error
	ListForManager(context.Context, string) ([]domain.CourseMaterial, error)
	OpenForLearner(context.Context, string) (io.ReadCloser, string, string, error)
}

func (handler *CourseMaterialHandler) Download(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	body, contentType, fileName, err := handler.service.OpenForLearner(c.Request.Context(), c.Param("id"))
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	defer body.Close()
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", attachmentContentDisposition(fileName))
	c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
}

func attachmentContentDisposition(fileName string) string {
	fileName = strings.Map(func(value rune) rune {
		switch value {
		case '\r', '\n', '"', '\\', '/':
			return -1
		default:
			return value
		}
	}, fileName)
	fileName = strings.TrimSpace(fileName)
	if fileName == "" || fileName == "." || fileName == ".." {
		fileName = "download"
	}
	return `attachment; filename="` + fileName + `"`
}

type CourseMaterialHandler struct{ service CourseMaterialService }

func NewCourseMaterialHandler(materialService CourseMaterialService) *CourseMaterialHandler {
	return &CourseMaterialHandler{service: materialService}
}

func (handler *CourseMaterialHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor", "superadmin") {
		return
	}
	items, err := handler.service.ListForManager(c.Request.Context(), c.Param("id"))
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items})
}

func (handler *CourseMaterialHandler) Add(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "superadmin") {
		return
	}
	input, ok := bindCourseMaterialInput(c)
	if !ok {
		return
	}
	material, err := handler.service.Add(c.Request.Context(), c.Param("id"), input)
	respond(c, material, err)
}

func (handler *CourseMaterialHandler) Update(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "superadmin") {
		return
	}
	input, ok := bindCourseMaterialInput(c)
	if !ok {
		return
	}
	material, err := handler.service.Update(c.Request.Context(), c.Param("id"), c.Param("materialID"), input)
	respond(c, material, err)
}

func (handler *CourseMaterialHandler) Remove(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "superadmin") {
		return
	}
	if err := handler.service.Remove(c.Request.Context(), c.Param("id"), c.Param("materialID")); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}

func bindCourseMaterialInput(c *gin.Context) (service.CourseMaterialInput, bool) {
	var request struct {
		ResourceID  string `json:"resource_id" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return service.CourseMaterialInput{}, false
	}
	return service.CourseMaterialInput{ResourceID: request.ResourceID, DisplayName: request.DisplayName, SortOrder: request.SortOrder}, true
}
