package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalidLearningTimeReport = errors.New("invalid learning time report")

type learningTimeGORMRepository struct {
	database *gorm.DB
	now      func() time.Time
}

func NewLearningTimeRepository(database *gorm.DB) *learningTimeGORMRepository {
	return &learningTimeGORMRepository{database: database, now: time.Now}
}

func (repo *learningTimeGORMRepository) Record(
	ctx context.Context,
	report *domain.LearningTimeReport,
	studyDate string,
) (bool, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || userID == "" || tenantID == "" ||
		report == nil || strings.TrimSpace(report.LessonID) == "" ||
		strings.TrimSpace(report.SessionID) == "" ||
		strings.TrimSpace(report.ReportID) == "" ||
		report.WatchedSecondsDelta < 1 || report.WatchedSecondsDelta > 60 ||
		strings.TrimSpace(studyDate) == "" {
		return false, ErrInvalidLearningTimeReport
	}

	report.TenantID = tenantID
	report.UserID = userID
	report.SessionID = strings.TrimSpace(report.SessionID)
	report.ReportID = strings.TrimSpace(report.ReportID)
	reportedAt := repo.now().UTC()
	report.CreatedAt = reportedAt
	report.UpdatedAt = reportedAt
	recorded := false
	err := repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var duplicate int64
		if err := tx.Model(&domain.LearningTimeReport{}).Where(
			"tenant_id = ? AND user_id = ? AND report_id = ?",
			tenantID, userID, report.ReportID,
		).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return nil
		}
		if err := validateLearningSessionTime(tx, report, reportedAt); err != nil {
			return err
		}
		result := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "tenant_id"}, {Name: "user_id"}, {Name: "report_id"},
			},
			DoNothing: true,
		}).Create(report)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		stat := &domain.LearningDailyStat{
			BaseModel:       domain.BaseModel{TenantID: tenantID},
			UserID:          userID,
			StudyDate:       studyDate,
			DurationSeconds: int64(report.WatchedSecondsDelta),
		}
		if err := tx.Clauses(
			learningDailyUpsertClause(report.WatchedSecondsDelta),
		).Create(stat).Error; err != nil {
			return err
		}
		recorded = true
		return nil
	})
	return recorded, err
}

func validateLearningSessionTime(
	database *gorm.DB, report *domain.LearningTimeReport, reportedAt time.Time,
) error {
	var reports []domain.LearningTimeReport
	if err := database.Where(
		"tenant_id = ? AND user_id = ? AND session_id = ?",
		report.TenantID, report.UserID, report.SessionID,
	).Order("created_at ASC").Find(&reports).Error; err != nil {
		return err
	}
	if len(reports) == 0 {
		if report.WatchedSecondsDelta > 20 {
			return ErrInvalidLearningTimeReport
		}
		return nil
	}
	accumulated := 0
	for _, existing := range reports {
		accumulated += existing.WatchedSecondsDelta
	}
	elapsed := int(reportedAt.Sub(reports[0].CreatedAt).Seconds())
	if elapsed < 0 {
		elapsed = 0
	}
	maximum := reports[0].WatchedSecondsDelta + elapsed + 5
	if accumulated+report.WatchedSecondsDelta > maximum {
		return ErrInvalidLearningTimeReport
	}
	return nil
}

func learningDailyUpsertClause(delta int) clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"}, {Name: "user_id"}, {Name: "study_date"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"duration_seconds": gorm.Expr(
				"? + ?",
				clause.Column{Table: "learning_daily_stats", Name: "duration_seconds"},
				delta,
			),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}
}
