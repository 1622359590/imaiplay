package api

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/1622359590/imaiplay/internal/sms"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func TestSMSConfigHandlerDoesNotEchoSecret(t *testing.T) {
	store, err := sms.NewConfigStore(filepath.Join(t.TempDir(), "sms.json"), "jwt", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	handler := NewSMSConfigHandler(store)
	router := gin.New()
	router.Use(asRole("superadmin", ""))
	router.GET("/sms-config", handler.Get)
	router.PUT("/sms-config", handler.Save)
	saved := requestJSON(t, router, http.MethodPut, "/sms-config", `{"provider":"aliyun","access_key_id":"id","access_key_secret":"secret","sign_name":"ImaiPlay","template_code":"SMS_1"}`)
	if saved.Code != http.StatusOK {
		t.Fatalf("save status=%d body=%s", saved.Code, saved.Body.String())
	}
	if response := requestJSON(t, router, http.MethodGet, "/sms-config", ""); response.Code != http.StatusOK || response.Body.String() == "" || contains(response.Body.String(), "secret") {
		t.Fatalf("get response=%d body=%s", response.Code, response.Body.String())
	}
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
