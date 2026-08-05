package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
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
	for _, courseID := range []string{publishedID, draftID} {
		if err := services.database.Create(&domain.CourseEnrollment{
			BaseModel: domain.BaseModel{TenantID: tenant.ID},
			CourseID:  courseID, UserID: "learner-1", Status: 1,
			AssignmentType: domain.AssignmentRequired,
		}).Error; err != nil {
			t.Fatalf("assign course: %v", err)
		}
	}

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

func TestStudentCourseDetailReturnsSafeLessonAndMaterialDTOs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	course := &domain.Course{
		BaseModel: domain.BaseModel{ID: "safe-course", TenantID: tenant.ID},
		Title:     "Safe course", Status: 1, CreatedBy: "admin",
	}
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{ID: "safe-chapter", TenantID: tenant.ID},
		CourseID:  course.ID, Title: "Chapter", SortOrder: 1,
	}
	stored := &domain.Resource{
		BaseModel: domain.BaseModel{ID: "stored-video", TenantID: tenant.ID},
		Name:      "video.mp4", ResourceType: "video", URL: "/uploads/private/video.mp4", CreatedBy: "admin",
	}
	attachment := &domain.Resource{
		BaseModel: domain.BaseModel{ID: "stored-attachment", TenantID: tenant.ID},
		Name:      "guide.pdf", ResourceType: "attachment", URL: "/uploads/private/guide.pdf", SizeBytes: 42, CreatedBy: "admin",
	}
	for _, value := range []any{course, chapter, stored, attachment} {
		if err := services.database.Create(value).Error; err != nil {
			t.Fatalf("create %T: %v", value, err)
		}
	}
	lessons := []*domain.CourseLesson{
		{BaseModel: domain.BaseModel{ID: "stored-lesson", TenantID: tenant.ID}, ChapterID: chapter.ID, Title: "Stored", ContentType: "video", ResourceID: &stored.ID, ContentURL: stored.URL, DurationSeconds: 60, SortOrder: 1},
		{BaseModel: domain.BaseModel{ID: "safe-external", TenantID: tenant.ID}, ChapterID: chapter.ID, Title: "External", ContentType: "video", ContentURL: "https://cdn.example.com/public.mp4", DurationSeconds: 30, SortOrder: 2},
		{BaseModel: domain.BaseModel{ID: "unsafe-external", TenantID: tenant.ID}, ChapterID: chapter.ID, Title: "Unsafe", ContentType: "video", ContentURL: "javascript:alert(1)", DurationSeconds: 30, SortOrder: 3},
	}
	for _, lesson := range lessons {
		if err := services.database.Create(lesson).Error; err != nil {
			t.Fatalf("create lesson: %v", err)
		}
	}
	if err := services.database.Create(&domain.CourseMaterial{
		BaseModel: domain.BaseModel{ID: "material-1", TenantID: tenant.ID},
		CourseID:  course.ID, ResourceID: attachment.ID, DisplayName: "Guide.pdf", CreatedBy: "admin",
	}).Error; err != nil {
		t.Fatalf("create material: %v", err)
	}
	if err := services.database.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: tenant.ID}, CourseID: course.ID,
		UserID: "learner-1", Status: 1, AssignmentType: domain.AssignmentRequired,
	}).Error; err != nil {
		t.Fatalf("create enrollment: %v", err)
	}

	handler := NewCourseHandler(services.courses)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, "learner-1"))
	router.GET("/api/v1/courses/:id", handler.PublishedDetail)
	response := requestJSON(t, router, http.MethodGet, "/api/v1/courses/"+course.ID, "")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data struct {
			Chapters []struct {
				Lessons []map[string]any `json:"lessons"`
			} `json:"chapters"`
			Materials []map[string]any `json:"materials"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data.Chapters) != 1 || len(envelope.Data.Chapters[0].Lessons) != 3 || len(envelope.Data.Materials) != 1 {
		t.Fatalf("unexpected detail shape: %s", response.Body.String())
	}
	storedLesson := envelope.Data.Chapters[0].Lessons[0]
	assertExactJSONKeys(t, storedLesson, "id", "title", "content_type", "resource_id", "resource_type", "content_url", "duration_seconds", "sort_order")
	if storedLesson["content_url"] != "" || storedLesson["resource_id"] != stored.ID {
		t.Fatalf("stored lesson = %#v", storedLesson)
	}
	if got := envelope.Data.Chapters[0].Lessons[1]["content_url"]; got != "https://cdn.example.com/public.mp4" {
		t.Fatalf("safe external content_url = %#v", got)
	}
	if got := envelope.Data.Chapters[0].Lessons[2]["content_url"]; got != "" {
		t.Fatalf("unsafe external content_url = %#v", got)
	}
	material := envelope.Data.Materials[0]
	assertExactJSONKeys(t, material, "id", "display_name", "resource_type", "size_bytes")
}

func assertExactJSONKeys(t *testing.T, value map[string]any, keys ...string) {
	t.Helper()
	if len(value) != len(keys) {
		t.Fatalf("JSON keys = %#v, want %v", value, keys)
	}
	for _, key := range keys {
		if _, ok := value[key]; !ok {
			t.Fatalf("JSON keys = %#v, missing %q", value, key)
		}
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
