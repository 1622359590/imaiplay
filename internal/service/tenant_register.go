package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
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
	return service.registerWithOptions(ctx, organizationName, adminEmail, phone, adminName, password, "", true)
}

func (service *TenantRegistrationService) CreateForSuperadmin(
	ctx context.Context, organizationName, adminEmail, phone, adminName, password, planID string,
) (*TenantRegistrationResult, error) {
	if err := requireRole(ctx, "superadmin"); err != nil {
		return nil, err
	}
	return service.registerWithOptions(ctx, organizationName, adminEmail, phone, adminName, password, planID, false)
}

func (service *TenantRegistrationService) registerWithOptions(
	ctx context.Context, organizationName, adminEmail, phone, adminName, password, planID string, issueToken bool,
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
		trialEndsAt := time.Now().UTC().Add(14 * 24 * time.Hour)
		tenant := &domain.Tenant{Code: code, Name: organizationName, Status: 1, LifecycleStatus: "trial", TrialEndsAt: &trialEndsAt}
		if err := tx.Create(tenant).Error; err != nil {
			return mapCreateError(err, "tenant code already exists", "create tenant failed")
		}
		var selectedPlan domain.Plan
		planQuery := tx.Where("is_default = ? AND status = ?", true, 1)
		if planID != "" {
			planQuery = tx.Where("id = ? AND status = ?", planID, 1)
		}
		if err := planQuery.First(&selectedPlan).Error; err == nil {
			tenant.PlanID = &selectedPlan.ID
			if err := tx.Model(tenant).Update("plan_id", tenant.PlanID).Error; err != nil {
				return err
			}
		} else if planID != "" {
			return errorsx.BadRequest("invalid plan")
		}
		admin := &domain.User{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Email: adminEmail, Phone: nullablePhone(phone), Password: hash, Name: adminName, Role: "tenant_admin", Status: 1}
		if err := tx.Create(admin).Error; err != nil {
			return mapCreateError(err, "email already exists", "create admin failed")
		}
		if err := seedDemoData(ctx, tx, tenant.ID, admin.ID); err != nil {
			return errorsx.Internal("create demo data failed")
		}
		token := ""
		if issueToken {
			var err error
			token, err = security.GenerateToken(admin.ID, tenant.ID, admin.Email, admin.Role, service.jwtSecret)
			if err != nil {
				return errorsx.Internal("generate token failed")
			}
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
		records := repository.NewTenantDemoRecordRepository(tx)
		registered, err := records.ListByTenant(ctx, tenantID)
		if err != nil {
			return err
		}
		ids := make(map[string][]string)
		batches := make(map[string]struct{})
		for _, record := range registered {
			ids[record.RecordType] = append(ids[record.RecordType], record.RecordID)
			batches[record.BatchID] = struct{}{}
		}
		if err := validateRegisteredDemoDependencies(tx, tenantID, ids); err != nil {
			return err
		}
		if err := clearRegisteredDemoRecords(tx, tenantID, ids); err != nil {
			return err
		}
		for batchID := range batches {
			if err := records.DeleteBatch(ctx, tenantID, batchID); err != nil {
				return err
			}
		}
		return nil
	})
}

func validateRegisteredDemoDependencies(
	tx *gorm.DB, tenantID string, ids map[string][]string,
) error {
	courseIDs := ids[repository.DemoRecordCourse]
	chapterIDs := ids[repository.DemoRecordCourseChapter]
	lessonIDs := ids[repository.DemoRecordCourseLesson]
	userIDs := ids[repository.DemoRecordUser]
	resourceIDs := ids[repository.DemoRecordResource]

	for _, reference := range []struct {
		model         interface{}
		field         string
		parentIDs     []string
		registeredIDs []string
	}{
		{&domain.CourseChapter{}, "course_id", courseIDs, chapterIDs},
		{&domain.CourseLesson{}, "chapter_id", chapterIDs, lessonIDs},
		{&domain.CourseLesson{}, "resource_id", resourceIDs, lessonIDs},
	} {
		if err := rejectUnregisteredDemoReferences(
			tx, tenantID, reference.model, reference.field,
			reference.parentIDs, reference.registeredIDs,
		); err != nil {
			return err
		}
	}

	for _, reference := range []struct {
		field string
		ids   []string
	}{
		{"course_id", courseIDs},
		{"resource_id", resourceIDs},
	} {
		if err := rejectDemoReferences(
			tx, tenantID, &domain.CourseMaterial{}, reference.field, reference.ids,
		); err != nil {
			return err
		}
	}

	for _, relation := range []struct {
		model                 interface{}
		leftField, rightField string
		leftIDs, rightIDs     []string
	}{
		{&domain.CourseEnrollment{}, "course_id", "user_id", courseIDs, userIDs},
		{&domain.LessonProgress{}, "lesson_id", "user_id", lessonIDs, userIDs},
		{&domain.LearningTimeReport{}, "lesson_id", "user_id", lessonIDs, userIDs},
	} {
		if err := rejectPartialDemoRelation(
			tx, tenantID, relation.model,
			relation.leftField, relation.rightField,
			relation.leftIDs, relation.rightIDs,
		); err != nil {
			return err
		}
	}
	return nil
}

func rejectUnregisteredDemoReferences(
	tx *gorm.DB,
	tenantID string,
	model interface{},
	field string,
	parentIDs, registeredIDs []string,
) error {
	if len(parentIDs) == 0 {
		return nil
	}
	query := tx.Model(model).Where(
		"tenant_id = ? AND "+field+" IN ?", tenantID, parentIDs,
	)
	if len(registeredIDs) > 0 {
		query = query.Where("id NOT IN ?", registeredIDs)
	}
	return rejectExistingDemoDependency(query)
}

func rejectDemoReferences(
	tx *gorm.DB, tenantID string, model interface{}, field string, ids []string,
) error {
	if len(ids) == 0 {
		return nil
	}
	return rejectExistingDemoDependency(tx.Model(model).Where(
		"tenant_id = ? AND "+field+" IN ?", tenantID, ids,
	))
}

func rejectPartialDemoRelation(
	tx *gorm.DB,
	tenantID string,
	model interface{},
	leftField, rightField string,
	leftIDs, rightIDs []string,
) error {
	if len(leftIDs) > 0 {
		query := tx.Model(model).Where(
			"tenant_id = ? AND "+leftField+" IN ?", tenantID, leftIDs,
		)
		if len(rightIDs) > 0 {
			query = query.Where(rightField+" NOT IN ?", rightIDs)
		}
		if err := rejectExistingDemoDependency(query); err != nil {
			return err
		}
	}
	if len(rightIDs) == 0 {
		return nil
	}
	query := tx.Model(model).Where(
		"tenant_id = ? AND "+rightField+" IN ?", tenantID, rightIDs,
	)
	if len(leftIDs) > 0 {
		query = query.Where(leftField+" NOT IN ?", leftIDs)
	}
	return rejectExistingDemoDependency(query)
}

func rejectExistingDemoDependency(query *gorm.DB) error {
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errorsx.Conflict("demo data has unregistered dependencies")
	}
	return nil
}

func clearRegisteredDemoRecords(
	tx *gorm.DB, tenantID string, ids map[string][]string,
) error {
	courseIDs := ids[repository.DemoRecordCourse]
	chapterIDs := ids[repository.DemoRecordCourseChapter]
	lessonIDs := ids[repository.DemoRecordCourseLesson]
	userIDs := ids[repository.DemoRecordUser]
	resourceIDs := ids[repository.DemoRecordResource]
	for _, deletion := range []struct {
		model interface{}
		field string
		ids   []string
	}{
		{&domain.LessonProgress{}, "lesson_id", lessonIDs},
		{&domain.LessonProgress{}, "user_id", userIDs},
		{&domain.LearningTimeReport{}, "lesson_id", lessonIDs},
		{&domain.LearningTimeReport{}, "user_id", userIDs},
		{&domain.LearningDailyStat{}, "user_id", userIDs},
		{&domain.RefreshToken{}, "user_id", userIDs},
		{&domain.CourseEnrollment{}, "course_id", courseIDs},
		{&domain.CourseEnrollment{}, "user_id", userIDs},
	} {
		if len(deletion.ids) == 0 {
			continue
		}
		if err := tx.Where(
			"tenant_id = ? AND "+deletion.field+" IN ?", tenantID, deletion.ids,
		).Delete(deletion.model).Error; err != nil {
			return err
		}
	}
	for _, deletion := range []struct {
		model interface{}
		ids   []string
	}{
		{&domain.CourseLesson{}, lessonIDs},
		{&domain.CourseChapter{}, chapterIDs},
		{&domain.Course{}, courseIDs},
		{&domain.Resource{}, resourceIDs},
		{&domain.User{}, userIDs},
	} {
		if len(deletion.ids) == 0 {
			continue
		}
		if err := tx.Where(
			"tenant_id = ? AND id IN ?", tenantID, deletion.ids,
		).Delete(deletion.model).Error; err != nil {
			return err
		}
	}
	return nil
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

func seedDemoData(ctx context.Context, tx *gorm.DB, tenantID, adminID string) error {
	hash, err := security.HashPassword("demo1234")
	if err != nil {
		return err
	}
	batchID := uuid.NewString()
	records := make([]domain.TenantDemoRecord, 0, 12)
	register := func(recordType, recordID string) {
		records = append(records, domain.TenantDemoRecord{
			BaseModel: domain.BaseModel{TenantID: tenantID},
			BatchID:   batchID, RecordType: recordType, RecordID: recordID,
		})
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
		register(repository.DemoRecordUser, user.ID)
	}
	course := &domain.Course{BaseModel: domain.BaseModel{TenantID: tenantID}, Title: demoCourseTitle, Description: "ImaiPlay 示例课程", Status: 1, CreatedBy: adminID}
	if err := tx.Create(course).Error; err != nil {
		return err
	}
	register(repository.DemoRecordCourse, course.ID)
	chapters := []*domain.CourseChapter{
		{BaseModel: domain.BaseModel{TenantID: tenantID}, CourseID: course.ID, Title: "第一章：公司介绍", SortOrder: 1},
		{BaseModel: domain.BaseModel{TenantID: tenantID}, CourseID: course.ID, Title: "第二章：规章制度", SortOrder: 2},
	}
	for _, chapter := range chapters {
		if err := tx.Create(chapter).Error; err != nil {
			return err
		}
		register(repository.DemoRecordCourseChapter, chapter.ID)
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
		register(repository.DemoRecordCourseLesson, lesson.ID)
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
		register(repository.DemoRecordResource, resource.ID)
	}
	return repository.NewTenantDemoRecordRepository(tx).RegisterBatch(ctx, records)
}
