package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestTenantAccessRejectsDisabledTenantRegardlessOfLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	trialEndsAt := time.Now().UTC().Add(time.Hour)
	for _, lifecycleStatus := range []string{"active", "trial"} {
		t.Run(lifecycleStatus, func(t *testing.T) {
			tenant := &domain.Tenant{
				ID:              "tenant-1",
				Status:          0,
				LifecycleStatus: lifecycleStatus,
				TrialEndsAt:     &trialEndsAt,
			}
			router := gin.New()
			router.Use(TenantAccess(tenantRepositoryStub{
				tenantsByID: map[string]*domain.Tenant{tenant.ID: tenant},
			}))
			reached := false
			router.GET("/", func(c *gin.Context) {
				reached = true
				c.Status(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(usercontext.WithUser(
				request.Context(),
				"user-1",
				tenant.ID,
				"user@example.com",
				"learner",
			))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusForbidden || reached {
				t.Fatalf(
					"status = %d, reached = %v; want 403 without handler execution",
					response.Code,
					reached,
				)
			}
		})
	}
}
