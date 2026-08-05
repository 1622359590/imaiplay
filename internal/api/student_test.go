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
	categoryID := "category-1"
	ownerTenantID := "owner-tenant"
	course := &domain.Course{
		BaseModel: domain.BaseModel{ID: "safe-course", TenantID: tenant.ID},
		Title:     "Safe course", Description: "Safe description",
		CoverImage: "/uploads/private/cover.png", Status: 1, CreatedBy: "admin",
		OwnerTenantID: &ownerTenantID, CategoryID: &categoryID, Enabled: true,
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
		{BaseModel: domain.BaseModel{ID: "text-lesson", TenantID: tenant.ID}, ChapterID: chapter.ID, Title: "Text", ContentType: "text", ContentURL: "lesson body", SortOrder: 4},
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
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertExactJSONKeys(t, envelope.Data, "course", "chapters", "materials")
	courseJSON := requireJSONObject(t, envelope.Data["course"])
	assertExactJSONKeys(t, courseJSON, "id", "title", "description", "cover_image", "category_id", "is_official")
	if courseJSON["cover_image"] != "" || courseJSON["category_id"] != categoryID {
		t.Fatalf("learner course = %#v", courseJSON)
	}
	chapters := requireJSONArray(t, envelope.Data["chapters"])
	materials := requireJSONArray(t, envelope.Data["materials"])
	if len(chapters) != 1 || len(materials) != 1 {
		t.Fatalf("unexpected detail shape: %s", response.Body.String())
	}
	chapterJSON := requireJSONObject(t, chapters[0])
	assertExactJSONKeys(t, chapterJSON, "id", "title", "sort_order", "lessons")
	lessonsJSON := requireJSONArray(t, chapterJSON["lessons"])
	if len(lessonsJSON) != 4 {
		t.Fatalf("lessons = %#v", lessonsJSON)
	}
	storedLesson := requireJSONObject(t, lessonsJSON[0])
	assertExactJSONKeys(t, storedLesson, "id", "title", "content_type", "resource_id", "resource_type", "content_url", "duration_seconds", "sort_order")
	if storedLesson["content_url"] != "" || storedLesson["resource_id"] != stored.ID {
		t.Fatalf("stored lesson = %#v", storedLesson)
	}
	safeExternal := requireJSONObject(t, lessonsJSON[1])
	assertExactJSONKeys(t, safeExternal, "id", "title", "content_type", "content_url", "duration_seconds", "sort_order")
	if got := safeExternal["content_url"]; got != "https://cdn.example.com/public.mp4" {
		t.Fatalf("safe external content_url = %#v", got)
	}
	unsafeExternal := requireJSONObject(t, lessonsJSON[2])
	assertExactJSONKeys(t, unsafeExternal, "id", "title", "content_type", "content_url", "duration_seconds", "sort_order")
	if got := unsafeExternal["content_url"]; got != "" {
		t.Fatalf("unsafe external content_url = %#v", got)
	}
	textLesson := requireJSONObject(t, lessonsJSON[3])
	assertExactJSONKeys(t, textLesson, "id", "title", "content_type", "content_url", "duration_seconds", "sort_order")
	if got := textLesson["content_url"]; got != "lesson body" {
		t.Fatalf("text content_url = %#v", got)
	}
	material := requireJSONObject(t, materials[0])
	assertExactJSONKeys(t, material, "id", "display_name", "resource_type", "size_bytes")
	for _, forbidden := range []string{
		`"tenant_id"`, `"created_by"`, `"owner_tenant_id"`,
		`"created_at"`, `"updated_at"`, `"enabled"`,
		`/uploads/private`, `s3.amazonaws.com`, `aliyuncs.com`,
	} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("learner detail contains %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestStudentCourseDetailCoverAllowsOnlyPublicNonStorageURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewCourseHandler(services.courses)
	router := gin.New()
	router.Use(asUser("learner", tenant.ID, "learner-1"))
	router.GET("/api/v1/courses/:id", handler.PublishedDetail)

	for index, test := range []struct {
		name  string
		cover string
		want  string
	}{
		{"public CDN", "https://images.example.com/course.png", "https://images.example.com/course.png"},
		{"local storage", "/uploads/tenant/private.png", ""},
		{"S3 object", "https://bucket.s3.amazonaws.com/private/course.png", ""},
		{"OSS object", "https://bucket.oss-cn-shanghai.aliyuncs.com/private/course.png", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			id := "cover-course-" + string(rune('a'+index))
			course := &domain.Course{
				BaseModel: domain.BaseModel{ID: id, TenantID: tenant.ID},
				Title:     id, CoverImage: test.cover, Status: 1, CreatedBy: "admin",
			}
			if err := services.database.Create(course).Error; err != nil {
				t.Fatalf("create course: %v", err)
			}
			if err := services.database.Create(&domain.CourseEnrollment{
				BaseModel: domain.BaseModel{TenantID: tenant.ID}, CourseID: id,
				UserID: "learner-1", Status: 1, AssignmentType: domain.AssignmentRequired,
			}).Error; err != nil {
				t.Fatalf("create enrollment: %v", err)
			}
			response := requestJSON(t, router, http.MethodGet, "/api/v1/courses/"+id, "")
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			var envelope struct {
				Data struct {
					Course map[string]any `json:"course"`
				} `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got := envelope.Data.Course["cover_image"]; got != test.want {
				t.Fatalf("cover_image = %#v, want %q", got, test.want)
			}
		})
	}
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

func requireJSONObject(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value = %#v, want JSON object", value)
	}
	return result
}

func requireJSONArray(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("value = %#v, want JSON array", value)
	}
	return result
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
