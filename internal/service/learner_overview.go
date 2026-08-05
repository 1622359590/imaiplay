package service

import (
	"context"
	"sort"
	"time"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type CourseCategorySummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CourseSummary struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CoverImage  string                 `json:"cover_image"`
	Category    *CourseCategorySummary `json:"category,omitempty"`
}

type LessonSummary struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	DurationSeconds     int    `json:"duration_seconds"`
	LastPositionSeconds int    `json:"last_position_seconds"`
}

type LearnerCourseSummary struct {
	Course               CourseSummary  `json:"course"`
	AssignmentType       string         `json:"assignment_type"`
	LessonCount          int            `json:"lesson_count"`
	CompletedLessonCount int            `json:"completed_lesson_count"`
	ProgressPercent      int            `json:"progress_percent"`
	LastLearnedAt        *time.Time     `json:"last_learned_at,omitempty"`
	RecentLesson         *LessonSummary `json:"recent_lesson,omitempty"`
}

type LearnerOverview struct {
	RequiredCompleted    int                     `json:"required_completed"`
	RequiredTotal        int                     `json:"required_total"`
	TodayLearningSeconds int64                   `json:"today_learning_seconds"`
	TotalLearningSeconds int64                   `json:"total_learning_seconds"`
	Categories           []CourseCategorySummary `json:"categories"`
	Courses              []LearnerCourseSummary  `json:"courses"`
}

type RecentLearningItem struct {
	Course              CourseSummary `json:"course"`
	RecentLesson        LessonSummary `json:"recent_lesson"`
	ProgressPercent     int           `json:"progress_percent"`
	LastPositionSeconds int           `json:"last_position_seconds"`
	LastLearnedAt       time.Time     `json:"last_learned_at"`
}

type LearnerOverviewService struct {
	overview repository.LearnerOverviewRepository
	now      func() time.Time
}

func NewLearnerOverviewService(
	overview repository.LearnerOverviewRepository,
) *LearnerOverviewService {
	return &LearnerOverviewService{overview: overview, now: time.Now}
}

func (service *LearnerOverviewService) Get(
	ctx context.Context,
) (LearnerOverview, error) {
	if _, _, err := learnerIdentity(ctx); err != nil {
		return LearnerOverview{}, err
	}
	data, err := service.overview.Get(ctx, shanghaiStudyDate(service.now()))
	if err != nil {
		return LearnerOverview{}, errorsx.Internal("load learner overview failed")
	}
	result := LearnerOverview{
		TodayLearningSeconds: data.TodayLearningSeconds,
		TotalLearningSeconds: data.TotalLearningSeconds,
		Categories:           []CourseCategorySummary{},
		Courses:              []LearnerCourseSummary{},
	}
	categories := make(map[string]CourseCategorySummary)
	for _, item := range data.Courses {
		course := presentLearnerCourse(item)
		result.Courses = append(result.Courses, course)
		if course.Course.Category != nil {
			categories[course.Course.Category.ID] = *course.Course.Category
		}
		if course.AssignmentType == "required" {
			result.RequiredTotal++
			if course.LessonCount > 0 && course.CompletedLessonCount == course.LessonCount {
				result.RequiredCompleted++
			}
		}
	}
	for _, category := range categories {
		result.Categories = append(result.Categories, category)
	}
	sort.Slice(result.Categories, func(left, right int) bool {
		if result.Categories[left].Name == result.Categories[right].Name {
			return result.Categories[left].ID < result.Categories[right].ID
		}
		return result.Categories[left].Name < result.Categories[right].Name
	})
	return result, nil
}

func (service *LearnerOverviewService) GetRecent(
	ctx context.Context,
	offset, limit int,
) ([]RecentLearningItem, int64, error) {
	if _, _, err := learnerIdentity(ctx); err != nil {
		return nil, 0, err
	}
	if offset < 0 || limit < 1 {
		return nil, 0, errorsx.BadRequest("invalid pagination")
	}
	data, err := service.overview.Get(ctx, shanghaiStudyDate(service.now()))
	if err != nil {
		return nil, 0, errorsx.Internal("load recent learning failed")
	}
	items := make([]RecentLearningItem, 0, len(data.Courses))
	for _, item := range data.Courses {
		if item.RecentLesson == nil || item.LastLearnedAt == nil {
			continue
		}
		course := presentLearnerCourse(item)
		items = append(items, RecentLearningItem{
			Course:              course.Course,
			RecentLesson:        *course.RecentLesson,
			ProgressPercent:     course.ProgressPercent,
			LastPositionSeconds: item.LastPositionSeconds,
			LastLearnedAt:       *item.LastLearnedAt,
		})
	}
	sort.Slice(items, func(left, right int) bool {
		if items[left].LastLearnedAt.Equal(items[right].LastLearnedAt) {
			return items[left].Course.ID < items[right].Course.ID
		}
		return items[left].LastLearnedAt.After(items[right].LastLearnedAt)
	})
	total := int64(len(items))
	if offset >= len(items) {
		return []RecentLearningItem{}, total, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], total, nil
}

func presentLearnerCourse(item repository.LearnerOverviewCourse) LearnerCourseSummary {
	course := LearnerCourseSummary{
		Course: CourseSummary{
			ID: item.Course.ID, Title: item.Course.Title,
			Description: item.Course.Description, CoverImage: item.Course.CoverImage,
		},
		AssignmentType: item.AssignmentType, LessonCount: item.LessonCount,
		CompletedLessonCount: item.CompletedLessonCount,
		ProgressPercent:      item.ProgressPercent,
		LastLearnedAt:        item.LastLearnedAt,
	}
	if item.Category != nil {
		course.Course.Category = &CourseCategorySummary{
			ID: item.Category.ID, Name: item.Category.Name,
		}
	}
	if item.RecentLesson != nil {
		course.RecentLesson = &LessonSummary{
			ID: item.RecentLesson.ID, Title: item.RecentLesson.Title,
			DurationSeconds:     item.RecentLesson.DurationSeconds,
			LastPositionSeconds: item.LastPositionSeconds,
		}
	}
	return course
}

func shanghaiStudyDate(now time.Time) string {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return now.In(location).Format("2006-01-02")
}
