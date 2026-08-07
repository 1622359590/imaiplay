package api

import (
	"context"
	"net/http"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	userservice "github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	Create(
		ctx context.Context,
		email, password, name, role string,
	) (*domain.User, error)
	CreateWithPhone(ctx context.Context, email, phone, password, name, role string) (*domain.User, error)
	Import(ctx context.Context, rows []userservice.UserImportRow) (userservice.UserImportResult, error)
	List(ctx context.Context, offset, limit int) ([]domain.User, int64, error)
	Get(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id, name string, status int, password string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
	ResetTenantAdminPassword(ctx context.Context, id, password string) error
}

func (handler *UserHandler) Import(c *gin.Context) {
	if !requireHandlerRole(c, "tenant_admin") {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
	header, err := c.FormFile("file")
	if err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("请上传导入文件，文件不能超过 10MB"))
		return
	}
	file, err := header.Open()
	if err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("无法读取导入文件"))
		return
	}
	defer func() { _ = file.Close() }()
	rows, err := userservice.ParseUserImportFile(header.Filename, file)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	result, err := handler.service.Import(c.Request.Context(), rows)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, result)
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
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	user, err := handler.service.CreateWithPhone(
		c.Request.Context(), request.Email, request.Phone, request.Password, request.Name, request.Role,
	)
	respond(c, user, err)
}

func (handler *UserHandler) List(c *gin.Context) {
	if !requireHandlerRoles(c, "tenant_admin", "superadmin") {
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
		Name     string `json:"name" binding:"required"`
		Status   *int   `json:"status" binding:"required"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	user, err := handler.service.Update(
		c.Request.Context(), c.Param("id"), request.Name, *request.Status, request.Password,
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

func (handler *UserHandler) ResetTenantAdminPassword(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.ResetTenantAdminPassword(c.Request.Context(), c.Param("id"), request.Password); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
