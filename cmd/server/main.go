package main

import (
	"fmt"
	"log"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/db"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/server"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/1622359590/imaiplay/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	database, err := db.New(cfg)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		if err := db.Close(database); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := migration.AutoMigrate(database); err != nil {
		return fmt.Errorf("migrate database: %w", err)
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
	dashboardRepo := repository.NewDashboardRepository(database)
	if cfg.StorageDriver != "local" {
		return fmt.Errorf("unsupported storage driver: %s", cfg.StorageDriver)
	}
	localStorage, err := storage.NewLocal(storage.LocalConfig{
		Root: cfg.StorageLocalRoot,
		URL:  cfg.StorageLocalURL,
	})
	if err != nil {
		return fmt.Errorf("initialize local storage: %w", err)
	}
	deps := server.Dependencies{
		AuthService:    service.NewAuthService(userRepo, tenantRepo, cfg.JWTSecret),
		TenantService:  service.NewTenantService(tenantRepo),
		UserService:    service.NewUserService(userRepo),
		CourseService:  service.NewCourseService(courseRepo, chapterRepo, lessonRepo),
		ChapterService: service.NewCourseChapterService(chapterRepo, courseRepo),
		LessonService: service.NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo,
		),
		EnrollmentService: service.NewEnrollmentService(
			enrollmentRepo, courseRepo, userRepo,
		),
		ProgressService: service.NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
		),
		ResourceService:         service.NewResourceService(resourceRepo, localStorage),
		ResourceCategoryService: service.NewResourceCategoryService(categoryRepo),
		DashboardService:        service.NewDashboardService(dashboardRepo),
	}
	if err := server.Run(
		cfg,
		func() error { return db.Ping(database) },
		deps,
	); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
