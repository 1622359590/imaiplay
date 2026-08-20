package service

import (
	"context"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

type learnerMotivationStub struct {
	snapshot        repository.LearnerMotivationSnapshot
	today           string
	yesterday       string
	dayBefore       string
	yesterdayStart  time.Time
	todayStart      time.Time
	issuedKind      string
	issuedDate      string
	acknowledgedKey string
}

func (stub *learnerMotivationStub) MarkFirstLogin(context.Context, string, time.Time) (bool, error) {
	return false, nil
}

func (stub *learnerMotivationStub) Load(
	_ context.Context,
	today, yesterday, dayBefore string,
	yesterdayStart, todayStart time.Time,
) (repository.LearnerMotivationSnapshot, error) {
	stub.today, stub.yesterday, stub.dayBefore = today, yesterday, dayBefore
	stub.yesterdayStart, stub.todayStart = yesterdayStart, todayStart
	return stub.snapshot, nil
}

func (stub *learnerMotivationStub) IssuePrompt(_ context.Context, kind, promptDate string, _ time.Time) (string, error) {
	stub.issuedKind, stub.issuedDate = kind, promptDate
	return "prompt-key", nil
}

func (stub *learnerMotivationStub) AcknowledgePrompt(_ context.Context, key string, _ time.Time) error {
	stub.acknowledgedKey = key
	return nil
}

func TestLearnerMotivationServiceSelectsPromptStatesAndShanghaiDates(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, location)
	firstLogin := now.Add(-time.Hour)
	course := &repository.RecommendedCourse{
		ID: "course-1", Title: "新员工入职培训", AssignmentType: domain.AssignmentRequired,
		LessonCount: 4, ProgressPercent: 25, LessonID: "lesson-1", LessonTitle: "欢迎视频", LastPositionSeconds: 12,
	}
	ctx := usercontext.WithUser(context.Background(), "learner-1", "tenant-1", "", "learner")

	t.Run("welcome", func(t *testing.T) {
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State: domain.LearnerEngagementState{FirstLoginAt: &firstLogin}, RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Kind != "welcome" || got.PromptKey == "" || got.Course == nil {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
		if stub.today != "2026-08-18" || stub.yesterday != "2026-08-17" || stub.dayBefore != "2026-08-16" {
			t.Fatalf("queried dates = %q, %q, %q", stub.today, stub.yesterday, stub.dayBefore)
		}
		if !stub.yesterdayStart.Equal(time.Date(2026, 8, 16, 16, 0, 0, 0, time.UTC)) ||
			!stub.todayStart.Equal(time.Date(2026, 8, 17, 16, 0, 0, 0, time.UTC)) {
			t.Fatalf("UTC day bounds = %v, %v", stub.yesterdayStart, stub.todayStart)
		}
	})

	t.Run("daily summary without small cohort ranking", func(t *testing.T) {
		welcomeSeen := now.Add(-24 * time.Hour)
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State:            domain.LearnerEngagementState{FirstLoginAt: &firstLogin, WelcomeSeenAt: &welcomeSeen},
			YesterdaySeconds: 1680, DayBeforeSeconds: 960, YesterdayLessonCount: 2,
			YesterdayCompletedLessonCount: 1, YesterdayCompletedCourseCount: 1,
			ActiveLearnerCount: 9, ExceededPercent: 77, RequiredCompleted: 1, RequiredTotal: 2,
			RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Kind != "daily_summary" || got.Metrics == nil {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
		if got.Comparison == nil || got.Comparison.ExceededPercent != nil {
			t.Fatalf("small-cohort comparison = %#v", got.Comparison)
		}
		if stub.issuedKind != "daily_summary" || stub.issuedDate != "2026-08-18" {
			t.Fatalf("issued prompt = %q/%q", stub.issuedKind, stub.issuedDate)
		}
	})

	t.Run("daily summary with tenant ranking", func(t *testing.T) {
		welcomeSeen := now.Add(-24 * time.Hour)
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State:            domain.LearnerEngagementState{FirstLoginAt: &firstLogin, WelcomeSeenAt: &welcomeSeen},
			YesterdaySeconds: 600, ActiveLearnerCount: 10, ExceededPercent: 70, RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Comparison == nil || got.Comparison.ExceededPercent == nil || *got.Comparison.ExceededPercent != 70 {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
	})

	t.Run("reengagement stays positive", func(t *testing.T) {
		welcomeSeen := now.Add(-24 * time.Hour)
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State: domain.LearnerEngagementState{FirstLoginAt: &firstLogin, WelcomeSeenAt: &welcomeSeen}, RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Kind != "reengagement" || got.Metrics != nil || got.Course == nil {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
		if got.Title != "欢迎回来，继续你的学习节奏" || got.Message == "" {
			t.Fatalf("reengagement copy = %q / %q", got.Title, got.Message)
		}
	})

	t.Run("already shown today", func(t *testing.T) {
		welcomeSeen := now.Add(-24 * time.Hour)
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State:            domain.LearnerEngagementState{FirstLoginAt: &firstLogin, WelcomeSeenAt: &welcomeSeen, LastDailyPromptDate: "2026-08-18"},
			YesterdaySeconds: 600, RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Kind != "none" || stub.issuedKind != "" {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
	})

	t.Run("welcome acknowledgement suppresses same-day daily prompt", func(t *testing.T) {
		welcomeSeen := now.Add(-time.Hour)
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State:            domain.LearnerEngagementState{FirstLoginAt: &firstLogin, WelcomeSeenAt: &welcomeSeen, LastDailyPromptDate: "2026-08-18"},
			YesterdaySeconds: 600, RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Kind != "none" {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
	})

	t.Run("historical backfill enters the returning learner flow immediately", func(t *testing.T) {
		backfilled := now.Add(-time.Hour)
		stub := &learnerMotivationStub{snapshot: repository.LearnerMotivationSnapshot{
			State:            domain.LearnerEngagementState{FirstLoginAt: &backfilled, WelcomeSeenAt: &backfilled},
			YesterdaySeconds: 600, RecommendedCourse: course,
		}}
		service := NewLearnerMotivationService(stub)
		service.now = func() time.Time { return now }
		got, err := service.Get(ctx)
		if err != nil || got.Kind != "daily_summary" {
			t.Fatalf("Get() = %#v, %v", got, err)
		}
	})
}

func TestLearnerMotivationServiceAcknowledgesOpaquePrompt(t *testing.T) {
	stub := &learnerMotivationStub{}
	service := NewLearnerMotivationService(stub)
	service.now = func() time.Time { return time.Date(2026, 8, 18, 1, 0, 0, 0, time.UTC) }
	ctx := usercontext.WithUser(context.Background(), "learner-1", "tenant-1", "", "learner")
	if err := service.Acknowledge(ctx, "prompt-key"); err != nil {
		t.Fatalf("Acknowledge() error = %v", err)
	}
	if stub.acknowledgedKey != "prompt-key" {
		t.Fatalf("acknowledged key = %q", stub.acknowledgedKey)
	}
}
