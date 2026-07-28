package api

import (
	"context"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/sms"
	"github.com/gin-gonic/gin"
)

type SMSConfigService interface {
	Get() sms.Config
	Save(sms.Config) error
	SendTest(context.Context, string) error
}

type SMSConfigHandler struct{ service SMSConfigService }

func NewSMSConfigHandler(service SMSConfigService) *SMSConfigHandler {
	return &SMSConfigHandler{service: service}
}

func (handler *SMSConfigHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	success(c, handler.service.Get())
}

func (handler *SMSConfigHandler) Save(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		Provider        string `json:"provider"`
		AccessKeyID     string `json:"access_key_id"`
		AccessKeySecret string `json:"access_key_secret"`
		SignName        string `json:"sign_name"`
		TemplateCode    string `json:"template_code"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.Save(sms.Config{Provider: request.Provider, AccessKeyID: request.AccessKeyID, AccessKeySecret: request.AccessKeySecret, SignName: request.SignName, TemplateCode: request.TemplateCode}); err != nil {
		errorsx.GinResponse(c, errorsx.Internal("save sms config failed"))
		return
	}
	success(c, handler.service.Get())
}

func (handler *SMSConfigHandler) Test(c *gin.Context) {
	if !requireHandlerRole(c, "superadmin") {
		return
	}
	var request struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	if err := handler.service.SendTest(c.Request.Context(), request.Phone); err != nil {
		errorsx.GinResponse(c, errorsx.Internal("send test sms failed"))
		return
	}
	success(c, gin.H{})
}
