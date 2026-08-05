package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestProgressHandlerReportGetAndRecent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	admin := withRole("tenant_admin", tenant.ID, "admin-1")
	learner, err := services.users.Create(
		admin, "learner@example.com", "password123", "Learner", "learner",
	)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	course, err := services.courses.Create(admin, "Course", "", "")
	if err != nil {
		t.Fatalf("create course: %v", err)
	}
	chapter, err := services.chapters.Create(admin, course.ID, "Chapter", 1)
	if err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson, err := services.lessons.Create(
		admin, chapter.ID, "Lesson", "video", "/uploads/video.mp4", 100, 1,
	)
	if err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	handler := NewProgressHandler(services.progress)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, learner.ID))
	router.POST("/lessons/:id/progress", handler.Report)
	router.GET("/lessons/:id/progress", handler.Get)
	router.GET("/recent-learning", handler.Recent)

	forbidden := requestJSON(t, router, http.MethodPost,
		"/lessons/"+lesson.ID+"/progress",
		`{"position_seconds":10,"progress_percent":10}`)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("not enrolled status=%d body=%s",
			forbidden.Code, forbidden.Body.String())
	}
	if _, err := services.enrollments.Enroll(admin, course.ID, learner.ID, domain.AssignmentRequired); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	reported := requestJSON(t, router, http.MethodPost,
		"/lessons/"+lesson.ID+"/progress",
		`{"position_seconds":50,"progress_percent":50}`)
	if reported.Code != http.StatusOK ||
		!strings.Contains(reported.Body.String(), `"progress_percent":50`) {
		t.Fatalf("Report status=%d body=%s", reported.Code, reported.Body.String())
	}
	if response := requestJSON(t, router, http.MethodGet,
		"/lessons/"+lesson.ID+"/progress", ""); response.Code != http.StatusOK {
		t.Fatalf("Get status=%d body=%s", response.Code, response.Body.String())
	}
	recent := requestJSON(t, router, http.MethodGet, "/recent-learning", "")
	if recent.Code != http.StatusOK ||
		!strings.Contains(recent.Body.String(), `"total":1`) ||
		!strings.Contains(recent.Body.String(), `"title":"Course"`) {
		t.Fatalf("Recent status=%d body=%s", recent.Code, recent.Body.String())
	}
}

func TestProgressHandlerRejectsInvalidBodyAndRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewProgressHandler(services.progress)
	learner := gin.New()
	learner.Use(asUser("learner", "tenant-1", "learner-1"))
	learner.POST("/lessons/:id/progress", handler.Report)
	if response := requestJSON(t, learner, http.MethodPost,
		"/lessons/lesson-1/progress", `{"progress_percent":101}`); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid body status=%d body=%s",
			response.Code, response.Body.String())
	}
	admin := gin.New()
	admin.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	admin.GET("/recent-learning", handler.Recent)
	if response := requestJSON(
		t, admin, http.MethodGet, "/recent-learning", "",
	); response.Code != http.StatusForbidden {
		t.Fatalf("admin status=%d body=%s", response.Code, response.Body.String())
	}
}
