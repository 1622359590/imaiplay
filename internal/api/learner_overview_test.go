package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestLearnerOverviewHandlerReturnsExactSafeFieldsAndRecentCourseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	admin := withRole("tenant_admin", tenant.ID, "admin-1")
	learner, err := services.users.Create(
		admin, "overview@example.com", "password123", "Overview Learner", "learner",
	)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	category := &domain.CourseCategory{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Name:      "Sales", NormalizedName: "sales", Status: 1,
	}
	if err := services.database.Create(category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}
	course, err := services.courses.Create(admin, "Course", "Description", "cover.png")
	if err != nil {
		t.Fatalf("create course: %v", err)
	}
	if err := services.database.Model(course).Update("category_id", category.ID).Error; err != nil {
		t.Fatalf("set category: %v", err)
	}
	course, err = services.courses.Update(admin, course.ID, course.Title, course.Description, course.CoverImage, 1)
	if err != nil {
		t.Fatalf("publish course: %v", err)
	}
	chapter, err := services.chapters.Create(admin, course.ID, "Chapter", 1)
	if err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson, err := services.lessons.Create(admin, chapter.ID, "Lesson", "video", "/video.mp4", 100, 1)
	if err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	if _, err := services.enrollments.Enroll(admin, course.ID, learner.ID, domain.AssignmentRequired); err != nil {
		t.Fatalf("enroll: %v", err)
	}
	learnerCtx := withRole("learner", tenant.ID, learner.ID)
	if _, err := services.progress.Report(learnerCtx, lesson.ID, 40, 40, 0, ""); err != nil {
		t.Fatalf("report progress: %v", err)
	}

	handler := NewLearnerOverviewHandler(services.overview)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, learner.ID))
	router.GET("/learner/overview", handler.Get)
	router.GET("/recent-learning", handler.Recent)

	response := requestJSON(t, router, http.MethodGet, "/learner/overview", "")
	if response.Code != http.StatusOK {
		t.Fatalf("overview status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	assertJSONKeys(t, envelope.Data,
		"categories", "courses", "required_completed", "required_total",
		"today_learning_seconds", "total_learning_seconds")
	var courses []map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Data["courses"], &courses); err != nil || len(courses) != 1 {
		t.Fatalf("decode courses = %#v, %v", courses, err)
	}
	assertJSONKeys(t, courses[0],
		"assignment_type", "completed_lesson_count", "course", "last_learned_at",
		"lesson_count", "progress_percent", "recent_lesson")
	var courseJSON map[string]json.RawMessage
	if err := json.Unmarshal(courses[0]["course"], &courseJSON); err != nil {
		t.Fatalf("decode course: %v", err)
	}
	assertJSONKeys(t, courseJSON, "category", "cover_image", "description", "id", "title")
	var lessonJSON map[string]json.RawMessage
	if err := json.Unmarshal(courses[0]["recent_lesson"], &lessonJSON); err != nil {
		t.Fatalf("decode recent lesson: %v", err)
	}
	assertJSONKeys(t, lessonJSON, "duration_seconds", "id", "last_position_seconds", "title")

	recent := requestJSON(t, router, http.MethodGet, "/recent-learning", "")
	if recent.Code != http.StatusOK {
		t.Fatalf("recent status=%d body=%s", recent.Code, recent.Body.String())
	}
	var recentEnvelope struct {
		Data struct {
			Items []map[string]json.RawMessage `json:"items"`
			Total int                          `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recent.Body.Bytes(), &recentEnvelope); err != nil ||
		recentEnvelope.Data.Total != 1 || len(recentEnvelope.Data.Items) != 1 {
		t.Fatalf("decode recent = %#v, %v", recentEnvelope, err)
	}
	assertJSONKeys(t, recentEnvelope.Data.Items[0],
		"course", "last_learned_at", "last_position_seconds", "progress_percent",
		"recent_lesson")
}

func TestLearnerOverviewHandlerRequiresLearnerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewLearnerOverviewHandler(services.overview)
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	router.GET("/learner/overview", handler.Get)
	if response := requestJSON(t, router, http.MethodGet, "/learner/overview", ""); response.Code != http.StatusForbidden {
		t.Fatalf("admin overview status=%d body=%s", response.Code, response.Body.String())
	}
}

func assertJSONKeys(t *testing.T, value map[string]json.RawMessage, expected ...string) {
	t.Helper()
	actual := make([]string, 0, len(value))
	for key := range value {
		actual = append(actual, key)
	}
	sort.Strings(actual)
	sort.Strings(expected)
	if len(actual) != len(expected) {
		t.Fatalf("JSON keys = %v, want %v", actual, expected)
	}
	for index := range actual {
		if actual[index] != expected[index] {
			t.Fatalf("JSON keys = %v, want %v", actual, expected)
		}
	}
}
