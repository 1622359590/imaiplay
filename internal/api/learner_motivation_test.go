package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

type learnerMotivationServiceStub struct {
	motivation      service.LearnerMotivation
	acknowledgedKey string
}

func (stub *learnerMotivationServiceStub) Get(context.Context) (service.LearnerMotivation, error) {
	return stub.motivation, nil
}

func (stub *learnerMotivationServiceStub) Acknowledge(_ context.Context, key string) error {
	stub.acknowledgedKey = key
	return nil
}

func TestLearnerMotivationHandlerReturnsPromptAndAcknowledgesOpaqueKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &learnerMotivationServiceStub{motivation: service.LearnerMotivation{
		Kind: "welcome", PromptKey: "prompt-key", Title: "欢迎开启你的学习旅程",
	}}
	handler := NewLearnerMotivationHandler(stub)
	router := gin.New()
	router.Use(asUser("learner", "tenant-1", "learner-1"))
	router.GET("/learner/motivation", handler.Get)
	router.POST("/learner/motivation/ack", handler.Acknowledge)

	response := requestJSON(t, router, http.MethodGet, "/learner/motivation", "")
	if response.Code != http.StatusOK {
		t.Fatalf("motivation status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data service.LearnerMotivation `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope.Data.Kind != "welcome" || envelope.Data.PromptKey != "prompt-key" {
		t.Fatalf("motivation response = %#v, %v", envelope, err)
	}

	ack := requestJSON(t, router, http.MethodPost, "/learner/motivation/ack", `{"prompt_key":"prompt-key"}`)
	if ack.Code != http.StatusOK || stub.acknowledgedKey != "prompt-key" {
		t.Fatalf("ack status=%d key=%q body=%s", ack.Code, stub.acknowledgedKey, ack.Body.String())
	}
	for _, body := range []string{"", `{}`, `{"prompt_key":"   "}`} {
		invalid := requestJSON(t, router, http.MethodPost, "/learner/motivation/ack", body)
		if invalid.Code != http.StatusBadRequest {
			t.Fatalf("invalid ack %q status=%d body=%s", body, invalid.Code, invalid.Body.String())
		}
	}
}

func TestLearnerMotivationHandlerRequiresLearnerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewLearnerMotivationHandler(&learnerMotivationServiceStub{})
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	router.GET("/learner/motivation", handler.Get)
	router.POST("/learner/motivation/ack", handler.Acknowledge)
	if response := requestJSON(t, router, http.MethodGet, "/learner/motivation", ""); response.Code != http.StatusForbidden {
		t.Fatalf("admin GET status=%d body=%s", response.Code, response.Body.String())
	}
	if response := requestJSON(t, router, http.MethodPost, "/learner/motivation/ack", `{"prompt_key":"prompt-key"}`); response.Code != http.StatusForbidden {
		t.Fatalf("admin ack status=%d body=%s", response.Code, response.Body.String())
	}
}
