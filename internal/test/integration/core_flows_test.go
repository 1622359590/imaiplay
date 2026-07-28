package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/db"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/1622359590/imaiplay/internal/server"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/1622359590/imaiplay/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const integrationSecret = "integration-test-secret"

type fixture struct {
	router *gin.Engine
	db     *gorm.DB
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared&_foreign_keys=1"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("get sqlite database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	tenantRepo := repository.NewTenantRepository(database)
	userRepo := repository.NewUserRepository(database)
	refreshRepo := repository.NewRefreshTokenRepository(database)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	progressRepo := repository.NewLessonProgressRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	categoryRepo := repository.NewResourceCategoryRepository(database)
	auditRepo := repository.NewAuditLogRepository(database)
	dashboardRepo := repository.NewDashboardRepository(database)
	planRepo := repository.NewPlanRepository(database)
	local, err := storage.NewLocal(storage.LocalConfig{Root: t.TempDir(), URL: "/uploads"})
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	auth := service.NewAuthServiceWithRefreshTokens(userRepo, tenantRepo, refreshRepo, integrationSecret)
	planService := service.NewPlanService(planRepo, tenantRepo, resourceRepo)
	deps := server.Dependencies{
		AuthService:               auth,
		TenantService:             service.NewTenantService(tenantRepo),
		TenantRegistrationService: service.NewTenantRegistrationService(database, integrationSecret),
		UserService:               service.NewUserService(userRepo),
		CourseService:             service.NewCourseService(courseRepo, chapterRepo, lessonRepo),
		ChapterService:            service.NewCourseChapterService(chapterRepo, courseRepo),
		LessonService:             service.NewCourseLessonService(lessonRepo, chapterRepo, courseRepo),
		EnrollmentService:         service.NewEnrollmentService(enrollmentRepo, courseRepo, userRepo),
		ProgressService:           service.NewProgressService(progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo),
		ResourceService:           service.NewResourceService(resourceRepo, local, planService),
		ResourceCategoryService:   service.NewResourceCategoryService(categoryRepo),
		DashboardService:          service.NewDashboardService(dashboardRepo),
		AuditService:              service.NewAuditService(auditRepo),
		TenantThemeService:        service.NewTenantThemeService(tenantRepo),
		PlanService:               planService,
		TenantRepository:          tenantRepo,
	}
	cfg := config.Config{AppName: "imaiplay", AppVersion: "0.1.0", JWTSecret: integrationSecret, StorageLocalRoot: t.TempDir(), StorageLocalURL: "/uploads"}
	return &fixture{router: server.New(cfg, func() error { return db.Ping(database) }, deps), db: database}
}

func TestTenantRegistrationLoginAndDashboard(t *testing.T) {
	fx := newFixture(t)
	response := fx.request(http.MethodPost, "/api/v1/tenants/register", map[string]interface{}{
		"organization_name": "Acme Learning", "admin_email": "admin@acme.test", "admin_name": "Acme Admin", "password": "password123",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("register status = %d body=%s", response.Code, response.Body.String())
	}
	var result struct {
		Data struct {
			Tenant domain.Tenant `json:"tenant"`
			Token  string        `json:"token"`
		} `json:"data"`
	}
	decode(t, response, &result)
	if result.Data.Tenant.ID == "" || result.Data.Token == "" {
		t.Fatalf("registration result = %#v", result.Data)
	}

	dashboard := fx.requestWithToken(http.MethodGet, "/backend/v1/dashboard", nil, result.Data.Token)
	if dashboard.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d body=%s", dashboard.Code, dashboard.Body.String())
	}
	var dashboardResult struct {
		Code int `json:"code"`
	}
	decode(t, dashboard, &dashboardResult)
	if dashboardResult.Code != 0 {
		t.Fatalf("dashboard response = %#v", dashboardResult)
	}
}

func TestCourseEnrollmentAndProgressFlow(t *testing.T) {
	fx := newFixture(t)
	adminToken, tenantID := fx.registerAdmin(t)
	var learner domain.User
	if err := fx.db.Where("tenant_id = ? AND email = ?", tenantID, "learner1@example.com").First(&learner).Error; err != nil {
		t.Fatalf("find seeded learner: %v", err)
	}
	course := fx.requestWithToken(http.MethodPost, "/backend/v1/courses", map[string]interface{}{"title": "Go Basics"}, adminToken)
	courseID := responseID(t, course)
	chapter := fx.requestWithToken(http.MethodPost, "/backend/v1/courses/"+courseID+"/chapters", map[string]interface{}{"title": "Intro"}, adminToken)
	chapterID := responseID(t, chapter)
	lesson := fx.requestWithToken(http.MethodPost, "/backend/v1/chapters/"+chapterID+"/lessons", map[string]interface{}{"title": "Welcome", "content_type": "text", "content_url": "/welcome"}, adminToken)
	lessonID := responseID(t, lesson)
	enroll := fx.requestWithToken(http.MethodPost, "/backend/v1/courses/"+courseID+"/enrollments", map[string]interface{}{"user_id": learner.ID}, adminToken)
	if enroll.Code != http.StatusOK {
		t.Fatalf("enrollment status = %d body=%s", enroll.Code, enroll.Body.String())
	}
	learnerToken, err := security.GenerateToken(learner.ID, tenantID, learner.Email, learner.Role, integrationSecret)
	if err != nil {
		t.Fatalf("generate learner token: %v", err)
	}
	progress := fx.requestWithToken(http.MethodPost, "/api/v1/lessons/"+lessonID+"/progress", map[string]interface{}{"position_seconds": 30, "progress_percent": 50}, learnerToken)
	if progress.Code != http.StatusOK {
		t.Fatalf("progress status = %d body=%s", progress.Code, progress.Body.String())
	}
}

func TestUploadResourceAndReferenceFromLesson(t *testing.T) {
	fx := newFixture(t)
	adminToken, _ := fx.registerAdmin(t)
	resource := fx.uploadPDF(t, adminToken)
	resourceID := responseID(t, resource)
	courseID := responseID(t, fx.requestWithToken(http.MethodPost, "/backend/v1/courses", map[string]interface{}{"title": "Resource Course"}, adminToken))
	chapterID := responseID(t, fx.requestWithToken(http.MethodPost, "/backend/v1/courses/"+courseID+"/chapters", map[string]interface{}{"title": "Files"}, adminToken))
	lesson := fx.requestWithToken(http.MethodPost, "/backend/v1/chapters/"+chapterID+"/lessons", map[string]interface{}{"title": "Guide", "content_type": "document", "resource_id": resourceID}, adminToken)
	if lesson.Code != http.StatusOK {
		t.Fatalf("lesson status = %d body=%s", lesson.Code, lesson.Body.String())
	}
	var result struct {
		Data domain.CourseLesson `json:"data"`
	}
	decode(t, lesson, &result)
	if result.Data.ResourceID == nil || *result.Data.ResourceID != resourceID {
		t.Fatalf("lesson resource_id = %v, want %q", result.Data.ResourceID, resourceID)
	}
}

func (fx *fixture) registerAdmin(t *testing.T) (string, string) {
	t.Helper()
	response := fx.request(http.MethodPost, "/api/v1/tenants/register", map[string]interface{}{"organization_name": "Flow Tenant", "admin_email": "admin@flow.test", "admin_name": "Admin", "password": "password123"})
	var result struct {
		Data struct {
			Tenant domain.Tenant `json:"tenant"`
			Token  string        `json:"token"`
		} `json:"data"`
	}
	if response.Code != http.StatusOK {
		t.Fatalf("register status = %d body=%s", response.Code, response.Body.String())
	}
	decode(t, response, &result)
	return result.Data.Token, result.Data.Tenant.ID
}

func (fx *fixture) uploadPDF(t *testing.T, token string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="guide.pdf"`)
	header.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create multipart part: %v", err)
	}
	_, _ = part.Write([]byte("%PDF-1.7 integration fixture"))
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/backend/v1/resources/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	fx.router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("upload status = %d body=%s", response.Code, response.Body.String())
	}
	return response
}

func (fx *fixture) request(method, path string, payload interface{}) *httptest.ResponseRecorder {
	return fx.requestWithToken(method, path, payload, "")
}

func (fx *fixture) requestWithToken(method, path string, payload interface{}, token string) *httptest.ResponseRecorder {
	var body io.Reader
	if payload != nil {
		encoded, _ := json.Marshal(payload)
		body = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, body)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	fx.router.ServeHTTP(response, request)
	return response
}

func decode(t *testing.T, response *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode %s: %v", response.Body.String(), err)
	}
}

func responseID(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("response status = %d body=%s", response.Code, response.Body.String())
	}
	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decode(t, response, &result)
	if result.Data.ID == "" {
		t.Fatalf("response missing id: %s", response.Body.String())
	}
	return result.Data.ID
}
