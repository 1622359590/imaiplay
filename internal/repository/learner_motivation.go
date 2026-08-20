package repository

import (
	"context"
	"time"
)

type LearnerMotivationRepository interface {
	MarkFirstLogin(ctx context.Context, userID string, at time.Time) (bool, error)
}
