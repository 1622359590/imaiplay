package api

import (
	"context"
	"strings"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type LearnerMotivationService interface {
	Get(context.Context) (service.LearnerMotivation, error)
	Acknowledge(context.Context, string) error
}

type LearnerMotivationHandler struct {
	service LearnerMotivationService
}

func NewLearnerMotivationHandler(service LearnerMotivationService) *LearnerMotivationHandler {
	return &LearnerMotivationHandler{service: service}
}

func (handler *LearnerMotivationHandler) Get(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	motivation, err := handler.service.Get(c.Request.Context())
	respond(c, motivation, err)
}

func (handler *LearnerMotivationHandler) Acknowledge(c *gin.Context) {
	if !requireHandlerRole(c, "learner") {
		return
	}
	var request struct {
		PromptKey string `json:"prompt_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.PromptKey) == "" {
		errorsx.GinResponse(c, errorsx.BadRequest("prompt key is required"))
		return
	}
	if err := handler.service.Acknowledge(c.Request.Context(), request.PromptKey); err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, gin.H{})
}
