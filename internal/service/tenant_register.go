package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	demoCourseTitle  = "新员工入职培训"
	demoImageName    = "示例图片"
	demoDocumentName = "示例文档"
)

type TenantRegistrationResult struct {
	Tenant *domain.Tenant `json:"tenant"`
	User   *domain.User   `json:"user"`
	Token  string         `json:"token"`
}

type TenantRegistrationService struct {
	database  *gorm.DB
	jwtSecret string
}

func NewTenantRegistrationService(database *gorm.DB, jwtSecret string) *TenantRegistrationService {
	return &TenantRegistrationService{database: database, jwtSecret: jwtSecret}
}

func (service *TenantRegistrationService) Register(
	ctx context.Context, organizationName, adminEmail, adminName, password string,
) (*TenantRegistrationResult, error) {
	return service.RegisterWithPhone(ctx, organizationName, adminEmail, "", adminName, password)
}

func (service *TenantRegistrationService) RegisterWithPhone(
	ctx context.Context, organizationName, adminEmail, phone, adminName, password string,
) (*TenantRegistrationResult, error) {
	organizationName = strings.TrimSpace(organizationName)
	adminEmail = strings.ToLower(strings.TrimSpace(adminEmail))
	phone = normalizePhone(phone)
	adminName = strings.TrimSpace(adminName)
	if organizationName == "" {
		return nil, errorsx.BadRequest("organization name is required")
	}
	if adminEmail == "" || adminName == "" {
		return nil, errorsx.BadRequest("admin email and name are required")
	}
	if len(password) < 8 {
		return nil, errorsx.BadRequest("password must be at least 8 characters")
	}
	if phone != "" && !validPhone(phone) {
		return nil, errorsx.BadRequest("invalid phone")
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, errorsx.Internal("hash password failed")
	}

	var result TenantRegistrationResult
	err = service.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code, err := uniqueTenantCode(tx, tenantCodeSlug(organizationName))
		if err != nil {
			return errorsx.Internal("generate tenant code failed")
		}
		tenant := &domain.Tenant{Code: code, Name: organizationName, Status: 1}
		if err := tx.Create(tenant).Error; err != nil {
			return mapCreateError(err, "tenant code already exists", "create tenant failed")
		}
		var defaultPlan domain.Plan
		if err := tx.Where("is_default = ? AND status = ?", true, 1).First(&defaultPlan).Error; err == nil {
			tenant.PlanID = &defaultPlan.ID
			if err := tx.Model(tenant).Update("plan_id", tenant.PlanID).Error; err != nil {
				return err
			}
		}
		admin := &domain.User{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Email: adminEmail, Phone: nullablePhone(phone), Password: hash, Name: adminName, Role: "tenant_admin", Status: 1}
		if err := tx.Create(admin).Error; err != nil {
			return mapCreateError(err, "email already exists", "create admin failed")
		}
		if err := seedDemoData(tx, tenant.ID, admin.ID); err != nil {
			return errorsx.Internal("create demo data failed")
		}
		token, err := security.GenerateToken(admin.ID, tenant.ID, admin.Email, admin.Role, service.jwtSecret)
		if err != nil {
			return errorsx.Internal("generate token failed")
		}
		result = TenantRegistrationResult{Tenant: tenant, User: admin, Token: token}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (service *TenantRegistrationService) ClearDemoData(ctx context.Context) error {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" || role != "tenant_admin" {
		return errorsx.Forbidden("permission denied")
	}
	return service.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var courseIDs, chapterIDs, lessonIDs []string
		if err := tx.Model(&domain.Course{}).Where("tenant_id = ? AND title = ?", tenantID, demoCourseTitle).Pluck("id", &courseIDs).Error; err != nil {
			return err
		}
		if len(courseIDs) > 0 {
			if err := tx.Model(&domain.CourseChapter{}).Where("tenant_id = ? AND course_id IN ?", tenantID, courseIDs).Pluck("id", &chapterIDs).Error; err != nil {
				return err
			}
		}
		if len(chapterIDs) > 0 {
			if err := tx.Model(&domain.CourseLesson{}).Where("tenant_id = ? AND chapter_id IN ?", tenantID, chapterIDs).Pluck("id", &lessonIDs).Error; err != nil {
				return err
			}
		}
		if len(lessonIDs) > 0 {
			if err := tx.Where("tenant_id = ? AND lesson_id IN ?", tenantID, lessonIDs).Delete(&domain.LessonProgress{}).Error; err != nil {
				return err
			}
		}
		if len(courseIDs) > 0 {
			if err := tx.Where("tenant_id = ? AND course_id IN ?", tenantID, courseIDs).Delete(&domain.CourseEnrollment{}).Error; err != nil {
				return err
			}
		}
		if len(lessonIDs) > 0 {
			if err := tx.Where("tenant_id = ? AND id IN ?", tenantID, lessonIDs).Delete(&domain.CourseLesson{}).Error; err != nil {
				return err
			}
		}
		if len(chapterIDs) > 0 {
			if err := tx.Where("tenant_id = ? AND id IN ?", tenantID, chapterIDs).Delete(&domain.CourseChapter{}).Error; err != nil {
				return err
			}
		}
		if len(courseIDs) > 0 {
			if err := tx.Where("tenant_id = ? AND id IN ?", tenantID, courseIDs).Delete(&domain.Course{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("tenant_id = ? AND email IN ?", tenantID, []string{"learner1@example.com", "learner2@example.com", "instructor@example.com"}).Delete(&domain.User{}).Error; err != nil {
			return err
		}
		return tx.Where("tenant_id = ? AND name IN ?", tenantID, []string{demoImageName, demoDocumentName}).Delete(&domain.Resource{}).Error
	})
}

func tenantCodeSlug(name string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			if builder.Len() > 0 && !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	base := strings.Trim(builder.String(), "-")
	if base == "" {
		return "t-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	}
	if len(base) > 32 {
		base = strings.Trim(base[:32], "-")
	}
	return base
}

func uniqueTenantCode(tx *gorm.DB, base string) (string, error) {
	for i := 0; i < 1000; i++ {
		candidate := base
		if i > 0 {
			suffix := fmt.Sprintf("-%d", i)
			prefixLength := 32 - len(suffix)
			if prefixLength < 1 {
				return "", errors.New("tenant code suffix too long")
			}
			prefix := base
			if len(prefix) > prefixLength {
				prefix = strings.TrimRight(prefix[:prefixLength], "-")
			}
			candidate = prefix + suffix
		}
		var count int64
		if err := tx.Model(&domain.Tenant{}).Where("code = ?", candidate).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", errors.New("tenant code space exhausted")
}

func seedDemoData(tx *gorm.DB, tenantID, adminID string) error {
	hash, err := security.HashPassword("demo1234")
	if err != nil {
		return err
	}
	users := []*domain.User{
		{BaseModel: domain.BaseModel{TenantID: tenantID}, Email: "learner1@example.com", Password: hash, Name: "示例学员 1", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, Email: "learner2@example.com", Password: hash, Name: "示例学员 2", Role: "learner", Status: 1},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, Email: "instructor@example.com", Password: hash, Name: "示例讲师", Role: "instructor", Status: 1},
	}
	for _, user := range users {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
	}
	course := &domain.Course{BaseModel: domain.BaseModel{TenantID: tenantID}, Title: demoCourseTitle, Description: "ImaiPlay 示例课程", Status: 1, CreatedBy: adminID}
	if err := tx.Create(course).Error; err != nil {
		return err
	}
	chapters := []*domain.CourseChapter{
		{BaseModel: domain.BaseModel{TenantID: tenantID}, CourseID: course.ID, Title: "第一章：公司介绍", SortOrder: 1},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, CourseID: course.ID, Title: "第二章：规章制度", SortOrder: 2},
	}
	for _, chapter := range chapters {
		if err := tx.Create(chapter).Error; err != nil {
			return err
		}
	}
	lessons := []*domain.CourseLesson{
		{BaseModel: domain.BaseModel{TenantID: tenantID}, ChapterID: chapters[0].ID, Title: "欢迎视频", ContentType: "video", ContentURL: "https://example.com/demo-welcome.mp4", SortOrder: 1},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, ChapterID: chapters[0].ID, Title: "企业文化手册", ContentType: "document", ContentURL: "https://example.com/demo-handbook.pdf", SortOrder: 2},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, ChapterID: chapters[1].ID, Title: "考勤制度", ContentType: "document", ContentURL: "https://example.com/demo-attendance.pdf", SortOrder: 1},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, ChapterID: chapters[1].ID, Title: "安全须知", ContentType: "text", ContentURL: "请遵守公司安全规定。", SortOrder: 2},
	}
	for _, lesson := range lessons {
		if err := tx.Create(lesson).Error; err != nil {
			return err
		}
	}
	for _, user := range users[:2] {
		if err := tx.Create(&domain.CourseEnrollment{BaseModel: domain.BaseModel{TenantID: tenantID}, CourseID: course.ID, UserID: user.ID, Status: 1}).Error; err != nil {
			return err
		}
	}
	resources := []*domain.Resource{
		{BaseModel: domain.BaseModel{TenantID: tenantID}, Name: demoImageName, ResourceType: "image", URL: "/uploads/demo-image.png", SizeBytes: 68, CreatedBy: adminID},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, Name: demoDocumentName, ResourceType: "document", URL: "/uploads/demo-document.pdf", SizeBytes: 0, CreatedBy: adminID},
	}
	for _, resource := range resources {
		if err := tx.Create(resource).Error; err != nil {
			return err
		}
	}
	return nil
}
