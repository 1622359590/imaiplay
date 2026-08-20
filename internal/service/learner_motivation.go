package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

const (
	LearnerMotivationNone         = "none"
	LearnerMotivationWelcome      = "welcome"
	LearnerMotivationDailySummary = "daily_summary"
	LearnerMotivationReengagement = "reengagement"
)

type LearnerMotivation struct {
	Kind       string                       `json:"kind"`
	PromptKey  string                       `json:"prompt_key,omitempty"`
	StudyDate  string                       `json:"study_date,omitempty"`
	Title      string                       `json:"title,omitempty"`
	Message    string                       `json:"message,omitempty"`
	Metrics    *LearnerMotivationMetrics    `json:"metrics,omitempty"`
	Comparison *LearnerMotivationComparison `json:"comparison,omitempty"`
	Course     *LearnerMotivationCourse     `json:"course,omitempty"`
}

type LearnerMotivationMetrics struct {
	YesterdaySeconds     int64 `json:"yesterday_seconds"`
	LessonCount          int   `json:"lesson_count"`
	CompletedLessonCount int   `json:"completed_lesson_count"`
	CompletedCourseCount int   `json:"completed_course_count"`
	RequiredCompleted    int   `json:"required_completed"`
	RequiredTotal        int   `json:"required_total"`
}

type LearnerMotivationComparison struct {
	DurationChangeSeconds *int64 `json:"duration_change_seconds,omitempty"`
	ExceededPercent       *int   `json:"exceeded_percent,omitempty"`
	ActiveLearnerCount    *int   `json:"active_learner_count,omitempty"`
}

type LearnerMotivationCourse struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	AssignmentType      string `json:"assignment_type"`
	LessonCount         int    `json:"lesson_count"`
	ProgressPercent     int    `json:"progress_percent"`
	LessonID            string `json:"lesson_id"`
	LessonTitle         string `json:"lesson_title"`
	LastPositionSeconds int    `json:"last_position_seconds"`
}

type LearnerMotivationService struct {
	repository repository.LearnerMotivationRepository
	now        func() time.Time
}

func NewLearnerMotivationService(
	repository repository.LearnerMotivationRepository,
) *LearnerMotivationService {
	return &LearnerMotivationService{repository: repository, now: time.Now}
}

func (service *LearnerMotivationService) Get(ctx context.Context) (LearnerMotivation, error) {
	if _, _, err := learnerIdentity(ctx); err != nil {
		return LearnerMotivation{}, err
	}
	now := service.now()
	location := shanghaiLocation()
	localNow := now.In(location)
	todayStartLocal := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	yesterdayStartLocal := todayStartLocal.AddDate(0, 0, -1)
	dayBeforeStartLocal := yesterdayStartLocal.AddDate(0, 0, -1)
	today := todayStartLocal.Format("2006-01-02")
	yesterday := yesterdayStartLocal.Format("2006-01-02")
	dayBefore := dayBeforeStartLocal.Format("2006-01-02")

	snapshot, err := service.repository.Load(
		ctx,
		today,
		yesterday,
		dayBefore,
		yesterdayStartLocal.UTC(),
		todayStartLocal.UTC(),
	)
	if err != nil {
		return LearnerMotivation{}, errorsx.Internal("load learner motivation failed")
	}
	if snapshot.State.FirstLoginAt == nil {
		return LearnerMotivation{Kind: LearnerMotivationNone}, nil
	}
	if snapshot.State.WelcomeSeenAt == nil {
		course := presentMotivationCourse(snapshot.RecommendedCourse)
		message := "从第一门课程开始，把每一次学习积累成成长。"
		if course == nil {
			message = "当前暂无学习任务，管理员发布课程后即可开始学习。"
		}
		return service.issue(ctx, LearnerMotivation{
			Kind:    LearnerMotivationWelcome,
			Title:   "欢迎开启你的学习旅程",
			Message: message,
			Course:  course,
		}, today, now)
	}
	if snapshot.State.LastDailyPromptDate == today {
		return LearnerMotivation{Kind: LearnerMotivationNone}, nil
	}
	if snapshot.RecommendedCourse == nil {
		return LearnerMotivation{Kind: LearnerMotivationNone}, nil
	}
	if snapshot.YesterdaySeconds <= 0 {
		return service.issue(ctx, LearnerMotivation{
			Kind:    LearnerMotivationReengagement,
			Title:   "欢迎回来，继续你的学习节奏",
			Message: "从上次学习的位置继续，今天也向前一点。",
			Course:  presentMotivationCourse(snapshot.RecommendedCourse),
		}, today, now)
	}

	motivation := LearnerMotivation{
		Kind:      LearnerMotivationDailySummary,
		StudyDate: yesterday,
		Title:     "昨天的学习有了新成果",
		Metrics: &LearnerMotivationMetrics{
			YesterdaySeconds:     snapshot.YesterdaySeconds,
			LessonCount:          snapshot.YesterdayLessonCount,
			CompletedLessonCount: snapshot.YesterdayCompletedLessonCount,
			CompletedCourseCount: snapshot.YesterdayCompletedCourseCount,
			RequiredCompleted:    snapshot.RequiredCompleted,
			RequiredTotal:        snapshot.RequiredTotal,
		},
		Course: presentMotivationCourse(snapshot.RecommendedCourse),
	}
	if snapshot.YesterdayCompletedCourseCount > 0 {
		motivation.Message = fmt.Sprintf("昨天完成了 %d 门课程，做得很好。", snapshot.YesterdayCompletedCourseCount)
	} else {
		motivation.Message = fmt.Sprintf("昨天投入了%s，学习节奏正在形成。", motivationDurationText(snapshot.YesterdaySeconds))
	}
	comparison := &LearnerMotivationComparison{}
	if snapshot.DayBeforeSeconds > 0 {
		change := snapshot.YesterdaySeconds - snapshot.DayBeforeSeconds
		comparison.DurationChangeSeconds = &change
		if change > 0 {
			motivation.Message = strings.TrimSuffix(motivation.Message, "。") +
				fmt.Sprintf("，比前一天多学习%s，继续保持。", motivationDurationText(change))
		}
	}
	if snapshot.ActiveLearnerCount >= 10 {
		exceeded := snapshot.ExceededPercent
		active := snapshot.ActiveLearnerCount
		comparison.ExceededPercent = &exceeded
		comparison.ActiveLearnerCount = &active
	}
	if comparison.DurationChangeSeconds != nil || comparison.ExceededPercent != nil {
		motivation.Comparison = comparison
	}
	return service.issue(ctx, motivation, today, now)
}

func (service *LearnerMotivationService) Acknowledge(ctx context.Context, key string) error {
	if _, _, err := learnerIdentity(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return errorsx.BadRequest("prompt key is required")
	}
	if err := service.repository.AcknowledgePrompt(ctx, strings.TrimSpace(key), service.now().UTC()); err != nil {
		return errorsx.BadRequest("invalid or expired prompt key")
	}
	return nil
}

func (service *LearnerMotivationService) issue(
	ctx context.Context,
	motivation LearnerMotivation,
	promptDate string,
	now time.Time,
) (LearnerMotivation, error) {
	key, err := service.repository.IssuePrompt(ctx, motivation.Kind, promptDate, now.UTC().Add(24*time.Hour))
	if err != nil {
		return LearnerMotivation{}, errorsx.Internal("issue learner motivation failed")
	}
	motivation.PromptKey = key
	return motivation, nil
}

func presentMotivationCourse(course *repository.RecommendedCourse) *LearnerMotivationCourse {
	if course == nil || course.ID == "" || course.LessonID == "" {
		return nil
	}
	return &LearnerMotivationCourse{
		ID: course.ID, Title: course.Title, AssignmentType: course.AssignmentType,
		LessonCount: course.LessonCount, ProgressPercent: course.ProgressPercent,
		LessonID: course.LessonID, LessonTitle: course.LessonTitle,
		LastPositionSeconds: course.LastPositionSeconds,
	}
}

func motivationDurationText(seconds int64) string {
	if seconds < 60 {
		return "不足 1 分钟"
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%d 分钟", minutes)
	}
	hours := minutes / 60
	remaining := minutes % 60
	if remaining == 0 {
		return fmt.Sprintf("%d 小时", hours)
	}
	return fmt.Sprintf("%d 小时 %d 分钟", hours, remaining)
}

func shanghaiLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return location
}
