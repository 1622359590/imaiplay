package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/gin-gonic/gin"
)

func TestCourseMaterialHandlerManagementRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &courseMaterialServiceStub{}
	handler := NewCourseMaterialHandler(stub)
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-1", "admin"))
	router.GET("/courses/:id/materials", handler.List)
	router.POST("/courses/:id/materials", handler.Add)
	router.PUT("/courses/:id/materials/:materialID", handler.Update)
	router.DELETE("/courses/:id/materials/:materialID", handler.Remove)

	created := requestJSON(t, router, http.MethodPost, "/courses/course-1/materials", `{"resource_id":"resource-1","display_name":"入门.pdf","sort_order":2}`)
	if created.Code != http.StatusOK || stub.lastCourseID != "course-1" || stub.lastInput.ResourceID != "resource-1" || stub.lastInput.SortOrder != 2 {
		t.Fatalf("Add status=%d body=%s captured=%#v", created.Code, created.Body.String(), stub.lastInput)
	}
	updated := requestJSON(t, router, http.MethodPut, "/courses/course-1/materials/material-1", `{"resource_id":"resource-2","display_name":"更新.pdf","sort_order":1}`)
	if updated.Code != http.StatusOK || stub.lastMaterialID != "material-1" || stub.lastInput.DisplayName != "更新.pdf" {
		t.Fatalf("Update status=%d body=%s", updated.Code, updated.Body.String())
	}
	listed := requestJSON(t, router, http.MethodGet, "/courses/course-1/materials", "")
	if listed.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", listed.Code, listed.Body.String())
	}
	removed := requestJSON(t, router, http.MethodDelete, "/courses/course-1/materials/material-1", "")
	if removed.Code != http.StatusOK || stub.lastMaterialID != "material-1" {
		t.Fatalf("Remove status=%d body=%s", removed.Code, removed.Body.String())
	}
}

func TestCourseMaterialHandlerDownloadUsesSafeAttachmentHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &courseMaterialServiceStub{downloadBody: []byte("course guide"), downloadType: "application/pdf", downloadName: "../\"guide\r\n.pdf"}
	handler := NewCourseMaterialHandler(stub)
	router := gin.New()
	router.Use(asUser("learner", "tenant-1", "learner"))
	router.GET("/course-materials/:id/download", handler.Download)
	response := requestJSON(t, router, http.MethodGet, "/course-materials/material-1/download", "")
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), stub.downloadBody) {
		t.Fatalf("Download status=%d body=%q", response.Code, response.Body.Bytes())
	}
	disposition := response.Header().Get("Content-Disposition")
	if !strings.HasPrefix(disposition, "attachment;") || strings.ContainsAny(disposition, "\r\n") || strings.Contains(disposition, "../") {
		t.Fatalf("Content-Disposition = %q", disposition)
	}
}

func TestCourseMaterialHandlerRejectsInvalidBodyAndRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewCourseMaterialHandler(&courseMaterialServiceStub{})
	admin := gin.New()
	admin.Use(asUser("tenant_admin", "tenant-1", "admin"))
	admin.POST("/courses/:id/materials", handler.Add)
	if response := requestJSON(t, admin, http.MethodPost, "/courses/course-1/materials", `{}`); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid body status=%d body=%s", response.Code, response.Body.String())
	}
	learner := gin.New()
	learner.Use(asUser("learner", "tenant-1", "learner"))
	learner.GET("/courses/:id/materials", handler.List)
	if response := requestJSON(t, learner, http.MethodGet, "/courses/course-1/materials", ""); response.Code != http.StatusForbidden {
		t.Fatalf("learner status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCourseMaterialHandlerInstructorIsReadOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &courseMaterialServiceStub{}
	handler := NewCourseMaterialHandler(stub)
	router := gin.New()
	router.Use(asUser("instructor", "tenant-1", "instructor"))
	router.GET("/courses/:id/materials", handler.List)
	router.POST("/courses/:id/materials", handler.Add)
	router.PUT("/courses/:id/materials/:materialID", handler.Update)
	router.DELETE("/courses/:id/materials/:materialID", handler.Remove)
	if response := requestJSON(t, router, http.MethodGet, "/courses/course-1/materials", ""); response.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", response.Code, response.Body.String())
	}
	for _, request := range []struct{ method, path, body string }{
		{http.MethodPost, "/courses/course-1/materials", `{"resource_id":"resource","display_name":"guide.pdf"}`},
		{http.MethodPut, "/courses/course-1/materials/material", `{"resource_id":"resource","display_name":"guide.pdf"}`},
		{http.MethodDelete, "/courses/course-1/materials/material", ""},
	} {
		if response := requestJSON(t, router, request.method, request.path, request.body); response.Code != http.StatusForbidden {
			t.Fatalf("%s %s status=%d body=%s", request.method, request.path, response.Code, response.Body.String())
		}
	}
}

type courseMaterialServiceStub struct {
	lastCourseID, lastMaterialID string
	lastInput                    service.CourseMaterialInput
	downloadBody                 []byte
	downloadType, downloadName   string
}

func (stub *courseMaterialServiceStub) OpenForLearner(_ context.Context, materialID string) (io.ReadCloser, string, string, error) {
	stub.lastMaterialID = materialID
	return io.NopCloser(bytes.NewReader(stub.downloadBody)), stub.downloadType, stub.downloadName, nil
}

func (stub *courseMaterialServiceStub) Add(_ context.Context, courseID string, input service.CourseMaterialInput) (*domain.CourseMaterial, error) {
	stub.lastCourseID, stub.lastInput = courseID, input
	return &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: "material-1"}, CourseID: courseID, ResourceID: input.ResourceID, DisplayName: input.DisplayName, SortOrder: input.SortOrder}, nil
}

func (stub *courseMaterialServiceStub) Update(_ context.Context, courseID, materialID string, input service.CourseMaterialInput) (*domain.CourseMaterial, error) {
	stub.lastCourseID, stub.lastMaterialID, stub.lastInput = courseID, materialID, input
	return &domain.CourseMaterial{BaseModel: domain.BaseModel{ID: materialID}, CourseID: courseID, ResourceID: input.ResourceID, DisplayName: input.DisplayName, SortOrder: input.SortOrder}, nil
}

func (stub *courseMaterialServiceStub) Remove(_ context.Context, courseID, materialID string) error {
	stub.lastCourseID, stub.lastMaterialID = courseID, materialID
	return nil
}

func (stub *courseMaterialServiceStub) ListForManager(_ context.Context, courseID string) ([]domain.CourseMaterial, error) {
	stub.lastCourseID = courseID
	return []domain.CourseMaterial{}, nil
}
