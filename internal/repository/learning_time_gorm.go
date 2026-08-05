package repository

import (
	"context"
	"errors"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalidLearningTimeReport = errors.New("invalid learning time report")

type learningTimeGORMRepository struct {
	database *gorm.DB
}

func NewLearningTimeRepository(database *gorm.DB) LearningTimeRepository {
	return &learningTimeGORMRepository{database: database}
}

func (repo *learningTimeGORMRepository) Record(
	ctx context.Context,
	report *domain.LearningTimeReport,
	studyDate string,
) (bool, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || userID == "" || tenantID == "" ||
		report == nil || strings.TrimSpace(report.LessonID) == "" ||
		strings.TrimSpace(report.ReportID) == "" ||
		report.WatchedSecondsDelta < 1 || report.WatchedSecondsDelta > 60 ||
		strings.TrimSpace(studyDate) == "" {
		return false, ErrInvalidLearningTimeReport
	}

	report.TenantID = tenantID
	report.UserID = userID
	report.ReportID = strings.TrimSpace(report.ReportID)
	recorded := false
	err := repo.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
