package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/gin-gonic/gin"
)

const maxResourceRequestSize int64 = 1024*1024*1024 + 1024*1024

type ResourceService interface {
	Upload(
		ctx context.Context, name string, reader io.Reader, size int64,
	) (*domain.Resource, error)
	UploadPlatform(
		ctx context.Context, name string, reader io.Reader, size int64,
	) (*domain.Resource, error)
	List(
		ctx context.Context, offset, limit int,
	) ([]domain.Resource, int64, error)
	ListPlatform(
		ctx context.Context, offset, limit int,
	) ([]domain.Resource, int64, error)
	Delete(ctx context.Context, id string) error
	DeletePlatform(ctx context.Context, id string) error
	File(ctx context.Context, id, storageRoot string) (path, contentType, fileName string, err error)
	OpenPlatformCover(
		context.Context, string,
	) (io.ReadCloser, string, string, error)
}

type resourceStreamService interface {
	Open(context.Context, string) (io.ReadCloser, string, string, error)
}

type ResourceHandler struct {
	service     ResourceService
	storageRoot string
}

func NewResourceHandler(service ResourceService, storageRoot ...string) *ResourceHandler {
	root := ""
	if len(storageRoot) > 0 {
		root = storageRoot[0]
	}
	return &ResourceHandler{service: service, storageRoot: root}
}

func (handler *ResourceHandler) Upload(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	handler.upload(c, false)
}

func (handler *ResourceHandler) UploadPlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.upload(c, true)
}

func (handler *ResourceHandler) upload(c *gin.Context, platform bool) {
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
	var resource *domain.Resource
	if platform {
		resource, err = handler.service.UploadPlatform(
			c.Request.Context(), header.Filename, file, header.Size,
		)
	} else {
		resource, err = handler.service.Upload(
			c.Request.Context(), header.Filename, file, header.Size,
		)
	}
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

func (handler *ResourceHandler) ListPlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	offset, limit, err := paginationQuery(c)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	items, total, err := handler.service.ListPlatform(
		c.Request.Context(), offset, limit,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{"items": items, "total": total})
}

func (handler *ResourceHandler) DeletePlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	if err := handler.service.DeletePlatform(
		c.Request.Context(), c.Param("id"),
	); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
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

func (handler *ResourceHandler) File(c *gin.Context) {
	if streamer, ok := handler.service.(resourceStreamService); ok {
		body, contentType, fileName, err := streamer.Open(c.Request.Context(), c.Param("id"))
		if err != nil {
			errorsx.GinResponse(c, err)
			return
		}
		defer body.Close()
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", inlineContentDisposition(fileName))
		c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
		return
	}
	path, contentType, fileName, err := handler.service.File(
		c.Request.Context(), c.Param("id"), handler.storageRoot,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", inlineContentDisposition(fileName))
	c.File(path)
}

func (handler *ResourceHandler) PlatformCover(c *gin.Context) {
	body, contentType, fileName, err := handler.service.OpenPlatformCover(
		c.Request.Context(), c.Param("id"),
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	defer body.Close()
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", inlineContentDisposition(fileName))
	c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
}

func inlineContentDisposition(fileName string) string {
	fileName = strings.NewReplacer("\r", "", "\n", "", `\`, `\\`, `"`, `\"`).Replace(fileName)
	return `inline; filename="` + fileName + `"`
}
