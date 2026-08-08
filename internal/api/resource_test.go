package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
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

func TestResourceHandlerPersistsUploadedVideoDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewResourceHandler(services.resources)
	router := gin.New()
	router.Use(asUser("tenant_admin", tenant.ID, "admin-1"))
	router.POST("/resources/upload", handler.Upload)
	video := []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm'}

	uploaded := requestMultipartFields(
		t, router, "/resources/upload", "lesson.mp4", video,
		map[string]string{"duration_seconds": "73"},
	)
	if uploaded.Code != http.StatusOK ||
		!bytes.Contains(uploaded.Body.Bytes(), []byte(`"duration_seconds":73`)) {
		t.Fatalf("Upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
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
			"文件类型不支持或文件大小超过限制",
		)) {
		t.Fatalf("oversized status=%d body=%s",
			response.Code, response.Body.String())
	}
}

func TestResourceHandlerUploadsTenantAndPlatformAttachments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewResourceHandler(services.resources)
	tenant := gin.New()
	tenant.Use(asUser("tenant_admin", "tenant-1", "admin"))
	tenant.POST("/resources/attachments/upload", handler.UploadAttachment)
	platform := gin.New()
	platform.Use(asUser("superadmin", "", "root"))
	platform.POST("/admin/resources/attachments/upload", handler.UploadPlatformAttachment)
	body := []byte("%PDF-1.7\n")
	for name, router := range map[string]*gin.Engine{"tenant": tenant, "platform": platform} {
		path := "/resources/attachments/upload"
		if name == "platform" {
			path = "/admin/resources/attachments/upload"
		}
		response := requestMultipart(t, router, path, "guide.pdf", body)
		if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"resource_type":"attachment"`)) {
			t.Fatalf("%s attachment status=%d body=%s", name, response.Code, response.Body.String())
		}
	}
}

func TestResourceHandlerInstructorCanUploadAndListButCannotDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, tenantRepo := newTestServices(t)
	tenant := createTenant(t, tenantRepo)
	handler := NewResourceHandler(services.resources)
	router := gin.New()
	router.Use(asUser("instructor", tenant.ID, "instructor-1"))
	router.POST("/resources/attachments/upload", handler.UploadAttachment)
	router.GET("/resources", handler.List)
	router.DELETE("/resources/:id", handler.Delete)
	uploaded := requestMultipart(t, router, "/resources/attachments/upload", "guide.pdf", []byte("%PDF-1.7\n"))
	if uploaded.Code != http.StatusOK {
		t.Fatalf("UploadAttachment status=%d body=%s", uploaded.Code, uploaded.Body.String())
	}
	resourceID := responseID(t, uploaded.Body.Bytes())
	if listed := requestJSON(t, router, http.MethodGet, "/resources", ""); listed.Code != http.StatusOK {
		t.Fatalf("List status=%d body=%s", listed.Code, listed.Body.String())
	}
	if deleted := requestJSON(t, router, http.MethodDelete, "/resources/"+resourceID, ""); deleted.Code != http.StatusForbidden {
		t.Fatalf("Delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
}

func TestResourceRequestLimitAllowsOneGiBPlusMultipartOverhead(t *testing.T) {
	const (
		oneGiB            int64 = 1024 * 1024 * 1024
		multipartOverhead int64 = 1024 * 1024
	)
	want := oneGiB + multipartOverhead
	if maxResourceRequestSize != want {
		t.Fatalf(
			"maxResourceRequestSize = %d, want %d",
			maxResourceRequestSize,
			want,
		)
	}
}

func TestResourceHandlerPlatformUploadListCoverAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewResourceHandler(services.resources)
	router := gin.New()
	router.Use(asUser("superadmin", "", "root"))
	router.POST("/admin/resources/upload", handler.UploadPlatform)
	router.GET("/admin/resources", handler.ListPlatform)
	router.DELETE("/admin/resources/:id", handler.DeletePlatform)
	router.GET("/platform-covers/:id", handler.PlatformCover)

	png := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		make([]byte, 8)...,
	)
	uploaded := requestMultipart(
		t, router, "/admin/resources/upload", "cover.png", png,
	)
	if uploaded.Code != http.StatusOK {
		t.Fatalf(
			"UploadPlatform status=%d body=%s",
			uploaded.Code, uploaded.Body.String(),
		)
	}
	resourceID := responseID(t, uploaded.Body.Bytes())
	list := requestJSON(t, router, http.MethodGet, "/admin/resources", "")
	if list.Code != http.StatusOK ||
		!bytes.Contains(list.Body.Bytes(), []byte(resourceID)) {
		t.Fatalf("ListPlatform status=%d body=%s", list.Code, list.Body.String())
	}
	cover := requestJSON(
		t, router, http.MethodGet, "/platform-covers/"+resourceID, "",
	)
	if cover.Code != http.StatusOK || !bytes.Equal(cover.Body.Bytes(), png) ||
		cover.Header().Get("Content-Type") != "image/png" {
		t.Fatalf(
			"PlatformCover status=%d type=%q body=%q",
			cover.Code, cover.Header().Get("Content-Type"), cover.Body.Bytes(),
		)
	}
	deleted := requestJSON(
		t, router, http.MethodDelete, "/admin/resources/"+resourceID, "",
	)
	if deleted.Code != http.StatusOK {
		t.Fatalf(
			"DeletePlatform status=%d body=%s",
			deleted.Code, deleted.Body.String(),
		)
	}
}

func TestResourceHandlerPlatformMutationRequiresSuperadmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	services, _ := newTestServices(t)
	handler := NewResourceHandler(services.resources)
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-a", "admin"))
	router.POST("/admin/resources/upload", handler.UploadPlatform)
	router.GET("/admin/resources", handler.ListPlatform)
	router.DELETE("/admin/resources/:id", handler.DeletePlatform)

	png := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		make([]byte, 8)...,
	)
	if response := requestMultipart(
		t, router, "/admin/resources/upload", "cover.png", png,
	); response.Code != http.StatusForbidden {
		t.Fatalf(
			"UploadPlatform tenant status=%d body=%s",
			response.Code, response.Body.String(),
		)
	}
	if response := requestJSON(
		t, router, http.MethodGet, "/admin/resources", "",
	); response.Code != http.StatusForbidden {
		t.Fatalf(
			"ListPlatform tenant status=%d body=%s",
			response.Code, response.Body.String(),
		)
	}
}

func TestResourceHandlerFileReturnsProtectedContent(t *testing.T) {
	root := t.TempDir()
	path := root + "/tenant-1/document.pdf"
	content := []byte("%PDF-1.7\nprotected")
	if err := os.MkdirAll(root+"/tenant-1", 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	handler := NewResourceHandler(resourceFileStub{
		path: path, contentType: "application/pdf", fileName: "document.pdf",
	}, root)
	router := gin.New()
	router.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	router.GET("/resources/:id/file", handler.File)
	request := httptest.NewRequest(http.MethodGet, "/resources/resource-1/file", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), content) {
		t.Fatalf("file status=%d body=%q", response.Code, response.Body.Bytes())
	}
	if got := response.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("content type=%q", got)
	}
	if got := response.Header().Get("Content-Disposition"); got != `inline; filename="document.pdf"` {
		t.Fatalf("content disposition=%q", got)
	}
}

func TestResourceHandlerStreamsByteRangesWithoutBufferingWholeVideo(t *testing.T) {
	content := []byte("0123456789")
	handler := NewResourceHandler(resourceStreamStub{
		resourceFileStub: resourceFileStub{
			contentType: "video/mp4", fileName: "lesson.mp4",
		},
		content: content,
	}).WithLearnerAccess(learnerAccessStub{courseID: "course-1"})
	router := gin.New()
	router.Use(asUser("learner", "tenant-1", "learner-1"))
	router.GET("/resources/:id/file", handler.File)
	request := httptest.NewRequest(http.MethodGet, "/resources/resource-1/file", nil)
	request.Header.Set("Range", "bytes=2-5")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusPartialContent {
		t.Fatalf("range status=%d, want %d", response.Code, http.StatusPartialContent)
	}
	if got := response.Body.String(); got != "2345" {
		t.Fatalf("range body=%q, want %q", got, "2345")
	}
	if got := response.Header().Get("Content-Range"); got != "bytes 2-5/10" {
		t.Fatalf("content range=%q", got)
	}
}

func TestResourceHandlerIssuesShortLivedPlaybackURLAndStreamsIt(t *testing.T) {
	content := []byte("0123456789")
	handler := NewResourceHandler(resourceStreamStub{
		resourceFileStub: resourceFileStub{
			contentType: "video/mp4", fileName: "lesson.mp4",
		},
		content: content,
	}).WithLearnerAccess(learnerAccessStub{courseID: "course-1"}).WithPlaybackSecret("secret")
	router := gin.New()
	protected := router.Group("")
	protected.Use(asUser("learner", "tenant-1", "learner-1"))
	protected.GET("/resources/:id/playback-url", handler.PlaybackURL)
	router.GET("/resource-playback/:id", handler.Playback)

	issued := requestJSON(
		t, router, http.MethodGet, "/resources/resource-1/playback-url", "",
	)
	if issued.Code != http.StatusOK {
		t.Fatalf("playback URL status=%d body=%s", issued.Code, issued.Body.String())
	}
	var envelope struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(issued.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode playback URL: %v", err)
	}
	parsed, err := url.Parse(envelope.Data.URL)
	if err != nil || parsed.Query().Get("ticket") == "" {
		t.Fatalf("playback URL=%q error=%v", envelope.Data.URL, err)
	}
	claims, err := security.ValidatePlaybackToken(parsed.Query().Get("ticket"), "secret")
	if err != nil || claims.CourseID != "course-1" || claims.ResourceID != "resource-1" {
		t.Fatalf("playback claims=%#v error=%v", claims, err)
	}
	playbackPath := strings.Replace(
		envelope.Data.URL, "/api/v1/resource-playback/", "/resource-playback/", 1,
	)
	request := httptest.NewRequest(http.MethodGet, playbackPath, nil)
	request.Header.Set("Range", "bytes=4-7")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusPartialContent || response.Body.String() != "4567" {
		t.Fatalf("playback status=%d body=%q", response.Code, response.Body.String())
	}

	invalid := requestJSON(
		t, router, http.MethodGet,
		"/resource-playback/resource-1?ticket=invalid", "",
	)
	if invalid.Code != http.StatusUnauthorized {
		t.Fatalf("invalid ticket status=%d", invalid.Code)
	}
}

func TestResourceHandlerReauthorizesPlaybackIdentityAndCourse(t *testing.T) {
	stream := resourceStreamStub{
		resourceFileStub: resourceFileStub{contentType: "video/mp4", fileName: "lesson.mp4"},
		content:          []byte("video"),
	}
	handler := NewResourceHandler(stream).
		WithLearnerAccess(learnerAccessStub{
			courseID: "course-1", userID: "learner-1", tenantID: "tenant-1", role: "learner",
		}).WithPlaybackSecret("secret")
	router := gin.New()
	router.GET("/resource-playback/:id", handler.Playback)

	for _, test := range []struct {
		name       string
		resourceID string
		courseID   string
		userID     string
		tenantID   string
		role       string
		want       int
	}{
		{"valid", "resource-1", "course-1", "learner-1", "tenant-1", "learner", http.StatusOK},
		{"course mismatch", "resource-1", "course-2", "learner-1", "tenant-1", "learner", http.StatusNotFound},
		{"resource mismatch", "resource-2", "course-1", "learner-1", "tenant-1", "learner", http.StatusUnauthorized},
		{"user mismatch", "resource-1", "course-1", "learner-2", "tenant-1", "learner", http.StatusNotFound},
		{"tenant mismatch", "resource-1", "course-1", "learner-1", "tenant-2", "learner", http.StatusNotFound},
		{"role mismatch", "resource-1", "course-1", "learner-1", "tenant-1", "tenant_admin", http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			ticket, err := security.GeneratePlaybackToken(
				test.resourceID, test.courseID, test.userID, test.tenantID,
				"learner@example.com", test.role, "secret", time.Minute,
			)
			if err != nil {
				t.Fatalf("GeneratePlaybackToken() error = %v", err)
			}
			response := requestJSON(t, router, http.MethodGet,
				"/resource-playback/resource-1?ticket="+url.QueryEscape(ticket), "")
			if response.Code != test.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestResourceHandlerPlaybackPreservesAuthorizationDatabaseFailure(t *testing.T) {
	handler := NewResourceHandler(resourceStreamStub{
		resourceFileStub: resourceFileStub{contentType: "video/mp4", fileName: "lesson.mp4"},
		content:          []byte("video"),
	}).WithLearnerAccess(learnerAccessStub{err: errorsx.Internal("database unavailable")}).
		WithPlaybackSecret("secret")
	router := gin.New()
	router.GET("/resource-playback/:id", handler.Playback)
	ticket, err := security.GeneratePlaybackToken(
		"resource-1", "course-1", "learner-1", "tenant-1",
		"learner@example.com", "learner", "secret", time.Minute,
	)
	if err != nil {
		t.Fatalf("GeneratePlaybackToken() error = %v", err)
	}
	response := requestJSON(t, router, http.MethodGet,
		"/resource-playback/resource-1?ticket="+url.QueryEscape(ticket), "")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestResourceHandlerLearnerFileAuthorizesButManagerPreviewKeepsRoleScope(t *testing.T) {
	stream := resourceStreamStub{
		resourceFileStub: resourceFileStub{contentType: "video/mp4", fileName: "lesson.mp4"},
		content:          []byte("video"),
	}
	handler := NewResourceHandler(stream).WithLearnerAccess(learnerAccessStub{deny: true})

	learnerRouter := gin.New()
	learnerRouter.Use(asUser("learner", "tenant-1", "learner-1"))
	learnerRouter.GET("/resources/:id/file", handler.File)
	if response := requestJSON(t, learnerRouter, http.MethodGet, "/resources/resource-1/file", ""); response.Code != http.StatusNotFound {
		t.Fatalf("learner file status=%d body=%s", response.Code, response.Body.String())
	}

	managerRouter := gin.New()
	managerRouter.Use(asUser("tenant_admin", "tenant-1", "admin-1"))
	managerRouter.GET("/resources/:id/file", handler.File)
	if response := requestJSON(t, managerRouter, http.MethodGet, "/resources/resource-1/file", ""); response.Code != http.StatusOK || response.Body.String() != "video" {
		t.Fatalf("manager preview status=%d body=%q", response.Code, response.Body.String())
	}
}

type learnerAccessStub struct {
	courseID, userID, tenantID, role string
	deny                             bool
	err                              error
}

func (stub learnerAccessStub) AuthorizeLessonResource(ctx context.Context, _ string) (*domain.Course, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if stub.deny || !ok || (stub.userID != "" && stub.userID != userID) ||
		(stub.tenantID != "" && stub.tenantID != tenantID) || (stub.role != "" && stub.role != role) {
		return nil, errorsx.NotFound("resource not found")
	}
	return &domain.Course{BaseModel: domain.BaseModel{ID: stub.courseID}}, nil
}

type resourceStreamStub struct {
	resourceFileStub
	content []byte
}

func (stub resourceStreamStub) Open(context.Context, string) (io.ReadCloser, string, string, error) {
	return &readSeekCloser{Reader: bytes.NewReader(stub.content)}, stub.contentType, stub.fileName, nil
}

type readSeekCloser struct{ *bytes.Reader }

func (*readSeekCloser) Close() error { return nil }

type resourceFileStub struct {
	path, contentType, fileName string
}

func (resourceFileStub) Upload(context.Context, string, io.Reader, int64) (*domain.Resource, error) {
	return nil, nil
}

func (resourceFileStub) UploadPlatform(context.Context, string, io.Reader, int64) (*domain.Resource, error) {
	return nil, nil
}

func (resourceFileStub) UploadAttachment(context.Context, string, io.Reader, int64) (*domain.Resource, error) {
	return nil, nil
}

func (resourceFileStub) UploadPlatformAttachment(context.Context, string, io.Reader, int64) (*domain.Resource, error) {
	return nil, nil
}

func (resourceFileStub) List(context.Context, int, int) ([]domain.Resource, int64, error) {
	return nil, 0, nil
}

func (resourceFileStub) ListPlatform(context.Context, int, int) ([]domain.Resource, int64, error) {
	return nil, 0, nil
}

func (resourceFileStub) Delete(context.Context, string) error { return nil }

func (resourceFileStub) DeletePlatform(context.Context, string) error { return nil }

func (stub resourceFileStub) File(context.Context, string, string) (string, string, string, error) {
	return stub.path, stub.contentType, stub.fileName, nil
}

func (stub resourceFileStub) OpenPlatformCover(context.Context, string) (io.ReadCloser, string, string, error) {
	return io.NopCloser(bytes.NewReader(nil)), stub.contentType, stub.fileName, nil
}

func requestMultipart(
	t *testing.T, router http.Handler, path, name string, data []byte,
) *httptest.ResponseRecorder {
	return requestMultipartFields(t, router, path, name, data, nil)
}

func requestMultipartFields(
	t *testing.T, router http.Handler, path, name string, data []byte,
	fields map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
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
