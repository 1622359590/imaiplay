package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"sort"
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
	materialRepo := repository.NewCourseMaterialRepository(database)
	categoryRepo := repository.NewResourceCategoryRepository(database)
	auditRepo := repository.NewAuditLogRepository(database)
	dashboardRepo := repository.NewDashboardRepository(database)
	planRepo := repository.NewPlanRepository(database)
	local, err := storage.NewLocal(storage.LocalConfig{Root: t.TempDir(), URL: "/uploads"})
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	auth := service.NewAuthServiceWithRefreshTokens(userRepo, tenantRepo, refreshRepo, integrationSecret)
	auth.SetLoginChallengeRepository(
		repository.NewLoginChallengeRepository(database),
	)
	auth.SetPortalService(service.NewPortalService(tenantRepo, "play.imai.work"))
	planService := service.NewPlanService(planRepo, tenantRepo, resourceRepo)
	resourceService := service.NewResourceService(resourceRepo, local, planService)
	learnerAccess := service.NewLearnerAccess(courseRepo, enrollmentRepo, materialRepo)
	materialService := service.NewCourseMaterialService(courseRepo, materialRepo, resourceRepo, resourceService).
		WithLearnerAccess(learnerAccess)
	deps := server.Dependencies{
		AuthService:               auth,
		TenantService:             service.NewTenantService(tenantRepo),
		TenantRegistrationService: service.NewTenantRegistrationService(database, integrationSecret),
		UserService:               service.NewUserService(userRepo),
		CourseService: service.NewCourseService(
			courseRepo, chapterRepo, lessonRepo, enrollmentRepo, materialRepo,
		),
		CourseMaterialService: materialService,
		LearnerAccessService:  learnerAccess,
		ChapterService:        service.NewCourseChapterService(chapterRepo, courseRepo),
		LessonService: service.NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo, resourceRepo,
		),
		EnrollmentService: service.NewEnrollmentService(enrollmentRepo, courseRepo, userRepo),
		ProgressService: service.NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
			repository.NewLearningTimeRepository(database),
		),
		LearnerOverviewService: service.NewLearnerOverviewService(
			repository.NewLearnerOverviewRepository(database),
		),
		ResourceService:         resourceService,
		ResourceCategoryService: service.NewResourceCategoryService(categoryRepo),
		DashboardService:        service.NewDashboardService(dashboardRepo),
		AuditService:            service.NewAuditService(auditRepo),
		TenantThemeService:      service.NewTenantThemeService(tenantRepo),
		PlanService:             planService,
		TenantRepository:        tenantRepo,
		PortalService:           service.NewPortalService(tenantRepo, "play.imai.work"),
	}
	cfg := config.Config{
		AppName:               "imaiplay",
		AppVersion:            "0.1.0",
		AdminHost:             "play.imai.work",
		JWTSecret:             integrationSecret,
		AuthRateLimit:         100,
		AuthRateWindowSeconds: 60,
		StorageLocalRoot:      t.TempDir(),
		StorageLocalURL:       "/uploads",
	}
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
	published := fx.requestWithToken(http.MethodPut, "/backend/v1/courses/"+courseID, map[string]interface{}{"title": "Go Basics", "status": 1}, adminToken)
	if published.Code != http.StatusOK {
		t.Fatalf("publish status = %d body=%s", published.Code, published.Body.String())
	}
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

func TestSharedLessonResourcePlaybackBindsDeterministicAuthorizedCourse(t *testing.T) {
	fx := newFixture(t)
	adminToken, tenantID := fx.registerAdmin(t)
	var learner domain.User
	if err := fx.db.Where(
		"tenant_id = ? AND email = ?", tenantID, "learner1@example.com",
	).First(&learner).Error; err != nil {
		t.Fatalf("find seeded learner: %v", err)
	}
	learnerToken, err := security.GenerateToken(
		learner.ID, tenantID, learner.Email, learner.Role, integrationSecret,
	)
	if err != nil {
		t.Fatalf("generate learner token: %v", err)
	}
	resourceID := responseID(t, fx.uploadPDF(t, adminToken))

	courseIDs := make([]string, 0, 2)
	for _, title := range []string{"Shared resource alpha", "Shared resource beta"} {
		courseID := responseID(t, fx.requestWithToken(
			http.MethodPost, "/backend/v1/courses",
			map[string]interface{}{"title": title}, adminToken,
		))
		chapterID := responseID(t, fx.requestWithToken(
			http.MethodPost, "/backend/v1/courses/"+courseID+"/chapters",
			map[string]interface{}{"title": "Shared files"}, adminToken,
		))
		lesson := fx.requestWithToken(
			http.MethodPost, "/backend/v1/chapters/"+chapterID+"/lessons",
			map[string]interface{}{
				"title": "Shared guide", "content_type": "document",
				"resource_id": resourceID,
			}, adminToken,
		)
		requireStatus(t, lesson, http.StatusOK)
		published := fx.requestWithToken(
			http.MethodPut, "/backend/v1/courses/"+courseID,
			map[string]interface{}{"title": title, "status": 1}, adminToken,
		)
		requireStatus(t, published, http.StatusOK)
		courseIDs = append(courseIDs, courseID)
	}
	sort.Strings(courseIDs)

	playbackEndpoint := "/api/v1/resources/" + resourceID + "/playback-url"
	requireStatus(
		t,
		fx.requestWithToken(http.MethodGet, playbackEndpoint, nil, learnerToken),
		http.StatusNotFound,
	)
	enroll := func(courseID string) {
		t.Helper()
		response := fx.requestWithToken(
			http.MethodPost, "/backend/v1/courses/"+courseID+"/enrollments",
			map[string]interface{}{"user_id": learner.ID}, adminToken,
		)
		requireStatus(t, response, http.StatusOK)
	}
	issuePlayback := func() (string, *security.PlaybackClaims) {
		t.Helper()
		response := fx.requestWithToken(
			http.MethodGet, playbackEndpoint, nil, learnerToken,
		)
		requireStatus(t, response, http.StatusOK)
		var body struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		}
		decode(t, response, &body)
		parsed, err := url.ParseRequestURI(body.Data.URL)
		if err != nil {
			t.Fatalf("parse playback URL %q: %v", body.Data.URL, err)
		}
		claims, err := security.ValidatePlaybackToken(
			parsed.Query().Get("ticket"), integrationSecret,
		)
		if err != nil {
			t.Fatalf("validate playback token: %v", err)
		}
		if claims.ResourceID != resourceID || claims.UserID != learner.ID ||
			claims.TenantID != tenantID || claims.Role != "learner" {
			t.Fatalf("playback claims = %#v", claims)
		}
		return body.Data.URL, claims
	}

	// Only the lexically later course is enrolled. The repository/service must
	// skip the first candidate and bind the ticket to the authorized course.
	enroll(courseIDs[1])
	laterPlaybackURL, laterClaims := issuePlayback()
	if laterClaims.CourseID != courseIDs[1] {
		t.Fatalf("course_id = %q, want later enrolled %q", laterClaims.CourseID, courseIDs[1])
	}
	served := fx.request(http.MethodGet, laterPlaybackURL, nil)
	requireStatus(t, served, http.StatusOK)
	if !bytes.HasPrefix(served.Body.Bytes(), []byte("%PDF-1.7")) {
		t.Fatalf("playback body = %q", served.Body.String())
	}

	// Once both candidates are enrolled, the lower course ID wins
	// deterministically. Playback reauthorizes, so the older ticket bound to the
	// now non-selected course no longer streams.
	enroll(courseIDs[0])
	firstPlaybackURL, firstClaims := issuePlayback()
	if firstClaims.CourseID != courseIDs[0] {
		t.Fatalf("course_id = %q, want stable first %q", firstClaims.CourseID, courseIDs[0])
	}
	requireStatus(t, fx.request(http.MethodGet, laterPlaybackURL, nil), http.StatusNotFound)
	served = fx.request(http.MethodGet, firstPlaybackURL, nil)
	requireStatus(t, served, http.StatusOK)
	if !bytes.HasPrefix(served.Body.Bytes(), []byte("%PDF-1.7")) {
		t.Fatalf("playback body = %q", served.Body.String())
	}
}

func TestDefaultPortalLearnerLoginAndCourseFlow(t *testing.T) {
	fx := newFixture(t)
	tenant, learner := fx.seedTenantUser(
		t, "acme", "learner@acme.test", "password123", "learner",
	)
	other := fx.seedTenant(t, "bravo", nil)
	acmeCourse := fx.seedPublishedCourse(t, tenant.ID, "Acme course")
	fx.seedPublishedCourse(t, other.ID, "Bravo course")
	if err := fx.db.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		CourseID:  acmeCourse.ID, UserID: learner.ID, Status: 1,
		AssignmentType: domain.AssignmentRequired,
	}).Error; err != nil {
		t.Fatalf("assign Acme course: %v", err)
	}

	portal := fx.requestWithHost(
		http.MethodGet,
		"/api/v1/portal?tenant_code=acme",
		nil,
		"play.imai.work",
	)
	requireStatus(t, portal, http.StatusOK)

	login := fx.requestWithTenant(
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]interface{}{
			"identifier": learner.Email,
			"password":   "password123",
		},
		"",
		tenant.Code,
	)
	token := responseToken(t, login)
	claims := requireTokenClaims(t, token, integrationSecret)
	if claims.TenantID != tenant.ID {
		t.Fatalf("tenant_id=%q want=%q", claims.TenantID, tenant.ID)
	}

	courses := fx.requestWithTenant(
		http.MethodGet,
		"/api/v1/courses",
		nil,
		token,
		tenant.Code,
	)
	requireStatus(t, courses, http.StatusOK)
	requireOnlyTenantOrOfficialCourses(t, courses, tenant.ID)
}

func TestPlatformLoginSelectsOrganizationWithoutLeakingOnWrongPassword(t *testing.T) {
	fx := newFixture(t)
	acme, _ := fx.seedTenantUser(
		t, "acme", "shared@test", "same-pass", "learner",
	)
	fx.seedTenantUser(
		t, "bravo", "shared@test", "same-pass", "instructor",
	)

	wrong := fx.requestWithHost(
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]interface{}{
			"identifier": "shared@test",
			"password":   "wrong",
		},
		"play.imai.work",
	)
	requireStatus(t, wrong, http.StatusUnauthorized)
	requireJSONFieldAbsent(t, wrong, "organizations")

	login := fx.requestWithHost(
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]interface{}{
			"identifier": "shared@test",
			"password":   "same-pass",
		},
		"play.imai.work",
	)
	selectionToken := requireSelectionResponse(t, login, 2)
	selected := fx.requestWithHost(
		http.MethodPost,
		"/api/v1/auth/select-tenant",
		map[string]interface{}{
			"selection_token": selectionToken,
			"tenant_code":     "acme",
		},
		"play.imai.work",
	)
	token := responseToken(t, selected)
	if claims := requireTokenClaims(t, token, integrationSecret); claims.TenantID != acme.ID {
		t.Fatalf("tenant_id=%q want=%q", claims.TenantID, acme.ID)
	}

	replay := fx.requestWithHost(
		http.MethodPost,
		"/api/v1/auth/select-tenant",
		map[string]interface{}{
			"selection_token": selectionToken,
			"tenant_code":     "acme",
		},
		"play.imai.work",
	)
	requireStatus(t, replay, http.StatusUnauthorized)
}

func TestCustomDomainAndDefaultPortalShareTenantButRejectForeignToken(t *testing.T) {
	fx := newFixture(t)
	customDomain := "learn.acme.test"
	acme := fx.seedTenant(t, "acme", &customDomain)
	bravo := fx.seedTenant(t, "bravo", nil)
	acmeToken := fx.learnerToken(t, acme)
	bravoToken := fx.learnerToken(t, bravo)

	byCode := fx.requestWithHost(
		http.MethodGet,
		"/api/v1/portal?tenant_code=acme",
		nil,
		"play.imai.work",
	)
	byDomain := fx.requestWithHost(
		http.MethodGet,
		"/api/v1/portal",
		nil,
		customDomain,
	)
	requireSamePortal(t, byCode, byDomain)

	foreign := fx.requestWithTenant(
		http.MethodGet,
		"/api/v1/courses",
		nil,
		bravoToken,
		acme.Code,
	)
	requireStatus(t, foreign, http.StatusForbidden)
	own := fx.requestWithTenant(
		http.MethodGet,
		"/api/v1/courses",
		nil,
		acmeToken,
		acme.Code,
	)
	requireStatus(t, own, http.StatusOK)
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
	request.Host = "play.imai.work"
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

func (fx *fixture) requestWithHost(
	method, path string,
	payload interface{},
	host string,
) *httptest.ResponseRecorder {
	var body io.Reader
	if payload != nil {
		encoded, _ := json.Marshal(payload)
		body = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, body)
	request.Host = host
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	fx.router.ServeHTTP(response, request)
	return response
}

func (fx *fixture) requestWithTenant(
	method, path string,
	payload interface{},
	token, tenantCode string,
) *httptest.ResponseRecorder {
	var body io.Reader
	if payload != nil {
		encoded, _ := json.Marshal(payload)
		body = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, body)
	request.Host = "play.imai.work"
	request.Header.Set("X-Tenant-Code", tenantCode)
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

func (fx *fixture) seedTenantUser(
	t *testing.T,
	code, email, password, role string,
) (*domain.Tenant, *domain.User) {
	t.Helper()
	tenant := fx.seedTenant(t, code, nil)
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     email,
		Password:  hash,
		Name:      code + " user",
		Role:      role,
		Status:    1,
	}
	if err := fx.db.WithContext(context.Background()).Create(user).Error; err != nil {
		t.Fatalf("create %s user: %v", code, err)
	}
	return tenant, user
}

func (fx *fixture) seedTenant(
	t *testing.T,
	code string,
	customDomain *string,
) *domain.Tenant {
	t.Helper()
	tenant := &domain.Tenant{
		Code:            code,
		Name:            code,
		Status:          1,
		LifecycleStatus: "active",
		CustomDomain:    customDomain,
	}
	if err := fx.db.WithContext(context.Background()).Create(tenant).Error; err != nil {
		t.Fatalf("create tenant %q: %v", code, err)
	}
	return tenant
}

func (fx *fixture) seedPublishedCourse(
	t *testing.T,
	tenantID, title string,
) *domain.Course {
	t.Helper()
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Title:     title,
		Status:    1,
		CreatedBy: "integration",
	}
	if err := fx.db.WithContext(context.Background()).Create(course).Error; err != nil {
		t.Fatalf("create course %q: %v", title, err)
	}
	return course
}

func (fx *fixture) learnerToken(
	t *testing.T,
	tenant *domain.Tenant,
) string {
	t.Helper()
	hash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash learner password: %v", err)
	}
	user := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     tenant.Code + "-learner@test",
		Password:  hash,
		Name:      tenant.Code + " learner",
		Role:      "learner",
		Status:    1,
	}
	if err := fx.db.WithContext(context.Background()).Create(user).Error; err != nil {
		t.Fatalf("create learner for %q: %v", tenant.Code, err)
	}
	token, err := security.GenerateToken(
		user.ID,
		tenant.ID,
		user.Email,
		user.Role,
		integrationSecret,
	)
	if err != nil {
		t.Fatalf("generate learner token: %v", err)
	}
	return token
}

func requireStatus(
	t *testing.T,
	response *httptest.ResponseRecorder,
	want int,
) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status=%d want=%d body=%s", response.Code, want, response.Body.String())
	}
}

func responseToken(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	requireStatus(t, response, http.StatusOK)
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	decode(t, response, &body)
	if body.Data.Token == "" {
		t.Fatalf("response missing token: %s", response.Body.String())
	}
	return body.Data.Token
}

func requireTokenClaims(
	t *testing.T,
	token, secret string,
) *security.Claims {
	t.Helper()
	claims, err := security.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	return claims
}

func requireOnlyTenantOrOfficialCourses(
	t *testing.T,
	response *httptest.ResponseRecorder,
	tenantID string,
) {
	t.Helper()
	var body struct {
		Data struct {
			Items []domain.Course `json:"items"`
		} `json:"data"`
	}
	decode(t, response, &body)
	if len(body.Data.Items) == 0 {
		t.Fatal("published courses response is empty")
	}
	for _, course := range body.Data.Items {
		if course.TenantID != tenantID && !course.IsOfficial {
			t.Fatalf("foreign course returned: %#v", course)
		}
	}
}

func requireJSONFieldAbsent(
	t *testing.T,
	response *httptest.ResponseRecorder,
	field string,
) {
	t.Helper()
	var body map[string]interface{}
	decode(t, response, &body)
	if _, exists := body[field]; exists {
		t.Fatalf("unexpected top-level field %q: %s", field, response.Body.String())
	}
	if data, ok := body["data"].(map[string]interface{}); ok {
		if _, exists := data[field]; exists {
			t.Fatalf("unexpected data field %q: %s", field, response.Body.String())
		}
	}
}

func requireSelectionResponse(
	t *testing.T,
	response *httptest.ResponseRecorder,
	wantOrganizations int,
) string {
	t.Helper()
	requireStatus(t, response, http.StatusOK)
	var body struct {
		Data struct {
			RequiresTenantSelection bool                         `json:"requires_tenant_selection"`
			SelectionToken          string                       `json:"selection_token"`
			Organizations           []service.OrganizationOption `json:"organizations"`
		} `json:"data"`
	}
	decode(t, response, &body)
	if !body.Data.RequiresTenantSelection || body.Data.SelectionToken == "" {
		t.Fatalf("selection response missing challenge: %s", response.Body.String())
	}
	if len(body.Data.Organizations) != wantOrganizations {
		t.Fatalf(
			"organizations=%d want=%d body=%s",
			len(body.Data.Organizations),
			wantOrganizations,
			response.Body.String(),
		)
	}
	return body.Data.SelectionToken
}

func requireSamePortal(
	t *testing.T,
	first, second *httptest.ResponseRecorder,
) {
	t.Helper()
	requireStatus(t, first, http.StatusOK)
	requireStatus(t, second, http.StatusOK)
	var firstBody, secondBody struct {
		Data service.Portal `json:"data"`
	}
	decode(t, first, &firstBody)
	decode(t, second, &secondBody)
	if firstBody.Data.TenantID == "" || firstBody.Data.TenantID != secondBody.Data.TenantID {
		t.Fatalf(
			"portal tenants differ: first=%#v second=%#v",
			firstBody.Data,
			secondBody.Data,
		)
	}
	if firstBody.Data.Code != secondBody.Data.Code ||
		firstBody.Data.DefaultPortalURL != secondBody.Data.DefaultPortalURL {
		t.Fatalf(
			"portal aliases differ: first=%#v second=%#v",
			firstBody.Data,
			secondBody.Data,
		)
	}
}

func (fx *fixture) requestWithToken(method, path string, payload interface{}, token string) *httptest.ResponseRecorder {
	var body io.Reader
	if payload != nil {
		encoded, _ := json.Marshal(payload)
		body = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, body)
	request.Host = "play.imai.work"
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
