package api

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	if _, err := services.courses.Update(admin, course.ID, course.Title, course.Description, course.CoverImage, 1); err != nil {
		t.Fatalf("publish course: %v", err)
	}
	handler := NewProgressHandler(services.progress)
	overviewHandler := NewLearnerOverviewHandler(services.overview)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, learner.ID))
	router.POST("/lessons/:id/progress", handler.Report)
	router.GET("/lessons/:id/progress", handler.Get)
	router.GET("/recent-learning", overviewHandler.Recent)

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
	var reportCount int64
	if err := services.database.Model(&domain.LearningTimeReport{}).Count(&reportCount).Error; err != nil || reportCount != 0 {
		t.Fatalf("legacy report count = %d, %v", reportCount, err)
	}
	for index := 0; index < 2; index++ {
		watched := requestJSON(t, router, http.MethodPost,
			"/lessons/"+lesson.ID+"/progress",
			`{"position_seconds":65,"progress_percent":65,"watched_seconds_delta":15,"report_id":"heartbeat-1","session_id":"session-1"}`)
		if watched.Code != http.StatusOK {
			t.Fatalf("heartbeat %d status=%d body=%s", index, watched.Code, watched.Body.String())
		}
	}
	var stat domain.LearningDailyStat
	if err := services.database.Where(
		"tenant_id = ? AND user_id = ?", tenant.ID, learner.ID,
	).First(&stat).Error; err != nil || stat.DurationSeconds != 15 {
		t.Fatalf("heartbeat daily stat = %#v, %v", stat, err)
	}
	for _, body := range []string{
		`{"position_seconds":1,"progress_percent":1,"watched_seconds_delta":1}`,
		`{"position_seconds":1,"progress_percent":1,"watched_seconds_delta":61,"report_id":"large"}`,
		`{"position_seconds":1,"progress_percent":1,"watched_seconds_delta":-1,"report_id":"negative"}`,
		`{"position_seconds":1,"progress_percent":1,"watched_seconds_delta":0,"report_id":"zero"}`,
		`{"position_seconds":1,"progress_percent":1,"watched_seconds_delta":1,"report_id":"missing-session"}`,
	} {
		response := requestJSON(t, router, http.MethodPost, "/lessons/"+lesson.ID+"/progress", body)
		if response.Code != http.StatusBadRequest {
			t.Errorf("invalid heartbeat status=%d body=%s", response.Code, response.Body.String())
		}
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
	admin.GET("/recent-learning", NewLearnerOverviewHandler(services.overview).Recent)
	if response := requestJSON(
		t, admin, http.MethodGet, "/recent-learning", "",
	); response.Code != http.StatusForbidden {
		t.Fatalf("admin status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestProgressHandlerHidesDraftCourseWithAndWithoutEnrollment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	admin := withRole("tenant_admin", tenant.ID, "admin-1")
	learner, err := services.users.Create(
		admin, "draft-learner@example.com", "password123", "Draft Learner", "learner",
	)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	course, err := services.courses.Create(admin, "Draft course", "", "")
	if err != nil {
		t.Fatalf("create draft course: %v", err)
	}
	chapter, err := services.chapters.Create(admin, course.ID, "Draft chapter", 1)
	if err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson, err := services.lessons.Create(
		admin, chapter.ID, "Draft lesson", "text", "body", 0, 1,
	)
	if err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	handler := NewProgressHandler(services.progress)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, learner.ID))
	router.GET("/lessons/:id/progress", handler.Get)
	router.POST("/lessons/:id/progress", handler.Report)

	requests := []struct {
		method, body string
	}{
		{http.MethodGet, ""},
		{http.MethodPost, `{"position_seconds":10,"progress_percent":10}`},
	}
	for _, request := range requests {
		response := requestJSON(t, router, request.method, "/lessons/"+lesson.ID+"/progress", request.body)
		if response.Code != http.StatusNotFound {
			t.Errorf("without enrollment %s status=%d body=%s", request.method, response.Code, response.Body.String())
		}
	}
	learnerContext := withRole("learner", tenant.ID, learner.ID)
	if _, err := services.enrollmentRepo.FindByCourseAndUser(learnerContext, course.ID, learner.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("unexpected enrollment after hidden requests: %v", err)
	}
	preassigned, err := services.enrollments.Enroll(admin, course.ID, learner.ID, domain.AssignmentRequired)
	if err != nil {
		t.Fatalf("preassign draft course: %v", err)
	}
	for _, request := range requests {
		response := requestJSON(t, router, request.method, "/lessons/"+lesson.ID+"/progress", request.body)
		if response.Code != http.StatusNotFound {
			t.Errorf("with enrollment %s status=%d body=%s", request.method, response.Code, response.Body.String())
		}
	}
	items, err := services.enrollmentRepo.FindByCourse(learnerContext, course.ID)
	if err != nil || len(items) != 1 || items[0].ID != preassigned.ID {
		t.Fatalf("draft enrollments = %#v, %v", items, err)
	}
}

func TestProgressHandlerEnforcesOfficialCourseAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	admin := withRole("tenant_admin", tenant.ID, "admin-1")
	superadmin := withRole("superadmin", "", "root-1")
	learner, err := services.users.Create(
		admin, "official-learner@example.com", "password123", "Official Learner", "learner",
	)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	handler := NewProgressHandler(services.progress)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, learner.ID))
	router.GET("/lessons/:id/progress", handler.Get)
	router.POST("/lessons/:id/progress", handler.Report)
	learnerContext := withRole("learner", tenant.ID, learner.ID)

	tests := []struct {
		name       string
		status     int
		activation *bool
		wantStatus int
	}{
		{name: "unpublished enabled", status: 0, activation: boolPointer(true), wantStatus: http.StatusNotFound},
		{name: "published disabled", status: 1, activation: boolPointer(false), wantStatus: http.StatusNotFound},
		{name: "published unactivated", status: 1, wantStatus: http.StatusNotFound},
		{name: "published enabled", status: 1, activation: boolPointer(true), wantStatus: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			course, err := services.courses.CreateOfficial(superadmin, test.name, "", "", test.status)
			if err != nil {
				t.Fatalf("create official course: %v", err)
			}
			chapter, err := services.chapters.Create(superadmin, course.ID, test.name+" chapter", 1)
			if err != nil {
				t.Fatalf("create official chapter: %v", err)
			}
			lesson, err := services.lessons.Create(superadmin, chapter.ID, test.name+" lesson", "text", "body", 0, 1)
			if err != nil {
				t.Fatalf("create official lesson: %v", err)
			}
			if test.activation != nil {
				if err := services.courses.EnableOfficial(admin, course.ID, *test.activation); err != nil {
					t.Fatalf("set official activation: %v", err)
				}
			}

			for _, request := range []struct {
				method, body string
			}{
				{http.MethodGet, ""},
				{http.MethodPost, `{"position_seconds":10,"progress_percent":10}`},
			} {
				response := requestJSON(t, router, request.method, "/lessons/"+lesson.ID+"/progress", request.body)
				if response.Code != test.wantStatus {
					t.Errorf("%s status=%d body=%s, want %d", request.method, response.Code, response.Body.String(), test.wantStatus)
				}
			}
			items, err := services.enrollmentRepo.FindByCourse(learnerContext, course.ID)
			wantEnrollments := 0
			if test.wantStatus == http.StatusOK {
				wantEnrollments = 1
			}
			if err != nil || len(items) != wantEnrollments {
				t.Fatalf("official enrollments = %#v, %v; want %d", items, err, wantEnrollments)
			}
		})
	}
}

func boolPointer(value bool) *bool { return &value }
