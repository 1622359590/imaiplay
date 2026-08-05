package server

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1622359590/imaiplay/internal/api"
	"github.com/1622359590/imaiplay/internal/config"
	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Dependencies struct {
	AuthService               api.AuthService
	TenantService             api.TenantService
	UserService               api.UserService
	CourseService             api.CourseService
	CourseMaterialService     api.CourseMaterialService
	ChapterService            api.CourseChapterService
	LessonService             api.CourseLessonService
	EnrollmentService         api.EnrollmentService
	ProgressService           api.ProgressService
	ResourceService           api.ResourceService
	ResourceCategoryService   api.ResourceCategoryService
	CourseCategoryService     api.CourseCategoryService
	DashboardService          api.DashboardService
	TenantRegistrationService api.TenantRegistrationService
	SMSConfigService          api.SMSConfigService
	AuditService              api.AuditService
	TenantThemeService        api.TenantThemeService
	PlanService               api.PlanService
	TenantRepository          repository.TenantRepository
	StorageConfigService      api.StorageConfigService
	DomainBindService         api.DomainBindService
	PortalService             api.PortalResolver
}

func New(
	cfg config.Config,
	dbCheck func() error,
	deps Dependencies,
) *gin.Engine {
	router := gin.New()
	logger := middleware.NewLogger(cfg.LogLevel, cfg.LogFormat)
	origins := cfg.AllowedOrigins
	if origins == "" {
		origins = config.DefaultAllowedOrigins
	}
	router.Use(cors(origins, deps.TenantRepository), middleware.RequestID(), middleware.Logging(logger), gin.Recovery(), middleware.PanicLogging(logger), middleware.Audit(deps.AuditService))
	router.GET("/health", middleware.TenantWithRepository(deps.TenantRepository), func(c *gin.Context) {
		code, source := tenantcontext.TenantFromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"app_name": cfg.AppName,
			"version":  cfg.AppVersion,
			"tenant": gin.H{
				"code":   code,
				"source": source,
			},
		})
	})
	router.GET("/health/db", func(c *gin.Context) {
		if err := dbCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})
	if cfg.SwaggerEnabled {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	registerRoutes(router, cfg, deps)
	return router
}

func registerRoutes(
	router *gin.Engine,
	cfg config.Config,
	deps Dependencies,
) {
	portalHandler := api.NewPortalHandler(deps.PortalService)
	router.GET("/api/v1/portal", portalHandler.Get)
	portalSession := router.Group("/api/v1/portal")
	portalSession.Use(middleware.Auth(cfg.JWTSecret))
	portalSession.GET("/session", portalHandler.GetSession)
	authHandler := api.NewAuthHandler(deps.AuthService)
	themeHandler := api.NewThemeHandler(deps.TenantThemeService)
	theme := router.Group("/api/v1")
	theme.Use(middleware.TenantWithRepository(deps.TenantRepository))
	theme.GET("/theme", themeHandler.Get)
	router.POST("/api/v1/bootstrap/superadmin", authHandler.BootstrapSuperadmin)
	auth := router.Group("/api/v1/auth")
	auth.Use(middleware.TenantWithRepositoryForAdminHost(deps.TenantRepository, cfg.AdminHost))
	limiter := middleware.NewRateLimiter(cfg.AuthRateLimit, time.Duration(cfg.AuthRateWindowSeconds)*time.Second)
	auth.POST("/register", limiter.Handler(), authHandler.Register)
	auth.POST("/login", limiter.Handler(), authHandler.Login)
	auth.POST("/select-tenant", limiter.Handler(), authHandler.SelectTenant)
	auth.POST("/login-code/send", limiter.Handler(), authHandler.SendLoginCode)
	auth.POST("/login-code", limiter.Handler(), authHandler.LoginWithCode)
	auth.POST("/refresh", authHandler.Refresh)
	authProtected := router.Group("/api/v1/auth")
	authProtected.Use(middleware.TenantWithRepositoryForAdminHost(deps.TenantRepository, cfg.AdminHost), middleware.Auth(cfg.JWTSecret), middleware.TenantAccess(deps.TenantRepository))
	authProtected.POST("/logout", authHandler.Logout)
	registrationHandler := api.NewTenantRegistrationHandler(deps.TenantRegistrationService)
	registration := router.Group("/api/v1/tenants")
	registration.Use(middleware.TenantWithRepository(deps.TenantRepository))
	registration.POST("/register", limiter.Handler(), registrationHandler.Register)
	auth.POST("/forgot-password", limiter.Handler(), authHandler.ForgotPassword)
	auth.POST("/reset-password", authHandler.ResetPassword)

	backend := router.Group("/backend/v1")
	backend.Use(middleware.AdminHost(cfg.AdminHost), middleware.TenantWithRepositoryForPlatformHost(deps.TenantRepository, cfg.AdminHost), middleware.Auth(cfg.JWTSecret), middleware.TenantMatch(deps.TenantRepository), middleware.TenantAccess(deps.TenantRepository))
	planHandler := api.NewPlanHandler(deps.PlanService)
	backend.GET("/plans", planHandler.List)
	backend.POST("/plans", planHandler.Create)
	backend.PUT("/plans/:id", planHandler.Update)
	backend.DELETE("/plans/:id", planHandler.Delete)
	backend.PUT("/tenant-plans/:tenantID", planHandler.Assign)
	backend.GET("/plan/current", planHandler.Current)
	backend.GET("/theme", themeHandler.Get)
	backend.PUT("/theme", themeHandler.Update)
	tenantHandler := api.NewTenantHandler(deps.TenantService)
	backend.POST("/tenants", tenantHandler.Create)
	backend.POST("/admin/tenants", registrationHandler.CreateForSuperadmin)
	backend.GET("/tenants", tenantHandler.List)
	backend.GET("/tenants/:id", tenantHandler.Get)
	backend.PUT("/tenants/:id", tenantHandler.Update)
	backend.DELETE("/tenants/:id", tenantHandler.Delete)
	if deps.DomainBindService != nil {
		domainBindHandler := api.NewDomainBindHandler(deps.DomainBindService)
		backend.POST("/domain-bind/verify", domainBindHandler.Verify)
		backend.POST("/domain-bind", domainBindHandler.Bind)
		backend.GET("/domain-bind/status", domainBindHandler.Status)
		backend.DELETE("/domain-bind", domainBindHandler.Unbind)
		backend.POST("/tenants/:id/domain-bind/verify", domainBindHandler.VerifyForTenant)
		backend.POST("/tenants/:id/domain-bind", domainBindHandler.BindForTenant)
		backend.GET("/tenants/:id/domain-bind/status", domainBindHandler.StatusForTenant)
		backend.DELETE("/tenants/:id/domain-bind", domainBindHandler.UnbindForTenant)
	}
	userHandler := api.NewUserHandler(deps.UserService)
	backend.POST("/users", userHandler.Create)
	backend.GET("/users", userHandler.List)
	backend.GET("/users/:id", userHandler.Get)
	backend.PUT("/users/:id", userHandler.Update)
	backend.DELETE("/users/:id", userHandler.Delete)
	backend.PUT("/users/:id/password", userHandler.ResetTenantAdminPassword)

	courseHandler := api.NewCourseHandler(deps.CourseService)
	chapterHandler := api.NewCourseChapterHandler(deps.ChapterService)
	lessonHandler := api.NewCourseLessonHandler(deps.LessonService)
	backend.POST("/courses", courseHandler.Create)
	backend.POST("/official-courses", courseHandler.CreateOfficial)
	backend.GET("/official-courses", courseHandler.OfficialList)
	backend.PUT("/official-courses/:id/enabled", courseHandler.EnableOfficial)
	backend.GET("/courses", courseHandler.List)
	backend.GET("/courses/:id", courseHandler.Get)
	backend.PUT("/courses/:id", courseHandler.Update)
	backend.DELETE("/courses/:id", courseHandler.Delete)
	backend.GET("/courses/:id/detail", courseHandler.Detail)
	materialHandler := api.NewCourseMaterialHandler(deps.CourseMaterialService)
	backend.GET("/courses/:id/materials", materialHandler.List)
	backend.POST("/courses/:id/materials", materialHandler.Add)
	backend.PUT("/courses/:id/materials/:materialID", materialHandler.Update)
	backend.DELETE("/courses/:id/materials/:materialID", materialHandler.Remove)
	backend.POST("/courses/:id/chapters", chapterHandler.Create)
	backend.GET("/courses/:id/chapters", chapterHandler.List)
	backend.PUT("/chapters/:id", chapterHandler.Update)
	backend.DELETE("/chapters/:id", chapterHandler.Delete)
	backend.POST("/chapters/:id/lessons", lessonHandler.Create)
	backend.GET("/chapters/:id/lessons", lessonHandler.List)
	backend.PUT("/lessons/:id", lessonHandler.Update)
	backend.DELETE("/lessons/:id", lessonHandler.Delete)
	enrollmentHandler := api.NewEnrollmentHandler(deps.EnrollmentService)
	backend.POST("/courses/:id/enrollments", enrollmentHandler.Enroll)
	backend.GET("/courses/:id/enrollments", enrollmentHandler.ListByCourse)
	backend.DELETE("/enrollments/:id", enrollmentHandler.Remove)
	resourceHandler := api.NewResourceHandler(
		deps.ResourceService, cfg.StorageLocalRoot,
	).WithPlaybackSecret(cfg.JWTSecret)
	backend.POST("/resources/upload", resourceHandler.Upload)
	backend.POST("/resources/attachments/upload", resourceHandler.UploadAttachment)
	backend.GET("/resources", resourceHandler.List)
	backend.POST("/admin/resources/upload", resourceHandler.UploadPlatform)
	backend.POST("/admin/resources/attachments/upload", resourceHandler.UploadPlatformAttachment)
	backend.GET("/admin/resources", resourceHandler.ListPlatform)
	backend.GET("/admin/resources/:id/file", resourceHandler.File)
	backend.DELETE("/admin/resources/:id", resourceHandler.DeletePlatform)
	backend.GET("/resources/:id/file", resourceHandler.File)
	backend.DELETE("/resources/:id", resourceHandler.Delete)
	router.GET("/api/v1/platform-covers/:id", resourceHandler.PlatformCover)
	categoryHandler := api.NewResourceCategoryHandler(
		deps.ResourceCategoryService,
	)
	backend.POST("/resource-categories", categoryHandler.Create)
	backend.GET("/resource-categories", categoryHandler.List)
	backend.PUT("/resource-categories/:id", categoryHandler.Update)
	backend.DELETE("/resource-categories/:id", categoryHandler.Delete)
	courseCategoryHandler := api.NewCourseCategoryHandler(
		deps.CourseCategoryService,
	)
	backend.GET("/course-categories", courseCategoryHandler.List)
	backend.POST("/course-categories", courseCategoryHandler.Create)
	backend.PUT("/course-categories/:id", courseCategoryHandler.Update)
	backend.DELETE("/course-categories/:id", courseCategoryHandler.Delete)
	backend.GET("/admin/course-categories", courseCategoryHandler.ListPlatform)
	backend.POST("/admin/course-categories", courseCategoryHandler.CreatePlatform)
	backend.PUT("/admin/course-categories/:id", courseCategoryHandler.UpdatePlatform)
	backend.DELETE("/admin/course-categories/:id", courseCategoryHandler.DeletePlatform)
	dashboardHandler := api.NewDashboardHandler(deps.DashboardService)
	backend.GET("/dashboard", dashboardHandler.Get)
	backend.DELETE("/tenants/demo-data", registrationHandler.ClearDemoData)
	if deps.SMSConfigService != nil {
		smsHandler := api.NewSMSConfigHandler(deps.SMSConfigService)
		backend.GET("/sms-config", smsHandler.Get)
		backend.PUT("/sms-config", smsHandler.Save)
		backend.POST("/sms-config/test", smsHandler.Test)
	}
	if deps.StorageConfigService != nil {
		storageHandler := api.NewStorageConfigHandler(deps.StorageConfigService)
		backend.GET("/storage-config", storageHandler.Get)
		backend.PUT("/storage-config", storageHandler.Save)
		backend.POST("/storage-config/test", storageHandler.Test)
	}
	auditHandler := api.NewAuditHandler(deps.AuditService)
	backend.GET("/audit-logs", auditHandler.ListTenant)
	backend.GET("/admin/audit-logs", auditHandler.ListAdmin)

	student := router.Group("/api/v1")
	student.Use(middleware.TenantWithRepositoryForPlatformHost(deps.TenantRepository, cfg.AdminHost), middleware.Auth(cfg.JWTSecret), middleware.TenantMatch(deps.TenantRepository), middleware.TenantAccess(deps.TenantRepository))
	student.GET("/courses", courseHandler.PublishedList)
	student.GET("/courses/:id", courseHandler.PublishedDetail)
	student.GET("/course-materials/:id/download", materialHandler.Download)
	student.GET("/resources/:id/file", resourceHandler.File)
	student.GET("/resources/:id/playback-url", resourceHandler.PlaybackURL)
	progressHandler := api.NewProgressHandler(deps.ProgressService)
	student.POST("/lessons/:id/progress", progressHandler.Report)
	student.GET("/lessons/:id/progress", progressHandler.Get)
	student.GET("/recent-learning", progressHandler.Recent)
	router.GET("/api/v1/resource-playback/:id", resourceHandler.Playback)
}

func cors(configuredOrigins string, tenants repository.TenantRepository) gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(configuredOrigins)
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false
		if _, ok := allowedOrigins[origin]; ok {
			allowed = true
		}
		if tenants != nil && origin != "" {
			if parsed, err := url.Parse(origin); err == nil {
				host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
				if tenant, err := tenants.FindByCustomDomain(c.Request.Context(), host); err == nil && tenant.CustomDomain != nil {
					allowed = true
				}
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header(
				"Access-Control-Allow-Headers",
				"Content-Type, Authorization, X-Tenant-Code, X-Tenant-ID",
			)
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func parseAllowedOrigins(value string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}
	return allowed
}

func Run(
	cfg config.Config,
	dbCheck func() error,
	deps Dependencies,
) error {
	return New(cfg, dbCheck, deps).Run(":" + cfg.ServerPort)
}
