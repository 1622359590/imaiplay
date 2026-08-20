package main

import (
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/server"
	"github.com/1622359590/imaiplay/internal/service"
	"gorm.io/gorm"
)

type appRepositories struct {
	tenant            repository.TenantRepository
	user              repository.UserRepository
	refreshToken      repository.RefreshTokenRepository
	loginChallenge    repository.LoginChallengeRepository
	passwordReset     repository.PasswordResetRepository
	course            repository.CourseRepository
	chapter           repository.CourseChapterRepository
	lesson            repository.CourseLessonRepository
	enrollment        repository.CourseEnrollmentRepository
	progress          repository.LessonProgressRepository
	learningTime      repository.LearningTimeRepository
	learnerOverview   repository.LearnerOverviewRepository
	learnerMotivation repository.LearnerMotivationRepository
	resource          repository.ResourceRepository
	material          repository.CourseMaterialRepository
	resourceCategory  repository.ResourceCategoryRepository
	courseCategory    repository.CourseCategoryRepository
	dashboard         repository.DashboardRepository
	audit             repository.AuditLogRepository
	plan              repository.PlanRepository
	domainBindJob     repository.DomainBindJobRepository
}

func newRepositories(database *gorm.DB) appRepositories {
	return appRepositories{
		tenant: repository.NewTenantRepository(database), user: repository.NewUserRepository(database),
		refreshToken: repository.NewRefreshTokenRepository(database), loginChallenge: repository.NewLoginChallengeRepository(database),
		passwordReset: repository.NewPasswordResetRepository(database), course: repository.NewCourseRepository(database),
		chapter: repository.NewCourseChapterRepository(database), lesson: repository.NewCourseLessonRepository(database),
		enrollment: repository.NewCourseEnrollmentRepository(database), progress: repository.NewLessonProgressRepository(database),
		learningTime: repository.NewLearningTimeRepository(database), learnerOverview: repository.NewLearnerOverviewRepository(database),
		learnerMotivation: repository.NewLearnerMotivationRepository(database),
		resource:          repository.NewResourceRepository(database), material: repository.NewCourseMaterialRepository(database),
		resourceCategory: repository.NewResourceCategoryRepository(database), courseCategory: repository.NewCourseCategoryRepository(database),
		dashboard: repository.NewDashboardRepository(database), audit: repository.NewAuditLogRepository(database), plan: repository.NewPlanRepository(database),
		domainBindJob: repository.NewDomainBindJobRepository(database),
	}
}

func buildServerDependencies(cfg config.Config, database *gorm.DB, repos appRepositories, infra appInfrastructure) server.Dependencies {
	auth := service.NewAuthServiceWithRefreshTokens(repos.user, repos.tenant, repos.refreshToken, cfg.JWTSecret)
	portal := service.NewPortalService(repos.tenant, cfg.AdminHost)
	auth.SetLoginChallengeRepository(repos.loginChallenge)
	auth.SetLearnerMotivationRepository(repos.learnerMotivation)
	auth.SetPortalService(portal)
	auth.SetPasswordResetRepository(repos.passwordReset)
	auth.SetSMSSender(infra.sms.Sender())
	limits := service.NewTenantLimitService(repos.tenant, repos.plan, repos.user, repos.course)
	userService := service.NewUserService(repos.user, service.UserLimitRepositories{
		Tenants: repos.tenant,
		Plans:   repos.plan,
	})
	auth.SetEmployeeCapacityChecker(userService)
	audit := service.NewAuditService(repos.audit)
	domainBind := service.NewDomainBindService(repos.tenant, infra.domainPanel, nil, audit, service.DomainBindConfig{
		ExpectedIP: infra.expectedIP, ReservedDomain: infra.reservedDomain,
		CNAMETarget: infra.cnameTarget, ProxyTarget: infra.proxyTarget,
	}, repos.domainBindJob)
	plan := service.NewPlanService(repos.plan, repos.tenant, repos.resource)
	resource := service.NewResourceService(repos.resource, infra.storage, plan)
	learnerAccess := service.NewLearnerAccess(repos.course, repos.enrollment, repos.material)
	material := service.NewCourseMaterialService(repos.course, repos.material, repos.resource, resource).WithLearnerAccess(learnerAccess)
	return server.Dependencies{
		AuthService: auth, TenantService: service.NewTenantService(repos.tenant),
		TenantRegistrationService: service.NewTenantRegistrationService(database, cfg.JWTSecret),
		UserService:               userService,
		CourseService:             service.NewCourseService(repos.course, repos.chapter, repos.lesson, repos.enrollment, repos.material).WithCourseCategories(repos.courseCategory).WithTenantLimits(limits),
		CourseMaterialService:     material, ChapterService: service.NewCourseChapterService(repos.chapter, repos.course),
		LessonService:            service.NewCourseLessonService(repos.lesson, repos.chapter, repos.course, repos.resource),
		EnrollmentService:        service.NewEnrollmentService(repos.enrollment, repos.course, repos.user),
		ProgressService:          service.NewProgressService(repos.progress, repos.enrollment, repos.lesson, repos.chapter, repos.course, repos.learningTime),
		LearnerOverviewService:   service.NewLearnerOverviewService(repos.learnerOverview),
		LearnerMotivationService: service.NewLearnerMotivationService(repos.learnerMotivation), LearnerAccessService: learnerAccess,
		ResourceService: resource, ResourceCategoryService: service.NewResourceCategoryService(repos.resourceCategory),
		CourseCategoryService: service.NewCourseCategoryService(repos.courseCategory), DashboardService: service.NewDashboardService(repos.dashboard),
		SMSConfigService: infra.sms, AuditService: audit, TenantThemeService: service.NewTenantThemeService(repos.tenant), PlanService: plan,
		TenantRepository: repos.tenant, StorageConfigService: infra.storage, DomainBindService: domainBind, PortalService: portal,
	}
}
