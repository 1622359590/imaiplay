package api

import (
	"net/http"
	"strings"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/gin-gonic/gin"
)

func TestStudentCoursesOnlyExposePublishedCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	admin := courseTestRouter(services, "tenant_admin", tenant.ID)
	published := requestJSON(t, admin, http.MethodPost, "/courses",
		`{"title":"Published","description":"","cover_image":""}`)
	publishedID := responseID(t, published.Body.Bytes())
	if response := requestJSON(t, admin, http.MethodPut, "/courses/"+publishedID,
		`{"title":"Published","description":"","cover_image":"","status":1}`); response.Code != http.StatusOK {
		t.Fatalf("publish status=%d body=%s", response.Code, response.Body.String())
	}
	draft := requestJSON(t, admin, http.MethodPost, "/courses",
		`{"title":"Draft","description":"","cover_image":""}`)
	draftID := responseID(t, draft.Body.Bytes())

	handler := NewCourseHandler(services.courses)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, "learner-1"))
	router.GET("/api/v1/courses", handler.PublishedList)
	router.GET("/api/v1/courses/:id", handler.PublishedDetail)

	list := requestJSON(t, router, http.MethodGet, "/api/v1/courses", "")
	if list.Code != http.StatusOK ||
		!strings.Contains(list.Body.String(), "Published") ||
		strings.Contains(list.Body.String(), `"title":"Draft"`) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if response := requestJSON(t, router, http.MethodGet,
		"/api/v1/courses/"+publishedID, ""); response.Code != http.StatusOK {
		t.Fatalf("published detail status=%d body=%s",
			response.Code, response.Body.String())
	}
	if response := requestJSON(t, router, http.MethodGet,
		"/api/v1/courses/"+draftID, ""); response.Code != http.StatusNotFound {
		t.Fatalf("draft detail status=%d body=%s",
			response.Code, response.Body.String())
	}
}

func TestStudentCoursesRejectNonLearner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewCourseHandler(services.courses)
	router := gin.New()
	router.Use(asUser("instructor", tenant.ID, "instructor-1"))
	router.GET("/api/v1/courses", handler.PublishedList)

	response := requestJSON(t, router, http.MethodGet, "/api/v1/courses", "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func asUser(role, tenantID, userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := usercontext.WithUser(
			c.Request.Context(), userID, tenantID, userID+"@example.com", role,
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
