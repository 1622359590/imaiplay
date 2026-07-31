package main

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	_ "github.com/1622359590/imaiplay/docs"
	"github.com/1622359590/imaiplay/internal/baota"
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/db"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/server"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/1622359590/imaiplay/internal/sms"
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
	if strings.TrimSpace(cfg.JWTSecret) == "" || cfg.JWTSecret == config.DefaultJWTSecret {
		log.Fatal("JWT_SECRET must be configured with a strong random value")
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
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)
	passwordResetRepo := repository.NewPasswordResetRepository(database)
	smsConfig, err := sms.NewConfigStore(cfg.SMSConfigFile, cfg.JWTSecret, slog.Default())
	if err != nil {
		return fmt.Errorf("initialize sms config: %w", err)
	}
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	progressRepo := repository.NewLessonProgressRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	categoryRepo := repository.NewResourceCategoryRepository(database)
	dashboardRepo := repository.NewDashboardRepository(database)
	auditRepo := repository.NewAuditLogRepository(database)
	planRepo := repository.NewPlanRepository(database)
	localStorage, err := storage.NewLocal(storage.LocalConfig{
		Root: cfg.StorageLocalRoot,
		URL:  cfg.StorageLocalURL,
	})
	if err != nil {
		return fmt.Errorf("initialize local storage: %w", err)
	}
	storageConfig, err := storage.NewConfigStore(
		cfg.StorageConfigFile,
		cfg.JWTSecret,
		storage.Config{
			Driver: cfg.StorageDriver,
			Local: storage.LocalConfig{
				Root: cfg.StorageLocalRoot,
				URL:  cfg.StorageLocalURL,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("initialize storage config: %w", err)
	}
	runtimeStorage, err := storage.NewRuntime(localStorage, storageConfig, cfg.StorageDriver)
	if err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}
	authService := service.NewAuthServiceWithRefreshTokens(userRepo, tenantRepo, refreshTokenRepo, cfg.JWTSecret)
	authService.SetPasswordResetRepository(passwordResetRepo)
	authService.SetSMSSender(smsConfig.Sender())
	auditService := service.NewAuditService(auditRepo)
	var domainPanel service.DomainPanel
	if strings.TrimSpace(cfg.BaotaPanelURL) != "" && strings.TrimSpace(cfg.BaotaAPIKey) != "" {
		domainPanel = &baota.Client{
			PanelURL: strings.TrimSpace(cfg.BaotaPanelURL),
			APIKey:   strings.TrimSpace(cfg.BaotaAPIKey),
		}
	}
	domainBindService := service.NewDomainBindService(
		tenantRepo,
		domainPanel,
		nil,
		auditService,
		service.DomainBindConfig{
			ExpectedIP:     strings.TrimSpace(cfg.BaotaServerIP),
			ReservedDomain: strings.TrimSpace(cfg.AdminHost),
			CNAMETarget:    strings.TrimSpace(cfg.AdminHost),
			ProxyTarget:    strings.TrimSpace(cfg.BaotaProxyTarget),
		},
	)
	deps := server.Dependencies{
		AuthService:               authService,
		TenantService:             service.NewTenantService(tenantRepo),
		TenantRegistrationService: service.NewTenantRegistrationService(database, cfg.JWTSecret),
		UserService:               service.NewUserService(userRepo),
		CourseService:             service.NewCourseService(courseRepo, chapterRepo, lessonRepo),
		ChapterService:            service.NewCourseChapterService(chapterRepo, courseRepo),
		LessonService: service.NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo, resourceRepo,
		),
		EnrollmentService: service.NewEnrollmentService(
			enrollmentRepo, courseRepo, userRepo,
		),
		ProgressService: service.NewProgressService(
			progressRepo, enrollmentRepo, lessonRepo, chapterRepo, courseRepo,
		),
		ResourceService:         service.NewResourceService(resourceRepo, runtimeStorage, service.NewPlanService(planRepo, tenantRepo, resourceRepo)),
		ResourceCategoryService: service.NewResourceCategoryService(categoryRepo),
		DashboardService:        service.NewDashboardService(dashboardRepo),
		SMSConfigService:        smsConfig,
		AuditService:            auditService,
		TenantThemeService:      service.NewTenantThemeService(tenantRepo),
		PlanService:             service.NewPlanService(planRepo, tenantRepo, resourceRepo),
		TenantRepository:        tenantRepo,
		StorageConfigService:    runtimeStorage,
		DomainBindService:       domainBindService,
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
