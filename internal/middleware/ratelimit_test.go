package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterWindowAndCleanup(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	now := time.Unix(100, 0)
	if !limiter.Allow("ip|tenant", now) || !limiter.Allow("ip|tenant", now) || limiter.Allow("ip|tenant", now) {
		t.Fatal("limit was not enforced")
	}
	if !limiter.Allow("other|tenant", now) {
		t.Fatal("keys should be isolated")
	}
	if !limiter.Allow("ip|tenant", now.Add(time.Minute)) {
		t.Fatal("window did not reset")
	}
	if len(limiter.entries) != 1 {
		t.Fatalf("expired entries were not cleaned: %d", len(limiter.entries))
	}
}

func TestRateLimiterHandlerReturns429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter(1, time.Minute)
	router := gin.New()
	router.Use(limiter.Handler())
	router.GET("/login", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	for i, want := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/login", nil))
		if response.Code != want {
			t.Fatalf("request %d status=%d, want %d", i, response.Code, want)
		}
	}
}
