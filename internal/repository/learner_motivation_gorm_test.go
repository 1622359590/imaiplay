package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
)

func TestLearnerMotivationRepositoryMarksFirstLoginOnce(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	createRepositoryTenant(t, database, "tenant-first-login")
	user := newTestUser("tenant-first-login", "first-login@example.com", "First Login")
	if err := NewUserRepository(database).Create(context.Background(), user); err != nil {
		t.Fatalf("create learner: %v", err)
	}

	repository := NewLearnerMotivationRepository(database)
	ctx := usercontext.WithUser(context.Background(), user.ID, user.TenantID, user.Email, user.Role)
	firstAt := time.Date(2026, time.August, 20, 2, 30, 0, 0, time.UTC)
	changed, err := repository.MarkFirstLogin(ctx, user.ID, firstAt)
	if err != nil || !changed {
		t.Fatalf("first MarkFirstLogin() = %v, %v, want true, nil", changed, err)
	}
	changed, err = repository.MarkFirstLogin(ctx, user.ID, firstAt.Add(time.Hour))
	if err != nil || changed {
		t.Fatalf("second MarkFirstLogin() = %v, %v, want false, nil", changed, err)
	}

	var state domain.LearnerEngagementState
	if err := database.Where("tenant_id = ? AND user_id = ?", user.TenantID, user.ID).First(&state).Error; err != nil {
		t.Fatalf("load engagement state: %v", err)
	}
	if state.FirstLoginAt == nil || !state.FirstLoginAt.Equal(firstAt) {
		t.Fatalf("first_login_at = %v, want %v", state.FirstLoginAt, firstAt)
	}
	if _, err := repository.MarkFirstLogin(context.Background(), user.ID, firstAt); err == nil {
		t.Fatal("MarkFirstLogin() accepted a missing learner identity")
	}
}

func TestLearnerMotivationRepositoryLoadsTenantScopedSnapshot(t *testing.T) {
	database := learningTimeDatabase(t)
	for _, tenant := range []*domain.Tenant{
		{ID: "motivation-tenant", Code: "motivation-tenant", Name: "Motivation", Status: 1},
		{ID: "motivation-foreign", Code: "motivation-foreign", Name: "Foreign", Status: 1},
	} {
		if err := database.Create(tenant).Error; err != nil {
			t.Fatalf("create tenant: %v", err)
		}
	}
	learner := &domain.User{
		BaseModel: domain.BaseModel{ID: "motivation-learner", TenantID: "motivation-tenant"},
		Email:     "motivation-learner@example.com", Password: "hash", Name: "Learner", Role: "learner", Status: 1,
	}
	if err := NewUserRepository(database).Create(context.Background(), learner); err != nil {
		t.Fatalf("create learner: %v", err)
	}
	loginAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if err := database.Model(&domain.LearnerEngagementState{}).
		Where("user_id = ?", learner.ID).
		Updates(map[string]interface{}{"first_login_at": loginAt, "welcome_seen_at": loginAt}).Error; err != nil {
		t.Fatalf("mark historical learner: %v", err)
	}

	partial := overviewCourse(t, database, "motivation-partial", learner.TenantID, 1, false, nil)
	partialLessons := []*domain.CourseLesson{
		overviewLesson(t, database, partial, "motivation-partial-a", 60),
		overviewLesson(t, database, partial, "motivation-partial-b", 90),
	}
	completed := overviewCourse(t, database, "motivation-completed", learner.TenantID, 1, false, nil)
	completedLesson := overviewLesson(t, database, completed, "motivation-completed", 45)
	draft := overviewCourse(t, database, "motivation-draft", learner.TenantID, 0, false, nil)
	draftLesson := overviewLesson(t, database, draft, "motivation-draft", 30)
	foreign := overviewCourse(t, database, "motivation-foreign", "motivation-foreign", 1, false, nil)
	foreignLesson := overviewLesson(t, database, foreign, "motivation-foreign", 30)
	for _, enrollment := range []*domain.CourseEnrollment{
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, CourseID: partial.ID, UserID: learner.ID, Status: 1, AssignmentType: domain.AssignmentRequired},
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, CourseID: completed.ID, UserID: learner.ID, Status: 1, AssignmentType: domain.AssignmentRequired},
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, CourseID: draft.ID, UserID: learner.ID, Status: 1, AssignmentType: domain.AssignmentRequired},
	} {
		if err := database.Create(enrollment).Error; err != nil {
			t.Fatalf("create enrollment: %v", err)
		}
	}

	yesterdayStart := time.Date(2026, 8, 16, 16, 0, 0, 0, time.UTC)
	todayStart := yesterdayStart.Add(24 * time.Hour)
	completedAt := yesterdayStart.Add(2 * time.Hour)
	for _, progress := range []*domain.LessonProgress{
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID, UpdatedAt: yesterdayStart.Add(time.Hour)}, UserID: learner.ID, LessonID: partialLessons[0].ID, ProgressPercent: 100, Status: 2, CompletedAt: &completedAt},
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID, UpdatedAt: yesterdayStart.Add(4 * time.Hour)}, UserID: learner.ID, LessonID: partialLessons[1].ID, ProgressPercent: 40, Status: 1, LastPositionSeconds: 33},
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID, UpdatedAt: yesterdayStart.Add(3 * time.Hour)}, UserID: learner.ID, LessonID: completedLesson.ID, ProgressPercent: 100, Status: 2, CompletedAt: &completedAt},
	} {
		if err := database.Create(progress).Error; err != nil {
			t.Fatalf("create progress: %v", err)
		}
	}
	for index, lessonID := range []string{partialLessons[0].ID, partialLessons[0].ID, partialLessons[1].ID, draftLesson.ID} {
		report := &domain.LearningTimeReport{
			BaseModel: domain.BaseModel{TenantID: learner.TenantID, CreatedAt: yesterdayStart.Add(time.Duration(index+1) * time.Minute)},
			UserID:    learner.ID, LessonID: lessonID, SessionID: "session", ReportID: fmt.Sprintf("report-%d", index), WatchedSecondsDelta: 30,
		}
		if err := database.Create(report).Error; err != nil {
			t.Fatalf("create report: %v", err)
		}
	}
	if err := database.Create(&domain.LearningTimeReport{
		BaseModel: domain.BaseModel{TenantID: "motivation-foreign", CreatedAt: yesterdayStart.Add(time.Hour)},
		UserID:    learner.ID, LessonID: foreignLesson.ID, SessionID: "foreign", ReportID: "foreign-report", WatchedSecondsDelta: 600,
	}).Error; err != nil {
		t.Fatalf("create foreign report: %v", err)
	}

	stats := []*domain.LearningDailyStat{
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, UserID: learner.ID, StudyDate: "2026-08-17", DurationSeconds: 1200},
		{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, UserID: learner.ID, StudyDate: "2026-08-16", DurationSeconds: 600},
	}
	peerDurations := []int64{100, 200, 300, 400, 500, 600, 700, 1200, 1300}
	for index, duration := range peerDurations {
		peer := &domain.User{
			BaseModel: domain.BaseModel{ID: fmt.Sprintf("motivation-peer-%d", index), TenantID: learner.TenantID},
			Email:     fmt.Sprintf("motivation-peer-%d@example.com", index), Password: "hash", Name: "Peer", Role: "learner", Status: 1,
		}
		if err := database.Create(peer).Error; err != nil {
			t.Fatalf("create peer: %v", err)
		}
		stats = append(stats, &domain.LearningDailyStat{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, UserID: peer.ID, StudyDate: "2026-08-17", DurationSeconds: duration})
	}
	disabled := &domain.User{BaseModel: domain.BaseModel{ID: "motivation-disabled", TenantID: learner.TenantID}, Email: "motivation-disabled@example.com", Password: "hash", Name: "Disabled", Role: "learner", Status: 0}
	if err := database.Create(disabled).Error; err != nil {
		t.Fatalf("create disabled learner: %v", err)
	}
	if err := database.Model(disabled).Update("status", 0).Error; err != nil {
		t.Fatalf("disable learner: %v", err)
	}
	stats = append(stats,
		&domain.LearningDailyStat{BaseModel: domain.BaseModel{TenantID: learner.TenantID}, UserID: disabled.ID, StudyDate: "2026-08-17", DurationSeconds: 50},
		&domain.LearningDailyStat{BaseModel: domain.BaseModel{TenantID: "motivation-foreign"}, UserID: "foreign-active", StudyDate: "2026-08-17", DurationSeconds: 5000},
	)
	for _, stat := range stats {
		if err := database.Create(stat).Error; err != nil {
			t.Fatalf("create daily stat: %v", err)
		}
	}

	repository := NewLearnerMotivationRepository(database)
	ctx := learnerRepositoryContext(learner.ID, learner.TenantID)
	got, err := repository.Load(ctx, "2026-08-18", "2026-08-17", "2026-08-16", yesterdayStart, todayStart)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.YesterdaySeconds != 1200 || got.DayBeforeSeconds != 600 {
		t.Fatalf("daily seconds = %d, %d", got.YesterdaySeconds, got.DayBeforeSeconds)
	}
	if got.YesterdayLessonCount != 2 || got.YesterdayCompletedLessonCount != 2 || got.YesterdayCompletedCourseCount != 1 {
		t.Fatalf("yesterday metrics = %#v", got)
	}
	if got.ActiveLearnerCount != 10 || got.ExceededPercent != 70 {
		t.Fatalf("tenant comparison = active %d, exceeded %d", got.ActiveLearnerCount, got.ExceededPercent)
	}
	if got.RequiredCompleted != 1 || got.RequiredTotal != 2 {
		t.Fatalf("required completion = %d/%d", got.RequiredCompleted, got.RequiredTotal)
	}
	if got.RecommendedCourse == nil || got.RecommendedCourse.ID != partial.ID ||
		got.RecommendedCourse.LessonID != partialLessons[1].ID || got.RecommendedCourse.LastPositionSeconds != 33 {
		t.Fatalf("recommended continuation = %#v", got.RecommendedCourse)
	}
}

func TestLearnerMotivationRepositoryIssuesAndAcknowledgesBoundPrompts(t *testing.T) {
	database := learningTimeDatabase(t)
	if err := database.Create(&domain.Tenant{ID: "prompt-tenant", Code: "prompt-tenant", Name: "Prompt", Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	learner := &domain.User{BaseModel: domain.BaseModel{ID: "prompt-learner", TenantID: "prompt-tenant"}, Email: "prompt@example.com", Password: "hash", Name: "Prompt", Role: "learner", Status: 1}
	if err := NewUserRepository(database).Create(context.Background(), learner); err != nil {
		t.Fatal(err)
	}
	loginAt := time.Date(2030, 1, 1, 1, 0, 0, 0, time.UTC)
	if err := database.Model(&domain.LearnerEngagementState{}).Where("user_id = ?", learner.ID).Update("first_login_at", loginAt).Error; err != nil {
		t.Fatal(err)
	}
	repository := NewLearnerMotivationRepository(database)
	ctx := learnerRepositoryContext(learner.ID, learner.TenantID)
	expiresAt := loginAt.Add(24 * time.Hour)
	key, err := repository.IssuePrompt(ctx, "welcome", "2030-01-01", expiresAt)
	if err != nil || key == "" {
		t.Fatalf("IssuePrompt() = %q, %v", key, err)
	}
	reused, err := repository.IssuePrompt(ctx, "welcome", "2030-01-01", expiresAt)
	if err != nil || reused != key {
		t.Fatalf("reused IssuePrompt() = %q, %v, want %q", reused, err, key)
	}
	if err := repository.AcknowledgePrompt(ctx, key, loginAt.Add(time.Hour)); err != nil {
		t.Fatalf("AcknowledgePrompt() error = %v", err)
	}
	if err := repository.AcknowledgePrompt(ctx, key, loginAt.Add(2*time.Hour)); err != nil {
		t.Fatalf("repeated AcknowledgePrompt() error = %v", err)
	}
	var state domain.LearnerEngagementState
	if err := database.Where("user_id = ?", learner.ID).First(&state).Error; err != nil {
		t.Fatal(err)
	}
	if state.WelcomeSeenAt == nil || state.LastDailyPromptDate != "2030-01-01" || state.PendingPromptKey != key {
		t.Fatalf("acknowledged welcome state = %#v", state)
	}
	if err := repository.AcknowledgePrompt(learnerRepositoryContext("other", learner.TenantID), key, loginAt.Add(time.Hour)); err == nil {
		t.Fatal("cross-user acknowledgement succeeded")
	}
	dailyKey, err := repository.IssuePrompt(ctx, "daily_summary", "2030-01-02", expiresAt.Add(24*time.Hour))
	if err != nil || dailyKey == key {
		t.Fatalf("daily IssuePrompt() = %q, %v", dailyKey, err)
	}
	if err := repository.AcknowledgePrompt(ctx, dailyKey, loginAt.Add(25*time.Hour)); err != nil {
		t.Fatalf("acknowledge daily prompt: %v", err)
	}
	if err := database.Where("user_id = ?", learner.ID).First(&state).Error; err != nil {
		t.Fatal(err)
	}
	if state.LastDailyPromptDate != "2030-01-02" {
		t.Fatalf("last daily prompt date = %q", state.LastDailyPromptDate)
	}
}

func TestLearnerMotivationRecommendationUsesStableProductPriority(t *testing.T) {
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	lesson := func(id string) domain.CourseLesson {
		return domain.CourseLesson{BaseModel: domain.BaseModel{ID: id}, Title: id}
	}
	course := func(id string, official bool, assignment string, enrollment *domain.CourseEnrollment) motivationVisibleCourse {
		lessons := []domain.CourseLesson{lesson(id + "-done"), lesson(id + "-next")}
		return motivationVisibleCourse{
			data: LearnerOverviewCourse{
				Course:         domain.Course{BaseModel: domain.BaseModel{ID: id}, Title: id, IsOfficial: official},
				AssignmentType: assignment, LessonCount: 2, CompletedLessonCount: 1, ProgressPercent: 60,
			},
			lessons: lessons,
			progress: map[string]domain.LessonProgress{
				lessons[0].ID: {LessonID: lessons[0].ID, Status: 2, ProgressPercent: 100},
				lessons[1].ID: {LessonID: lessons[1].ID, Status: 1, ProgressPercent: 20, LastPositionSeconds: 44},
			},
			enrollment: enrollment,
		}
	}
	requiredEnrollment := &domain.CourseEnrollment{BaseModel: domain.BaseModel{CreatedAt: base.Add(time.Hour)}, AssignmentType: domain.AssignmentRequired}
	olderRequiredEnrollment := &domain.CourseEnrollment{BaseModel: domain.BaseModel{CreatedAt: base}, AssignmentType: domain.AssignmentRequired}
	optionalEnrollment := &domain.CourseEnrollment{BaseModel: domain.BaseModel{CreatedAt: base.Add(-time.Hour)}, AssignmentType: domain.AssignmentOptional}
	assignedRequired := course("assigned-required", false, domain.AssignmentRequired, requiredEnrollment)
	olderAssignedRequired := course("older-assigned-required", false, domain.AssignmentRequired, olderRequiredEnrollment)
	officialRequired := course("official-required", true, domain.AssignmentRequired, nil)
	assignedOptional := course("assigned-optional", false, domain.AssignmentOptional, optionalEnrollment)
	accessible := course("accessible", true, domain.AssignmentOptional, nil)

	for _, test := range []struct {
		name    string
		courses []motivationVisibleCourse
		want    string
	}{
		{"oldest assigned required first", []motivationVisibleCourse{assignedRequired, olderAssignedRequired, officialRequired, assignedOptional, accessible}, olderAssignedRequired.data.Course.ID},
		{"official required before other assignment", []motivationVisibleCourse{officialRequired, assignedOptional, accessible}, officialRequired.data.Course.ID},
		{"other assignment before accessible fallback", []motivationVisibleCourse{assignedOptional, accessible}, assignedOptional.data.Course.ID},
		{"accessible fallback", []motivationVisibleCourse{accessible}, accessible.data.Course.ID},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := selectMotivationCourse(test.courses)
			if got == nil || got.ID != test.want || got.LessonID != test.want+"-next" || got.LastPositionSeconds != 44 {
				t.Fatalf("selectMotivationCourse() = %#v, want %s next lesson", got, test.want)
			}
		})
	}
}
