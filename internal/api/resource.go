package api

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

const maxResourceRequestSize int64 = 500*1024*1024 + 1024*1024

type ResourceService interface {
	Upload(
		ctx context.Context, name string, reader io.Reader, size int64,
	) (*domain.Resource, error)
	List(
		ctx context.Context, offset, limit int,
	) ([]domain.Resource, int64, error)
	Delete(ctx context.Context, id string) error
}

type ResourceHandler struct {
	service ResourceService
}

func NewResourceHandler(service ResourceService) *ResourceHandler {
	return &ResourceHandler{service: service}
}

func (handler *ResourceHandler) Upload(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	if c.Request.ContentLength > maxResourceRequestSize {
		errorsx.GinResponse(
			c, errorsx.BadRequest(
				"unsupported file type or size exceeds limit",
			),
		)
		return
	}
	c.Request.Body = http.MaxBytesReader(
		c.Writer, c.Request.Body, maxResourceRequestSize,
	)
	header, err := c.FormFile("file")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			errorsx.GinResponse(
				c, errorsx.BadRequest(
					"unsupported file type or size exceeds limit",
				),
			)
			return
		}
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	file, err := header.Open()
	if err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	defer file.Close()
	resource, err := handler.service.Upload(
		c.Request.Context(), header.Filename, file, header.Size,
	)
	respond(c, resource, err)
}

func (handler *ResourceHandler) List(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.List(
		c.Request.Context(), offset, limit,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}

func (handler *ResourceHandler) Delete(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	if err := handler.service.Delete(
		c.Request.Context(), c.Param("id"),
	); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
