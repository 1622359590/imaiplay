package server

import (
	"github.com/1622359590/imaiplay/internal/api"
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

type routeHandlers struct {
	portal           *api.PortalHandler
	auth             *api.AuthHandler
	theme            *api.ThemeHandler
	registration     *api.TenantRegistrationHandler
	plan             *api.PlanHandler
	tenant           *api.TenantHandler
	user             *api.UserHandler
	course           *api.CourseHandler
	chapter          *api.CourseChapterHandler
	lesson           *api.CourseLessonHandler
	material         *api.CourseMaterialHandler
	enrollment       *api.EnrollmentHandler
	resource         *api.ResourceHandler
	resourceCategory *api.ResourceCategoryHandler
	courseCategory   *api.CourseCategoryHandler
	dashboard        *api.DashboardHandler
	audit            *api.AuditHandler
	progress         *api.ProgressHandler
	learnerOverview  *api.LearnerOverviewHandler
}

func newRouteHandlers(cfg config.Config, deps Dependencies) routeHandlers {
	return routeHandlers{
		portal: api.NewPortalHandler(deps.PortalService), auth: api.NewAuthHandler(deps.AuthService).WithBootstrapSecret(cfg.SuperadminBootstrapSecret),
		theme: api.NewThemeHandler(deps.TenantThemeService), registration: api.NewTenantRegistrationHandler(deps.TenantRegistrationService),
		plan: api.NewPlanHandler(deps.PlanService), tenant: api.NewTenantHandler(deps.TenantService), user: api.NewUserHandler(deps.UserService),
		course: api.NewCourseHandler(deps.CourseService), chapter: api.NewCourseChapterHandler(deps.ChapterService), lesson: api.NewCourseLessonHandler(deps.LessonService),
		material: api.NewCourseMaterialHandler(deps.CourseMaterialService), enrollment: api.NewEnrollmentHandler(deps.EnrollmentService),
		resource:         api.NewResourceHandler(deps.ResourceService, cfg.StorageLocalRoot).WithLearnerAccess(deps.LearnerAccessService).WithPlaybackSecret(cfg.JWTSecret),
		resourceCategory: api.NewResourceCategoryHandler(deps.ResourceCategoryService), courseCategory: api.NewCourseCategoryHandler(deps.CourseCategoryService),
		dashboard: api.NewDashboardHandler(deps.DashboardService), audit: api.NewAuditHandler(deps.AuditService),
		progress: api.NewProgressHandler(deps.ProgressService), learnerOverview: api.NewLearnerOverviewHandler(deps.LearnerOverviewService),
	}
}

func registerRoutes(router *gin.Engine, cfg config.Config, deps Dependencies) {
	handlers := newRouteHandlers(cfg, deps)
	registerAuthRoutes(router, cfg, deps, handlers)
	backend := router.Group("/backend/v1")
	backend.Use(middleware.AdminHost(cfg.AdminHost), middleware.TenantWithRepositoryForPlatformHost(deps.TenantRepository, cfg.AdminHost), middleware.Auth(cfg.JWTSecret), middleware.TenantMatch(deps.TenantRepository), middleware.TenantAccess(deps.TenantRepository))
	registerAdminRoutes(backend, deps, handlers)
	registerCourseRoutes(backend, handlers)
	registerInfrastructureRoutes(router, backend, cfg, deps, handlers)
	registerLearnerRoutes(router, cfg, deps, handlers)
}
