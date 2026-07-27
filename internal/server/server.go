package server

import (
	"net/http"

	"github.com/1622359590/imaiplay/internal/api"
	"github.com/1622359590/imaiplay/internal/config"
	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthService             api.AuthService
	TenantService           api.TenantService
	UserService             api.UserService
	CourseService           api.CourseService
	ChapterService          api.CourseChapterService
	LessonService           api.CourseLessonService
	EnrollmentService       api.EnrollmentService
	ProgressService         api.ProgressService
	ResourceService         api.ResourceService
	ResourceCategoryService api.ResourceCategoryService
}

func New(
	cfg config.Config,
	dbCheck func() error,
	deps Dependencies,
) *gin.Engine {
	router := gin.New()
	router.Use(cors(), gin.Logger(), gin.Recovery())
	if cfg.StorageLocalRoot != "" {
		router.Static("/uploads", cfg.StorageLocalRoot)
	}
	router.GET("/health", middleware.Tenant(), func(c *gin.Context) {
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
	registerRoutes(router, cfg, deps)
	return router
}

func registerRoutes(
	router *gin.Engine,
	cfg config.Config,
	deps Dependencies,
) {
	authHandler := api.NewAuthHandler(deps.AuthService)
	auth := router.Group("/api/v1/auth")
	auth.Use(middleware.Tenant())
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	backend := router.Group("/backend/v1")
	backend.Use(middleware.Tenant(), middleware.Auth(cfg.JWTSecret))
	tenantHandler := api.NewTenantHandler(deps.TenantService)
	backend.POST("/tenants", tenantHandler.Create)
	backend.GET("/tenants", tenantHandler.List)
	backend.GET("/tenants/:id", tenantHandler.Get)
	backend.PUT("/tenants/:id", tenantHandler.Update)
	backend.DELETE("/tenants/:id", tenantHandler.Delete)
	userHandler := api.NewUserHandler(deps.UserService)
	backend.POST("/users", userHandler.Create)
	backend.GET("/users", userHandler.List)
	backend.GET("/users/:id", userHandler.Get)
	backend.PUT("/users/:id", userHandler.Update)
	backend.DELETE("/users/:id", userHandler.Delete)

	courseHandler := api.NewCourseHandler(deps.CourseService)
	chapterHandler := api.NewCourseChapterHandler(deps.ChapterService)
	lessonHandler := api.NewCourseLessonHandler(deps.LessonService)
	backend.POST("/courses", courseHandler.Create)
	backend.GET("/courses", courseHandler.List)
	backend.GET("/courses/:id", courseHandler.Get)
	backend.PUT("/courses/:id", courseHandler.Update)
	backend.DELETE("/courses/:id", courseHandler.Delete)
	backend.GET("/courses/:id/detail", courseHandler.Detail)
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
	resourceHandler := api.NewResourceHandler(deps.ResourceService)
	backend.POST("/resources/upload", resourceHandler.Upload)
	backend.GET("/resources", resourceHandler.List)
	backend.DELETE("/resources/:id", resourceHandler.Delete)
	categoryHandler := api.NewResourceCategoryHandler(
		deps.ResourceCategoryService,
	)
	backend.POST("/resource-categories", categoryHandler.Create)
	backend.GET("/resource-categories", categoryHandler.List)
	backend.PUT("/resource-categories/:id", categoryHandler.Update)
	backend.DELETE("/resource-categories/:id", categoryHandler.Delete)

	student := router.Group("/api/v1")
	student.Use(middleware.Tenant(), middleware.Auth(cfg.JWTSecret))
	student.GET("/courses", courseHandler.PublishedList)
	student.GET("/courses/:id", courseHandler.PublishedDetail)
	progressHandler := api.NewProgressHandler(deps.ProgressService)
	student.POST("/lessons/:id/progress", progressHandler.Report)
	student.GET("/lessons/:id/progress", progressHandler.Get)
	student.GET("/recent-learning", progressHandler.Recent)
}

func cors() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://localhost:5174": {},
		"http://localhost:5175": {},
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; ok {
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

func Run(
	cfg config.Config,
	dbCheck func() error,
	deps Dependencies,
) error {
	return New(cfg, dbCheck, deps).Run(":" + cfg.ServerPort)
}
