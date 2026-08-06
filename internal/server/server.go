package server

import (
	"github.com/1622359590/imaiplay/internal/api"
	"github.com/1622359590/imaiplay/internal/config"
	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"net/url"
	"strings"
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
	LearnerOverviewService    api.LearnerOverviewService
	LearnerAccessService      api.LearnerAccessService
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
