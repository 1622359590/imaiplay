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
	auth               *service.AuthService
	tenants            *service.TenantService
	users              *service.UserService
	courses            *service.CourseService
	chapters           *service.CourseChapterService
	lessons            *service.CourseLessonService
	enrollments        *service.EnrollmentService
	progress           *service.ProgressService
	resources          *service.ResourceService
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
	resourceRepo := repository.NewResourceRepository(database)
	categoryRepo := repository.NewResourceCategoryRepository(database)
	localStorage, err := storage.NewLocal(storage.LocalConfig{
		Root: t.TempDir(), URL: "/uploads",
	})
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}
	return testServices{
		auth:     service.NewAuthService(userRepo, tenantRepo, "secret"),
		tenants:  service.NewTenantService(tenantRepo),
		users:    service.NewUserService(userRepo),
		courses:  service.NewCourseService(courseRepo, chapterRepo, lessonRepo),
		chapters: service.NewCourseChapterService(chapterRepo, courseRepo),
		lessons: service.NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo, resourceRepo,
		),
		enrollments: service.NewEnrollmentService(
			enrollmentRepo, courseRepo, userRepo,
		),
		progress: service.NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
		),
		resources:          service.NewResourceService(resourceRepo, localStorage),
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
