package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuperadminManagesOfficialCourseContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	router := courseTestRouter(services, "superadmin", "")
	video := []byte{
		0, 0, 0, 24, 'f', 't', 'y', 'p',
		'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm',
	}
	resource, err := services.resources.UploadPlatform(
		withRole("superadmin", "", "user-1"),
		"lesson.mp4", bytes.NewReader(video), int64(len(video)),
	)
	if err != nil {
		t.Fatalf("upload platform resource: %v", err)
	}

	created := requestJSON(t, router, http.MethodPost, "/official-courses",
		`{"title":"Official Go","description":"intro","status":0}`)
	if created.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var body struct {
		Data struct {
			ID         string `json:"id"`
			IsOfficial bool   `json:"is_official"`
			TenantID   string `json:"tenant_id"`
			Status     int    `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if body.Data.ID == "" || !body.Data.IsOfficial ||
		body.Data.TenantID != "" || body.Data.Status != 0 {
		t.Fatalf("unexpected official course: %s", created.Body.String())
	}

	chapter := requestJSON(t, router, http.MethodPost,
		"/courses/"+body.Data.ID+"/chapters", `{"title":"Start","sort_order":1}`)
	if chapter.Code != http.StatusOK {
		t.Fatalf("chapter status=%d body=%s", chapter.Code, chapter.Body.String())
	}
	chapterID := responseID(t, chapter.Body.Bytes())
	lesson := requestJSON(t, router, http.MethodPost,
		"/chapters/"+chapterID+"/lessons",
		fmt.Sprintf(
			`{"title":"Install","content_type":"video","resource_id":%q}`,
			resource.ID,
		))
	if lesson.Code != http.StatusOK {
		t.Fatalf("lesson status=%d body=%s", lesson.Code, lesson.Body.String())
	}
	if response := requestJSON(t, router, http.MethodPut, "/courses/"+body.Data.ID,
		`{"title":"Official Go 2","description":"updated","status":1}`); response.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", response.Code, response.Body.String())
	}
	detail := requestJSON(t, router, http.MethodGet, "/courses/"+body.Data.ID+"/detail", "")
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}
	if response := requestJSON(t, router, http.MethodDelete, "/courses/"+body.Data.ID, ""); response.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", response.Code, response.Body.String())
	}
}
