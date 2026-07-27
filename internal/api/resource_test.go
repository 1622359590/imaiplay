package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResourceHandlerUploadListAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewResourceHandler(services.resources)
	router := gin.New()
	router.Use(asUser("tenant_admin", tenant.ID, "admin-1"))
	router.POST("/resources/upload", handler.Upload)
	router.GET("/resources", handler.List)
	router.DELETE("/resources/:id", handler.Delete)

	uploaded := requestMultipart(
		t, router, "/resources/upload", "guide.pdf", []byte("%PDF-1.7\n"),
	)
	if uploaded.Code != http.StatusOK {
		t.Fatalf("Upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
	}
	resourceID := responseID(t, uploaded.Body.Bytes())
	list := requestJSON(t, router, http.MethodGet, "/resources", "")
	if list.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", list.Code, list.Body.String())
	}
	var body struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &body); err != nil ||
		body.Data.Total != 1 {
		t.Fatalf("List body=%s error=%v", list.Body.String(), err)
	}
	if response := requestJSON(t, router, http.MethodDelete,
		"/resources/"+resourceID, ""); response.Code != http.StatusOK {
		t.Fatalf("Delete status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestResourceHandlerRejectsUnsupportedFileAndRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewResourceHandler(services.resources)
	admin := gin.New()
	admin.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	admin.POST("/resources/upload", handler.Upload)
	if response := requestMultipart(
		t, admin, "/resources/upload", "malware.exe", []byte("MZ executable"),
	); response.Code != http.StatusBadRequest {
		t.Fatalf("unsupported status=%d body=%s",
			response.Code, response.Body.String())
	}
	learner := gin.New()
	learner.Use(asUser("learner", "tenant-1", "learner-1"))
	learner.GET("/resources", handler.List)
	if response := requestJSON(
		t, learner, http.MethodGet, "/resources", "",
	); response.Code != http.StatusForbidden {
		t.Fatalf("learner status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestResourceHandlerRejectsOversizedRequestBeforeParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewResourceHandler(services.resources)
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	router.POST("/resources/upload", handler.Upload)
	request := httptest.NewRequest(
		http.MethodPost, "/resources/upload", bytes.NewReader(nil),
	)
	request.Header.Set("Content-Type", "multipart/form-data; boundary=test")
	request.ContentLength = maxResourceRequestSize + 1
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest ||
		!bytes.Contains(response.Body.Bytes(), []byte(
			"unsupported file type or size exceeds limit",
		)) {
		t.Fatalf("oversized status=%d body=%s",
			response.Code, response.Body.String())
	}
}

func requestMultipart(
	t *testing.T, router http.Handler, path, name string, data []byte,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write multipart data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
