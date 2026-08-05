package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/gin-gonic/gin"
)

const maxResourceRequestSize int64 = 1024*1024*1024 + 1024*1024
const maxAttachmentRequestSize int64 = 200*1024*1024 + 1024*1024

type ResourceService interface {
	Upload(
		ctx context.Context, name string, reader io.Reader, size int64,
	) (*domain.Resource, error)
	UploadPlatform(
		ctx context.Context, name string, reader io.Reader, size int64,
	) (*domain.Resource, error)
	UploadAttachment(
		ctx context.Context, name string, reader io.Reader, size int64,
	) (*domain.Resource, error)
	UploadPlatformAttachment(
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

type LearnerAccessService interface {
	AuthorizeLessonResource(context.Context, string) (*domain.Course, error)
}

type ResourceHandler struct {
	service        ResourceService
	storageRoot    string
	playbackSecret string
	learnerAccess  LearnerAccessService
}

func (handler *ResourceHandler) WithLearnerAccess(access LearnerAccessService) *ResourceHandler {
	handler.learnerAccess = access
	return handler
}

func (handler *ResourceHandler) WithPlaybackSecret(secret string) *ResourceHandler {
	handler.playbackSecret = secret
	return handler
}

func NewResourceHandler(service ResourceService, storageRoot ...string) *ResourceHandler {
	root := ""
	if len(storageRoot) > 0 {
		root = storageRoot[0]
	}
	return &ResourceHandler{service: service, storageRoot: root}
}

func (handler *ResourceHandler) Upload(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	handler.upload(c, false, false)
}

func (handler *ResourceHandler) UploadPlatform(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.upload(c, true, false)
}

func (handler *ResourceHandler) UploadAttachment(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
		return
	}
	handler.upload(c, false, true)
}

func (handler *ResourceHandler) UploadPlatformAttachment(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	handler.upload(c, true, true)
}

func (handler *ResourceHandler) upload(c *gin.Context, platform, attachment bool) {
	requestLimit := maxResourceRequestSize
	if attachment {
		requestLimit = maxAttachmentRequestSize
	}
	if c.Request.ContentLength > requestLimit {
		errorsx.GinResponse(
			c, errorsx.BadRequest(
				"unsupported file type or size exceeds limit",
			),
		)
		return
	}
	c.Request.Body = http.MaxBytesReader(
		c.Writer, c.Request.Body, requestLimit,
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
	if platform && attachment {
		resource, err = handler.service.UploadPlatformAttachment(
			c.Request.Context(), header.Filename, file, header.Size,
		)
	} else if attachment {
		resource, err = handler.service.UploadAttachment(
			c.Request.Context(), header.Filename, file, header.Size,
		)
	} else if platform {
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
	if !requireHandlerRoles(c, "tenant_admin", "instructor") {
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
	if _, _, _, role, ok := usercontext.UserFromContext(c.Request.Context()); ok && role == "learner" {
		if _, err := handler.authorizeLearnerResource(c.Request.Context(), c.Param("id")); err != nil {
			errorsx.GinResponse(c, err)
			return
		}
	}
	if streamer, ok := handler.service.(resourceStreamService); ok {
		body, contentType, fileName, err := streamer.Open(c.Request.Context(), c.Param("id"))
		if err != nil {
			errorsx.GinResponse(c, err)
			return
		}
		handler.serve(c, body, contentType, fileName)
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

func (handler *ResourceHandler) PlaybackURL(c *gin.Context) {
	if handler.playbackSecret == "" {
		errorsx.GinResponse(c, errorsx.Internal("playback is not configured"))
		return
	}
	course, err := handler.authorizeLearnerResource(c.Request.Context(), c.Param("id"))
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	if course == nil || course.ID == "" {
		errorsx.GinResponse(c, errorsx.NotFound("resource not found"))
		return
	}
	streamer, ok := handler.service.(resourceStreamService)
	if !ok {
		errorsx.GinResponse(c, errorsx.NotFound("resource not found"))
		return
	}
	body, _, _, err := streamer.Open(c.Request.Context(), c.Param("id"))
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	_ = body.Close()
	userID, tenantID, email, role, ok := usercontext.UserFromContext(c.Request.Context())
	if !ok {
		errorsx.GinResponse(c, errorsx.Unauthorized("missing or invalid token"))
		return
	}
	ticket, err := security.GeneratePlaybackToken(
		c.Param("id"), course.ID, userID, tenantID, email, role,
		handler.playbackSecret, 2*time.Minute,
	)
	if err != nil {
		errorsx.GinResponse(c, errorsx.Internal("generate playback URL failed"))
		return
	}
	success(c, gin.H{
		"url": "/api/v1/resource-playback/" + url.PathEscape(c.Param("id")) +
			"?ticket=" + url.QueryEscape(ticket),
	})
}

func (handler *ResourceHandler) Playback(c *gin.Context) {
	claims, err := security.ValidatePlaybackToken(
		c.Query("ticket"), handler.playbackSecret,
	)
	if err != nil || claims.ResourceID != c.Param("id") {
		errorsx.GinResponse(c, errorsx.Unauthorized("invalid playback ticket"))
		return
	}
	ctx := usercontext.WithUser(
		c.Request.Context(), claims.UserID, claims.TenantID,
		claims.Email, claims.Role,
	)
	course, err := handler.authorizeLearnerResource(ctx, c.Param("id"))
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	if course == nil || course.ID != claims.CourseID {
		errorsx.GinResponse(c, errorsx.NotFound("resource not found"))
		return
	}
	streamer, ok := handler.service.(resourceStreamService)
	if !ok {
		errorsx.GinResponse(c, errorsx.NotFound("resource not found"))
		return
	}
	body, contentType, fileName, err := streamer.Open(ctx, c.Param("id"))
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	handler.serve(c, body, contentType, fileName)
}

func (handler *ResourceHandler) authorizeLearnerResource(
	ctx context.Context, resourceID string,
) (*domain.Course, error) {
	if handler.learnerAccess == nil {
		return nil, errorsx.Internal("learner access unavailable")
	}
	return handler.learnerAccess.AuthorizeLessonResource(ctx, resourceID)
}

func (handler *ResourceHandler) serve(
	c *gin.Context, body io.ReadCloser, contentType, fileName string,
) {
	defer body.Close()
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", inlineContentDisposition(fileName))
	if seeker, ok := body.(io.ReadSeeker); ok {
		http.ServeContent(c.Writer, c.Request, fileName, time.Time{}, seeker)
		return
	}
	c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
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
