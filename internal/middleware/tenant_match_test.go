package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestTenantMatchRejectsSessionFromAnotherTenant(t *testing.T) {
	acme := &domain.Tenant{ID: "tenant-acme", Code: "acme"}
	router := tenantMatchRouter(
		tenantRepositoryStub{
			tenantsByCode: map[string]*domain.Tenant{acme.Code: acme},
		},
		"tenant-bravo",
		"learner",
		acme.Code,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTenantMatchAcceptsMatchingSession(t *testing.T) {
	acme := &domain.Tenant{ID: "tenant-acme", Code: "acme"}
	router := tenantMatchRouter(
		tenantRepositoryStub{
			tenantsByCode: map[string]*domain.Tenant{acme.Code: acme},
		},
		acme.ID,
		"learner",
		acme.Code,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTenantMatchAllowsPlatformContext(t *testing.T) {
	router := tenantMatchRouter(
		tenantRepositoryStub{},
		"tenant-acme",
		"learner",
		usercontext.UnknownTenant,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTenantMatchAllowsSuperadminAcrossPortals(t *testing.T) {
	router := tenantMatchRouter(
		tenantRepositoryStub{},
		"",
		"superadmin",
		"acme",
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func tenantMatchRouter(
	tenants tenantRepositoryStub,
	sessionTenantID string,
	role string,
	portalCode string,
) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := usercontext.WithTenant(
			c.Request.Context(),
			portalCode,
			usercontext.SourceHeaderCode,
		)
		ctx = usercontext.WithUser(
			ctx,
			"user-id",
			sessionTenantID,
			"user@example.com",
			role,
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(TenantMatch(tenants))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}
