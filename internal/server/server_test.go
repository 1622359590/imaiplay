package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/1622359590/imaiplay/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHealthIncludesApplicationAndTenant(t *testing.T) {
	router := New(config.Config{
		ServerPort: "8080",
		AppName:    "imaiplay",
		AppVersion: "0.1.0",
	}, func() error { return nil }, Dependencies{})

	request := httptest.NewRequest(http.MethodGet, "http://localhost/health", nil)
	request.Host = "tenant1.imaiplay.local"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var got struct {
		Status  string `json:"status"`
		AppName string `json:"app_name"`
		Version string `json:"version"`
		Tenant  struct {
			Code   string `json:"code"`
			Source string `json:"source"`
		} `json:"tenant"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != "ok" || got.AppName != "imaiplay" || got.Version != "0.1.0" {
		t.Fatalf("application fields = %#v", got)
	}
	if got.Tenant.Code != "tenant1" || got.Tenant.Source != "subdomain" {
		t.Fatalf("tenant = %#v", got.Tenant)
	}
}

func TestDatabaseHealth(t *testing.T) {
	tests := []struct {
		name         string
		dbCheck      func() error
		wantHTTPCode int
		wantStatus   string
		wantDatabase string
	}{
		{
			name:         "connected",
			dbCheck:      func() error { return nil },
			wantHTTPCode: http.StatusOK,
			wantStatus:   "ok",
			wantDatabase: "connected",
		},
		{
			name:         "disconnected",
			dbCheck:      func() error { return errors.New("connection lost") },
			wantHTTPCode: http.StatusServiceUnavailable,
			wantStatus:   "error",
			wantDatabase: "disconnected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := New(config.Config{}, tt.dbCheck, Dependencies{})
			request := httptest.NewRequest(http.MethodGet, "http://localhost/health/db", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantHTTPCode {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantHTTPCode)
			}
			var got struct {
				Status   string `json:"status"`
				Database string `json:"database"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Status != tt.wantStatus || got.Database != tt.wantDatabase {
				t.Fatalf("response = %#v, want status=%q database=%q",
					got, tt.wantStatus, tt.wantDatabase)
			}
		})
	}
}

func TestSwaggerRouteCanBeEnabledOrDisabled(t *testing.T) {
	for _, test := range []struct {
		name    string
		enabled bool
		want    int
	}{
		{name: "enabled", enabled: true, want: http.StatusOK},
		{name: "disabled", enabled: false, want: http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := New(
				config.Config{SwaggerEnabled: test.enabled},
				func() error { return nil }, Dependencies{},
			)
			request := httptest.NewRequest(
				http.MethodGet, "/swagger/index.html", nil,
			)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestCORSAllowsConfiguredFrontendOrigins(t *testing.T) {
	router := New(config.Config{AllowedOrigins: config.DefaultAllowedOrigins}, func() error { return nil }, Dependencies{})
	for _, origin := range []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:5174",
		"http://127.0.0.1:5175",
	} {
		request := httptest.NewRequest(http.MethodOptions, "/api/v1/courses", nil)
		request.Header.Set("Origin", origin)
		request.Header.Set("Access-Control-Request-Method", http.MethodGet)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("%s status=%d, want %d", origin, response.Code, http.StatusNoContent)
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("allow origin=%q, want %q", got, origin)
		}
		if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("allow credentials=%q", got)
		}
	}

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/courses", nil)
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow origin=%q", got)
	}
}

func TestStaticUploadsAreNoLongerPublic(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "guide.pdf"), []byte("%PDF-1.7\n"), 0o600,
	); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	router := New(config.Config{
		StorageLocalRoot: root,
	}, func() error { return nil }, Dependencies{})
	response := httptest.NewRecorder()
	router.ServeHTTP(
		response, httptest.NewRequest(http.MethodGet, "/uploads/guide.pdf", nil),
	)
	if response.Code != http.StatusNotFound {
		t.Fatalf("legacy static status=%d, want %d", response.Code, http.StatusNotFound)
	}
	missing := httptest.NewRecorder()
	router.ServeHTTP(
		missing, httptest.NewRequest(http.MethodGet, "/uploads/missing.pdf", nil),
	)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("legacy missing status=%d", missing.Code)
	}
}

func TestResourceFileRoutesRequireAuthentication(t *testing.T) {
	router := New(config.Config{StorageLocalRoot: t.TempDir(), JWTSecret: "secret"}, func() error { return nil }, Dependencies{})
	for _, path := range []string{
		"/api/v1/resources/resource-1/file",
		"/backend/v1/resources/resource-1/file",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s status=%d, want %d", path, response.Code, http.StatusUnauthorized)
		}
	}
}

func TestDashboardRouteUsesAuthenticatedManager(t *testing.T) {
	router := New(
		config.Config{JWTSecret: "secret"},
		func() error { return nil },
		Dependencies{DashboardService: serverDashboardStub{}},
	)
	token, err := security.GenerateToken(
		"admin", "tenant-1", "admin@example.com", "tenant_admin", "secret",
	)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	request := httptest.NewRequest(
		http.MethodGet, "/backend/v1/dashboard", nil,
	)
	request.Host = "acme.imaiplay.local"
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf(
			"status=%d body=%s", response.Code, response.Body.String(),
		)
	}
}

type serverDashboardStub struct{}

func (serverDashboardStub) Stats(
	context.Context,
) (service.DashboardStats, error) {
	return service.DashboardStats{UserCount: 1}, nil
}

func TestBackendRoutesRequireJWTAndRole(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	tenantRepo := repository.NewTenantRepository(database)
	userRepo := repository.NewUserRepository(database)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	progressRepo := repository.NewLessonProgressRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	categoryRepo := repository.NewResourceCategoryRepository(database)
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := tenantRepo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	deps := Dependencies{
		AuthService:   service.NewAuthService(userRepo, tenantRepo, "secret"),
		TenantService: service.NewTenantService(tenantRepo),
		UserService:   service.NewUserService(userRepo),
		CourseService: service.NewCourseService(courseRepo, chapterRepo, lessonRepo),
		ChapterService: service.NewCourseChapterService(
			chapterRepo, courseRepo,
		),
		LessonService: service.NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo,
		),
		EnrollmentService: service.NewEnrollmentService(
			enrollmentRepo, courseRepo, userRepo,
		),
		ProgressService: service.NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
		),
		ResourceService: service.NewResourceService(
			resourceRepo, mustLocalStorage(t),
		),
		ResourceCategoryService: service.NewResourceCategoryService(categoryRepo),
	}
	router := New(config.Config{JWTSecret: "secret"}, func() error { return nil }, deps)

	assertRouteStatus(t, router, "/backend/v1/tenants", "", http.StatusUnauthorized)
	tenantAdminToken, err := security.GenerateToken(
		"admin", tenant.ID, "admin@example.com", "tenant_admin", "secret",
	)
	if err != nil {
		t.Fatalf("generate tenant admin token: %v", err)
	}
	assertRouteStatus(
		t, router, "/backend/v1/tenants", tenantAdminToken, http.StatusForbidden,
	)
	superadminToken, err := security.GenerateToken(
		"root", "", "root@example.com", "superadmin", "secret",
	)
	if err != nil {
		t.Fatalf("generate superadmin token: %v", err)
	}
	assertRouteStatus(
		t, router, "/backend/v1/tenants", superadminToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/backend/v1/users", tenantAdminToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/backend/v1/courses", tenantAdminToken, http.StatusOK,
	)
	learnerToken, err := security.GenerateToken(
		"learner", tenant.ID, "learner@example.com", "learner", "secret",
	)
	if err != nil {
		t.Fatalf("generate learner token: %v", err)
	}
	assertRouteStatus(
		t, router, "/api/v1/courses", learnerToken, http.StatusOK,
	)
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Title:     "Course", CreatedBy: "admin", Status: 1,
	}
	if err := courseRepo.Create(context.Background(), course); err != nil {
		t.Fatalf("create course: %v", err)
	}
	assertRouteStatus(
		t, router, "/backend/v1/courses/"+course.ID+"/enrollments",
		tenantAdminToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/api/v1/recent-learning", learnerToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/backend/v1/resources", tenantAdminToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/backend/v1/admin/resources",
		tenantAdminToken, http.StatusForbidden,
	)
	assertRouteStatus(
		t, router, "/backend/v1/admin/resources",
		superadminToken, http.StatusOK,
	)
	assertRouteStatus(
		t, router, "/api/v1/platform-covers/missing", "",
		http.StatusNotFound,
	)
}

func mustLocalStorage(t *testing.T) *storage.Local {
	t.Helper()
	local, err := storage.NewLocal(storage.LocalConfig{
		Root: t.TempDir(), URL: "/uploads",
	})
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	return local
}

func assertRouteStatus(
	t *testing.T,
	handler http.Handler,
	path, token string,
	want int,
) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Host = "acme.imaiplay.local"
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s status = %d, want %d body=%s",
			path, response.Code, want, response.Body.String())
	}
}
