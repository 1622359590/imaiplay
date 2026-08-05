package api

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/1622359590/imaiplay/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func withRole(role, tenantID, userID string) context.Context {
	return usercontext.WithUser(
		context.Background(), userID, tenantID, userID+"@example.com", role,
	)
}

type testServices struct {
	database           *gorm.DB
	auth               *service.AuthService
	tenants            *service.TenantService
	users              *service.UserService
	courses            *service.CourseService
	chapters           *service.CourseChapterService
	lessons            *service.CourseLessonService
	enrollments        *service.EnrollmentService
	enrollmentRepo     repository.CourseEnrollmentRepository
	progress           *service.ProgressService
	overview           *service.LearnerOverviewService
	resources          *service.ResourceService
	materials          *service.CourseMaterialService
	resourceCategories *service.ResourceCategoryService
}

func newTestServices(t *testing.T) (testServices, repository.TenantRepository) {
	t.Helper()
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
	learningTimeRepo := repository.NewLearningTimeRepository(database)
	overviewRepo := repository.NewLearnerOverviewRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	materialRepo := repository.NewCourseMaterialRepository(database)
	categoryRepo := repository.NewResourceCategoryRepository(database)
	authService := service.NewAuthServiceWithRefreshTokens(
		userRepo,
		tenantRepo,
		repository.NewRefreshTokenRepository(database),
		"secret",
	)
	authService.SetLoginChallengeRepository(
		repository.NewLoginChallengeRepository(database),
	)
	authService.SetPortalService(
		service.NewPortalService(tenantRepo, "play.imai.work"),
	)
	localStorage, err := storage.NewLocal(storage.LocalConfig{
		Root: t.TempDir(), URL: "/uploads",
	})
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	resourceService := service.NewResourceService(resourceRepo, localStorage)
	return testServices{
		database: database,
		auth:     authService,
		tenants:  service.NewTenantService(tenantRepo),
		users:    service.NewUserService(userRepo),
		courses: service.NewCourseService(
			courseRepo, chapterRepo, lessonRepo, enrollmentRepo, materialRepo,
		),
		chapters: service.NewCourseChapterService(chapterRepo, courseRepo),
		lessons: service.NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo, resourceRepo,
		),
		enrollments: service.NewEnrollmentService(
			enrollmentRepo, courseRepo, userRepo,
		),
		enrollmentRepo: enrollmentRepo,
		progress: service.NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
			learningTimeRepo,
		),
		overview:           service.NewLearnerOverviewService(overviewRepo),
		resources:          resourceService,
		materials:          service.NewCourseMaterialService(courseRepo, materialRepo, resourceRepo, resourceService),
		resourceCategories: service.NewResourceCategoryService(categoryRepo),
	}, tenantRepo
}

func asRole(role, tenantID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := usercontext.WithUser(
			c.Request.Context(), "user-1", tenantID, "admin@example.com", role,
		)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func createTenant(t *testing.T, repo repository.TenantRepository) *domain.Tenant {
	t.Helper()
	tenant := &domain.Tenant{Code: "acme", Name: "Acme", Status: 1}
	if err := repo.Create(context.Background(), tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	return tenant
}
