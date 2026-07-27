package errorsx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinResponseMapsApplicationErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{"bad request", BadRequest("bad"), http.StatusBadRequest, 40000},
		{"unauthorized", Unauthorized("no token"), http.StatusUnauthorized, 40100},
		{"forbidden", Forbidden("denied"), http.StatusForbidden, 40300},
		{"not found", NotFound("missing"), http.StatusNotFound, 40400},
		{"conflict", Conflict("exists"), http.StatusConflict, 40900},
		{"internal", Internal("failed"), http.StatusInternalServerError, 50000},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError, 50000},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			GinResponse(ctx, tt.err)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			var body struct {
				Code    int         `json:"code"`
				Message string      `json:"message"`
				Data    interface{} `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Code != tt.wantCode || body.Data != nil {
				t.Fatalf("body = %#v, want code=%d data=nil", body, tt.wantCode)
			}
		})
	}
}
