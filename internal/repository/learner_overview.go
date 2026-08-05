package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
)

type LearnerOverviewCourse struct {
	Course               domain.Course
	AssignmentType       string
	Category             *domain.CourseCategory
	LessonCount          int
	CompletedLessonCount int
	ProgressPercent      int
	LastLearnedAt        *time.Time
	RecentLesson         *domain.CourseLesson
	LastPositionSeconds  int
}

type LearnerOverviewData struct {
	TodayLearningSeconds int64
	TotalLearningSeconds int64
	Courses              []LearnerOverviewCourse
}

type LearnerOverviewRepository interface {
	Get(ctx context.Context, studyDate string) (LearnerOverviewData, error)
}
