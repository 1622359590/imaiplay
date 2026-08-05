package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestClearDemoDataDeletesRefreshTokensForRegisteredUsers(t *testing.T) {
	database, registration, result := registeredTenantForReview(t)
	userID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordUser)
	registeredToken := &domain.RefreshToken{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		UserID:    userID, TokenHash: "registered-demo-user-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	mustCreateReview(t, database, registeredToken)
	businessUser := createBusinessUser(t, database, result, "business-token@example.com")
	businessToken := &domain.RefreshToken{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		UserID:    businessUser.ID, TokenHash: "business-user-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	mustCreateReview(t, database, businessToken)
	foreignToken := &domain.RefreshToken{
		BaseModel: domain.BaseModel{TenantID: "foreign-tenant"},
		UserID:    userID, TokenHash: "foreign-tenant-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	mustCreateReview(t, database, foreignToken)
	ctx := usercontext.WithUser(
		context.Background(), result.User.ID, result.Tenant.ID,
		result.User.Email, "tenant_admin",
	)
	if err := registration.ClearDemoData(ctx); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := database.Model(&domain.RefreshToken{}).
		Where("tenant_id = ? AND id = ?", result.Tenant.ID, registeredToken.ID).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("registered demo user refresh token count=%d, want 0", count)
	}
	for _, token := range []*domain.RefreshToken{businessToken, foreignToken} {
		if err := database.Model(&domain.RefreshToken{}).
			Where("tenant_id = ? AND id = ?", token.TenantID, token.ID).
			Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("preserved refresh token %s count=%d, want 1", token.ID, count)
		}
	}
}

func TestClearDemoDataRollsBackRefreshTokenDeletion(t *testing.T) {
	database, registration, result := registeredTenantForReview(t)
	userID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordUser)
	token := &domain.RefreshToken{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		UserID:    userID, TokenHash: "rollback-demo-user-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	mustCreateReview(t, database, token)
	if err := database.Callback().Delete().Before("gorm:delete").Register(
		"fail_review_demo_record_delete",
		func(tx *gorm.DB) {
			if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "tenant_demo_records" {
				tx.AddError(errors.New("demo registry delete failed"))
			}
		},
	); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(
		context.Background(), result.User.ID, result.Tenant.ID,
		result.User.Email, "tenant_admin",
	)
	if err := registration.ClearDemoData(ctx); err == nil {
		t.Fatal("ClearDemoData() error=nil, want registry deletion failure")
	}
	var count int64
	if err := database.Model(&domain.RefreshToken{}).
		Where("tenant_id = ? AND id = ?", result.Tenant.ID, token.ID).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("refresh token count after rollback=%d, want 1", count)
	}
}

func TestClearDemoDataRejectsUnregisteredStructuralDependents(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *gorm.DB, *TenantRegistrationResult)
	}{
		{
			name: "chapter under registered course",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				courseID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordCourse)
				mustCreateReview(t, database, &domain.CourseChapter{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					CourseID:  courseID, Title: "Business chapter",
				})
			},
		},
		{
			name: "lesson under registered chapter",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				chapterID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordCourseChapter)
				mustCreateReview(t, database, &domain.CourseLesson{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					ChapterID: chapterID, Title: "Business lesson", ContentType: "text",
				})
			},
		},
		{
			name: "business lesson referencing registered resource",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				_, chapter, lesson := createBusinessLesson(t, database, result, "Business resource lesson")
				var resource domain.Resource
				if err := database.Where(
					"tenant_id = ? AND resource_type = ?", result.Tenant.ID, "document",
				).First(&resource).Error; err != nil {
					t.Fatal(err)
				}
				lesson.ChapterID = chapter.ID
				lesson.ResourceID = &resource.ID
				mustCreateReview(t, database, lesson)
			},
		},
		{
			name: "course material under registered course",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				courseID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordCourse)
				resource := createBusinessResource(t, database, result, "business-course-material.pdf")
				mustCreateReview(t, database, &domain.CourseMaterial{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					CourseID:  courseID, ResourceID: resource.ID,
					DisplayName: "Business course material", CreatedBy: result.User.ID,
				})
			},
		},
		{
			name: "business course material referencing registered resource",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				course, _, _ := createBusinessLesson(t, database, result, "Material owner")
				resourceID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordResource)
				mustCreateReview(t, database, &domain.CourseMaterial{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					CourseID:  course.ID, ResourceID: resourceID,
					DisplayName: "Business resource material", CreatedBy: result.User.ID,
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database, registration, result := registeredTenantForReview(t)
			test.setup(t, database, result)
			assertClearDemoConflictWithoutDeletes(t, database, registration, result)
		})
	}
}

func TestClearDemoDataRejectsUnregisteredActivityLinksToSurvivingRecords(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *gorm.DB, *TenantRegistrationResult)
	}{
		{
			name: "business enrollment in registered course",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				courseID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordCourse)
				user := createBusinessUser(t, database, result, "enrollment-course@example.com")
				mustCreateReview(t, database, &domain.CourseEnrollment{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					CourseID:  courseID, UserID: user.ID, Status: 1,
				})
			},
		},
		{
			name: "registered user enrolled in business course",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				course, _, _ := createBusinessLesson(t, database, result, "Enrollment owner")
				userID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordUser)
				mustCreateReview(t, database, &domain.CourseEnrollment{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					CourseID:  course.ID, UserID: userID, Status: 1,
				})
			},
		},
		{
			name: "business user progress in registered lesson",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				lessonID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordCourseLesson)
				user := createBusinessUser(t, database, result, "progress-lesson@example.com")
				mustCreateReview(t, database, &domain.LessonProgress{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					UserID:    user.ID, LessonID: lessonID,
				})
			},
		},
		{
			name: "registered user progress in business lesson",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				_, _, lesson := createBusinessLesson(t, database, result, "Progress owner")
				mustCreateReview(t, database, lesson)
				userID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordUser)
				mustCreateReview(t, database, &domain.LessonProgress{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					UserID:    userID, LessonID: lesson.ID,
				})
			},
		},
		{
			name: "business user time report for registered lesson",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				lessonID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordCourseLesson)
				user := createBusinessUser(t, database, result, "report-lesson@example.com")
				mustCreateReview(t, database, &domain.LearningTimeReport{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					UserID:    user.ID, LessonID: lessonID,
					ReportID: "business-report", WatchedSecondsDelta: 15,
				})
			},
		},
		{
			name: "registered user time report for business lesson",
			setup: func(t *testing.T, database *gorm.DB, result *TenantRegistrationResult) {
				_, _, lesson := createBusinessLesson(t, database, result, "Report owner")
				mustCreateReview(t, database, lesson)
				userID := registeredDemoID(t, database, result.Tenant.ID, repository.DemoRecordUser)
				mustCreateReview(t, database, &domain.LearningTimeReport{
					BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
					UserID:    userID, LessonID: lesson.ID,
					ReportID: "business-report", WatchedSecondsDelta: 15,
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database, registration, result := registeredTenantForReview(t)
			test.setup(t, database, result)
			assertClearDemoConflictWithoutDeletes(t, database, registration, result)
		})
	}
}

func registeredTenantForReview(
	t *testing.T,
) (*gorm.DB, *TenantRegistrationService, *TenantRegistrationResult) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	registration := NewTenantRegistrationService(database, "secret")
	result, err := registration.Register(
		context.Background(), "Review tenant", "admin@review.test", "Admin", "password123",
	)
	if err != nil {
		t.Fatal(err)
	}
	return database, registration, result
}

func assertClearDemoConflictWithoutDeletes(
	t *testing.T,
	database *gorm.DB,
	registration *TenantRegistrationService,
	result *TenantRegistrationResult,
) {
	t.Helper()
	before := tenantModelCounts(t, database, result.Tenant.ID)
	deleteCalls := 0
	if err := database.Callback().Delete().Before("gorm:delete").Register(
		"detect_demo_delete_before_preflight",
		func(*gorm.DB) { deleteCalls++ },
	); err != nil {
		t.Fatal(err)
	}
	ctx := usercontext.WithUser(
		context.Background(), result.User.ID, result.Tenant.ID,
		result.User.Email, "tenant_admin",
	)
	if err := registration.ClearDemoData(ctx); errorCode(err) != 40900 {
		t.Fatalf("ClearDemoData() error=%#v, want conflict", err)
	}
	if deleteCalls != 0 {
		t.Fatalf("ClearDemoData() issued %d delete operations before rejecting", deleteCalls)
	}
	after := tenantModelCounts(t, database, result.Tenant.ID)
	for model, want := range before {
		if after[model] != want {
			t.Fatalf("%s count after conflict=%d, want %d", model, after[model], want)
		}
	}
}

func tenantModelCounts(t *testing.T, database *gorm.DB, tenantID string) map[string]int64 {
	t.Helper()
	models := map[string]interface{}{
		"courses":               &domain.Course{},
		"course_chapters":       &domain.CourseChapter{},
		"course_lessons":        &domain.CourseLesson{},
		"course_materials":      &domain.CourseMaterial{},
		"course_enrollments":    &domain.CourseEnrollment{},
		"lesson_progress":       &domain.LessonProgress{},
		"learning_time_reports": &domain.LearningTimeReport{},
		"learning_daily_stats":  &domain.LearningDailyStat{},
		"resources":             &domain.Resource{},
		"users":                 &domain.User{},
		"refresh_tokens":        &domain.RefreshToken{},
		"tenant_demo_records":   &domain.TenantDemoRecord{},
	}
	counts := make(map[string]int64, len(models))
	for name, model := range models {
		var count int64
		if err := database.Model(model).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		counts[name] = count
	}
	return counts
}

func registeredDemoID(
	t *testing.T, database *gorm.DB, tenantID, recordType string,
) string {
	t.Helper()
	var record domain.TenantDemoRecord
	if err := database.Where(
		"tenant_id = ? AND record_type = ?", tenantID, recordType,
	).First(&record).Error; err != nil {
		t.Fatal(err)
	}
	return record.RecordID
}

func createBusinessLesson(
	t *testing.T, database *gorm.DB, result *TenantRegistrationResult, title string,
) (*domain.Course, *domain.CourseChapter, *domain.CourseLesson) {
	t.Helper()
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		Title:     title, CreatedBy: result.User.ID,
	}
	mustCreateReview(t, database, course)
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		CourseID:  course.ID, Title: title,
	}
	mustCreateReview(t, database, chapter)
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		ChapterID: chapter.ID, Title: title, ContentType: "text",
	}
	return course, chapter, lesson
}

func createBusinessUser(
	t *testing.T, database *gorm.DB, result *TenantRegistrationResult, email string,
) *domain.User {
	t.Helper()
	user := &domain.User{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		Email:     email, Password: "business", Name: "Business user", Role: "learner", Status: 1,
	}
	mustCreateReview(t, database, user)
	return user
}

func createBusinessResource(
	t *testing.T, database *gorm.DB, result *TenantRegistrationResult, name string,
) *domain.Resource {
	t.Helper()
	resource := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: result.Tenant.ID},
		Name:      name, ResourceType: "document", URL: "/business/" + name,
		CreatedBy: result.User.ID,
	}
	mustCreateReview(t, database, resource)
	return resource
}

func mustCreateReview(t *testing.T, database *gorm.DB, value interface{}) {
	t.Helper()
	if err := database.Create(value).Error; err != nil {
		t.Fatal(fmt.Errorf("create review fixture: %w", err))
	}
}
