package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
)

type LearnerMotivationRepository interface {
	LearnerFirstLoginRepository
	Load(
		ctx context.Context,
		today, yesterday, dayBefore string,
		yesterdayStart, todayStart time.Time,
	) (LearnerMotivationSnapshot, error)
	IssuePrompt(ctx context.Context, kind, promptDate string, expiresAt time.Time) (string, error)
	AcknowledgePrompt(ctx context.Context, key string, at time.Time) error
}

type LearnerFirstLoginRepository interface {
	MarkFirstLogin(ctx context.Context, userID string, at time.Time) (bool, error)
}

type RecommendedCourse struct {
	ID                  string
	Title               string
	AssignmentType      string
	LessonCount         int
	ProgressPercent     int
	LessonID            string
	LessonTitle         string
	LastPositionSeconds int
}

type LearnerMotivationSnapshot struct {
	State                         domain.LearnerEngagementState
	YesterdaySeconds              int64
	DayBeforeSeconds              int64
	YesterdayLessonCount          int
	YesterdayCompletedLessonCount int
	YesterdayCompletedCourseCount int
	ActiveLearnerCount            int
	ExceededPercent               int
	RequiredCompleted             int
	RequiredTotal                 int
	RecommendedCourse             *RecommendedCourse
}
