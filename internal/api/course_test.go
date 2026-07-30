package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCourseHandlersCRUDAndDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	router := courseTestRouter(services, "tenant_admin", tenant.ID)

	created := requestJSON(t, router, http.MethodPost, "/courses",
		`{"title":"Go Basics","description":"intro","cover_image":"cover.png"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	courseID := responseID(t, created.Body.Bytes())

	chapter := requestJSON(t, router, http.MethodPost,
		"/courses/"+courseID+"/chapters", `{"title":"Start","sort_order":1}`)
	if chapter.Code != http.StatusOK {
		t.Fatalf("chapter status=%d body=%s", chapter.Code, chapter.Body.String())
	}
	chapterID := responseID(t, chapter.Body.Bytes())
	video := []byte{
		0, 0, 0, 24, 'f', 't', 'y', 'p',
		'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm',
	}
	resource, err := services.resources.Upload(
		withRole("tenant_admin", tenant.ID, "user-1"),
		"lesson.mp4", bytes.NewReader(video), int64(len(video)),
	)
	if err != nil {
		t.Fatalf("upload lesson resource: %v", err)
	}

	lesson := requestJSON(t, router, http.MethodPost,
		"/chapters/"+chapterID+"/lessons",
		fmt.Sprintf(
			`{"title":"Install","content_type":"video","resource_id":%q,"duration_seconds":60}`,
			resource.ID,
		))
	if lesson.Code != http.StatusOK {
		t.Fatalf("lesson status=%d body=%s", lesson.Code, lesson.Body.String())
	}

	detail := requestJSON(t, router, http.MethodGet, "/courses/"+courseID+"/detail", "")
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}
	var body struct {
		Data struct {
			Chapters []struct {
				Lessons []struct {
					Title string `json:"title"`
				} `json:"lessons"`
			} `json:"chapters"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if len(body.Data.Chapters) != 1 ||
		len(body.Data.Chapters[0].Lessons) != 1 ||
		body.Data.Chapters[0].Lessons[0].Title != "Install" {
		t.Fatalf("detail body=%s", detail.Body.String())
	}
	var lessonData struct {
		Data struct {
			Chapters []struct {
				Lessons []struct {
					ResourceID string `json:"resource_id"`
				} `json:"lessons"`
			} `json:"chapters"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &lessonData); err != nil ||
		lessonData.Data.Chapters[0].Lessons[0].ResourceID != resource.ID {
		t.Fatalf("resource id missing from detail: %s", detail.Body.String())
	}

	if response := requestJSON(t, router, http.MethodPut, "/courses/"+courseID,
		`{"title":"Go Advanced","description":"updated","cover_image":"","status":1}`); response.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", response.Code, response.Body.String())
	}
	list := requestJSON(t, router, http.MethodGet, "/courses", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if response := requestJSON(t, router, http.MethodDelete, "/courses/"+courseID, ""); response.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCourseHandlersRejectLearnerAndOtherInstructor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	admin := courseTestRouter(services, "tenant_admin", tenant.ID)
	created := requestJSON(t, admin, http.MethodPost, "/courses",
		`{"title":"Owned","description":"","cover_image":""}`)
	courseID := responseID(t, created.Body.Bytes())

	learner := courseTestRouter(services, "learner", tenant.ID)
	if response := requestJSON(t, learner, http.MethodGet, "/courses", ""); response.Code != http.StatusForbidden {
		t.Fatalf("learner status=%d body=%s", response.Code, response.Body.String())
	}
	otherInstructor := courseTestRouterWithUser(
		services, "instructor", tenant.ID, "other-instructor",
	)
	if response := requestJSON(t, otherInstructor, http.MethodGet,
		"/courses/"+courseID, ""); response.Code != http.StatusNotFound {
		t.Fatalf("other instructor status=%d body=%s",
			response.Code, response.Body.String())
	}
}

func courseTestRouter(
	services testServices, role, tenantID string,
) *gin.Engine {
	return courseTestRouterWithUser(services, role, tenantID, "user-1")
}

func courseTestRouterWithUser(
	services testServices, role, tenantID, userID string,
) *gin.Engine {
	router := gin.New()
	router.Use(asUser(role, tenantID, userID))
	courses := NewCourseHandler(services.courses)
	chapters := NewCourseChapterHandler(services.chapters)
	lessons := NewCourseLessonHandler(services.lessons)
	router.POST("/courses", courses.Create)
	router.POST("/official-courses", courses.CreateOfficial)
	router.GET("/courses", courses.List)
	router.GET("/courses/:id", courses.Get)
	router.PUT("/courses/:id", courses.Update)
	router.DELETE("/courses/:id", courses.Delete)
	router.GET("/courses/:id/detail", courses.Detail)
	router.POST("/courses/:id/chapters", chapters.Create)
	router.GET("/courses/:id/chapters", chapters.List)
	router.PUT("/chapters/:id", chapters.Update)
	router.DELETE("/chapters/:id", chapters.Delete)
	router.POST("/chapters/:id/lessons", lessons.Create)
	router.GET("/chapters/:id/lessons", lessons.List)
	router.PUT("/lessons/:id", lessons.Update)
	router.DELETE("/lessons/:id", lessons.Delete)
	return router
}

func responseID(t *testing.T, data []byte) string {
	t.Helper()
	var response struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ID == "" {
		t.Fatalf("missing id in %s", data)
	}
	return response.Data.ID
}
