package repository

import (
	"context"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type learnerMotivationGORMRepository struct {
	database *gorm.DB
}

func NewLearnerMotivationRepository(database *gorm.DB) LearnerMotivationRepository {
	return &learnerMotivationGORMRepository{database: database}
}

func (repository *learnerMotivationGORMRepository) MarkFirstLogin(
	ctx context.Context,
	userID string,
	at time.Time,
) (bool, error) {
	contextUserID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || tenantID == "" || userID == "" || contextUserID != userID {
		return false, gorm.ErrRecordNotFound
	}
	result := repository.database.WithContext(ctx).
		Model(&domain.LearnerEngagementState{}).
		Where("tenant_id = ? AND user_id = ? AND first_login_at IS NULL", tenantID, userID).
		Update("first_login_at", at.UTC())
	return result.RowsAffected == 1, result.Error
}
