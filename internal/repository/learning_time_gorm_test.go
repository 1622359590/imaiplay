package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLearningDailyUpsertQualifiesPostgresConflictColumn(t *testing.T) {
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=test password=test dbname=test sslmode=disable",
	}), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open dry-run postgres: %v", err)
	}
	stat := &domain.LearningDailyStat{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		UserID:    "learner-1", StudyDate: "2026-08-05", DurationSeconds: 15,
	}
	statement := database.Clauses(learningDailyUpsertClause(15)).Create(stat).Statement.SQL.String()
	if !strings.Contains(statement, `"learning_daily_stats"."duration_seconds" +`) {
		t.Fatalf("postgres upsert leaves duration_seconds ambiguous: %s", statement)
	}
}

func TestLearningTimeRepositoryRecordsAcceptedDeltasAndBindsLearnerIdentity(t *testing.T) {
	database := learningTimeDatabase(t)
	repo := NewLearningTimeRepository(database)
	ctx := learnerRepositoryContext("learner-1", "tenant-1")

	for index, delta := range []int{1, 15} {
		recorded, err := repo.Record(ctx, &domain.LearningTimeReport{
			BaseModel:           domain.BaseModel{TenantID: "attacker-tenant"},
			UserID:              "attacker-user",
			LessonID:            "lesson-1",
			SessionID:           "session-" + string(rune('1'+index)),
			ReportID:            "report-" + string(rune('1'+index)),
			WatchedSecondsDelta: delta,
		}, "2026-08-05")
		if err != nil || !recorded {
			t.Fatalf("Record(delta=%d) = %v, %v", delta, recorded, err)
		}
	}

	var stat domain.LearningDailyStat
	if err := database.Where(
		"tenant_id = ? AND user_id = ? AND study_date = ?",
		"tenant-1", "learner-1", "2026-08-05",
	).First(&stat).Error; err != nil || stat.DurationSeconds != 16 {
		t.Fatalf("daily stat = %#v, %v", stat, err)
	}
	var reports []domain.LearningTimeReport
	if err := database.Order("report_id ASC").Find(&reports).Error; err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("reports = %#v", reports)
	}
	for _, report := range reports {
		if report.TenantID != "tenant-1" || report.UserID != "learner-1" {
			t.Fatalf("report identity was not bound to context: %#v", report)
		}
	}
}

func TestLearningTimeRepositoryRejectsInvalidReports(t *testing.T) {
	database := learningTimeDatabase(t)
	repo := NewLearningTimeRepository(database)
	learner := learnerRepositoryContext("learner-1", "tenant-1")
	admin := usercontext.WithUser(context.Background(), "admin-1", "tenant-1", "", "tenant_admin")

	tests := []struct {
		name      string
		ctx       context.Context
		delta     int
		reportID  string
		sessionID string
	}{
		{name: "zero", ctx: learner, delta: 0, reportID: "zero", sessionID: "session-1"},
		{name: "negative", ctx: learner, delta: -1, reportID: "negative", sessionID: "session-1"},
		{name: "over maximum", ctx: learner, delta: 61, reportID: "large", sessionID: "session-1"},
		{name: "missing report id", ctx: learner, delta: 1, sessionID: "session-1"},
		{name: "missing session id", ctx: learner, delta: 1, reportID: "missing-session"},
		{name: "non learner context", ctx: admin, delta: 1, reportID: "admin", sessionID: "session-1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorded, err := repo.Record(test.ctx, &domain.LearningTimeReport{
				LessonID: "lesson-1", SessionID: test.sessionID, ReportID: test.reportID,
				WatchedSecondsDelta: test.delta,
			}, "2026-08-05")
			if err == nil || recorded {
				t.Fatalf("Record() = %v, %v; want rejected", recorded, err)
			}
		})
	}
	var count int64
	if err := database.Model(&domain.LearningTimeReport{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("report count = %d, %v", count, err)
	}
}

func TestLearningTimeRepositoryDuplicateReportDoesNotAccumulate(t *testing.T) {
	database := learningTimeDatabase(t)
	repo := NewLearningTimeRepository(database)
	ctx := learnerRepositoryContext("learner-1", "tenant-1")
	newReport := func(delta int) *domain.LearningTimeReport {
		return &domain.LearningTimeReport{
			LessonID: "lesson-1", SessionID: "session-1", ReportID: "report-1", WatchedSecondsDelta: delta,
		}
	}

	first, err := repo.Record(ctx, newReport(15), "2026-08-05")
	second, err2 := repo.Record(ctx, newReport(60), "2026-08-05")
	if err != nil || err2 != nil || !first || second {
		t.Fatalf("Record duplicate = %v/%v, errors=%v/%v", first, second, err, err2)
	}
	var stat domain.LearningDailyStat
	if err := database.Where("tenant_id = ? AND user_id = ?", "tenant-1", "learner-1").First(&stat).Error; err != nil || stat.DurationSeconds != 15 {
		t.Fatalf("daily stat = %#v, %v", stat, err)
	}
}

func TestLearningTimeRepositoryRejectsForgedSessionTime(t *testing.T) {
	database := learningTimeDatabase(t)
	repo := NewLearningTimeRepository(database)
	ctx := learnerRepositoryContext("learner-1", "tenant-1")
	now := time.Date(2026, 8, 11, 8, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return now }

	for _, report := range []*domain.LearningTimeReport{
		{LessonID: "lesson-1", SessionID: "forged", ReportID: "too-fast", WatchedSecondsDelta: 60},
		{LessonID: "lesson-1", SessionID: "valid", ReportID: "first", WatchedSecondsDelta: 15},
	} {
		recorded, err := repo.Record(ctx, report, "2026-08-11")
		if report.ReportID == "too-fast" {
			if !errors.Is(err, ErrInvalidLearningTimeReport) || recorded {
				t.Fatalf("forged first heartbeat = %v, %v", recorded, err)
			}
			continue
		}
		if err != nil || !recorded {
			t.Fatalf("valid first heartbeat = %v, %v", recorded, err)
		}
	}

	if recorded, err := repo.Record(ctx, &domain.LearningTimeReport{
		LessonID: "lesson-1", SessionID: "valid", ReportID: "instant-second", WatchedSecondsDelta: 15,
	}, "2026-08-11"); !errors.Is(err, ErrInvalidLearningTimeReport) || recorded {
		t.Fatalf("instant second heartbeat = %v, %v", recorded, err)
	}
	now = now.Add(15 * time.Second)
	if recorded, err := repo.Record(ctx, &domain.LearningTimeReport{
		LessonID: "lesson-1", SessionID: "valid", ReportID: "elapsed-second", WatchedSecondsDelta: 15,
	}, "2026-08-11"); err != nil || !recorded {
		t.Fatalf("elapsed second heartbeat = %v, %v", recorded, err)
	}
}

func TestLearningTimeRepositoryRollsBackReportWhenDailyUpdateFails(t *testing.T) {
	database := learningTimeDatabase(t)
	if err := database.Exec(`
		CREATE TRIGGER reject_learning_daily
		BEFORE INSERT ON learning_daily_stats
		BEGIN
			SELECT RAISE(ABORT, 'reject daily stat');
		END
	`).Error; err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	repo := NewLearningTimeRepository(database)
	ctx := learnerRepositoryContext("learner-1", "tenant-1")

	recorded, err := repo.Record(ctx, &domain.LearningTimeReport{
		LessonID: "lesson-1", SessionID: "session-1", ReportID: "rollback-report", WatchedSecondsDelta: 15,
	}, "2026-08-05")
	if err == nil || recorded {
		t.Fatalf("Record() = %v, %v; want transaction failure", recorded, err)
	}
	var count int64
	if err := database.Model(&domain.LearningTimeReport{}).
		Where("report_id = ?", "rollback-report").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("rolled back report count = %d, %v", count, err)
	}
}

func learningTimeDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return database
}

func learnerRepositoryContext(userID, tenantID string) context.Context {
	return usercontext.WithUser(context.Background(), userID, tenantID, "", "learner")
}
