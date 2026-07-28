package middleware

import (
	"net/http"
	"sync"
	"time"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	started time.Time
	count   int
}

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateEntry
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{limit: limit, window: window, entries: make(map[string]rateEntry)}
}

func (limiter *RateLimiter) Allow(key string, now time.Time) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	for entryKey, entry := range limiter.entries {
		if now.Sub(entry.started) >= limiter.window {
			delete(limiter.entries, entryKey)
		}
	}
	entry, ok := limiter.entries[key]
	if !ok || now.Sub(entry.started) >= limiter.window {
		limiter.entries[key] = rateEntry{started: now, count: 1}
		return true
	}
	if entry.count >= limiter.limit {
		return false
	}
	entry.count++
	limiter.entries[key] = entry
	return true
}

func (limiter *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant, _ := tenantcontext.TenantFromContext(c.Request.Context())
		key := c.ClientIP() + "|" + tenant
		if !limiter.Allow(key, time.Now().UTC()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": 42900, "message": "too many requests", "data": nil})
			return
		}
		c.Next()
	}
}
