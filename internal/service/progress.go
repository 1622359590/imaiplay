package service

import (
	"context"
	"errors"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type RecentLearnItem struct {
	Course        domain.Course         `json:"course"`
	Lesson        domain.CourseLesson   `json:"lesson"`
	Progress      domain.LessonProgress `json:"progress"`
	LastLearnedAt time.Time             `json:"last_learned_at"`
}

type ProgressService struct {
	progress    repository.LessonProgressRepository
	enrollments repository.CourseEnrollmentRepository
	lessons     repository.CourseLessonRepository
	chapters    repository.CourseChapterRepository
	courses     repository.CourseRepository
}

func NewProgressService(
	progress repository.LessonProgressRepository,
	enrollments repository.CourseEnrollmentRepository,
	lessons repository.CourseLessonRepository,
	chapters repository.CourseChapterRepository,
	courses repository.CourseRepository,
) *ProgressService {
	return &ProgressService{
		progress: progress, enrollments: enrollments, lessons: lessons,
		chapters: chapters, courses: courses,
	}
}

func (service *ProgressService) Report(
	ctx context.Context, lessonID string, positionSeconds, percent int,
) (*domain.LessonProgress, error) {
	userID, tenantID, err := learnerIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if positionSeconds < 0 || percent < 0 || percent > 100 {
		return nil, errorsx.BadRequest("invalid progress")
	}
	_, course, err := service.lessonCourse(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	enrollment, err := service.enrollments.FindByCourseAndUser(
		ctx, course.ID, userID,
	)
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err == nil && enrollment.Status != 1) {
		return nil, errorsx.Forbidden("not enrolled in this course")
	}
	if err != nil {
		return nil, errorsx.Internal("find enrollment failed")
	}
	progress, err := service.progress.FindByUserAndLesson(ctx, userID, lessonID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		progress = &domain.LessonProgress{
			BaseModel: domain.BaseModel{TenantID: tenantID},
			UserID:    userID, LessonID: lessonID,
		}
	} else if err != nil {
		return nil, errorsx.Internal("find progress failed")
	}
	progress.ProgressPercent = percent
	progress.LastPositionSeconds = positionSeconds
	progress.Status = progressStatus(percent)
	progress.CompletedAt = nil
	if percent == 100 {
		now := time.Now().UTC()
		progress.CompletedAt = &now
	}
	if err := service.progress.Upsert(ctx, progress); err != nil {
		return nil, errorsx.Internal("save progress failed")
	}
	return progress, nil
}

func (service *ProgressService) Get(
	ctx context.Context, lessonID string,
) (*domain.LessonProgress, error) {
	userID, _, err := learnerIdentity(ctx)
	if err != nil {
		return nil, err
	}
	progress, err := service.progress.FindByUserAndLesson(ctx, userID, lessonID)
	return progress, mapNotFound(err, "progress not found")
}

func (service *ProgressService) GetRecent(
	ctx context.Context, offset, limit int,
) ([]RecentLearnItem, int64, error) {
	userID, _, err := learnerIdentity(ctx)
	if err != nil {
		return nil, 0, err
	}
	progressItems, total, err := service.progress.FindByUser(
		ctx, userID, offset, limit,
	)
	if err != nil {
		return nil, 0, errorsx.Internal("list progress failed")
	}
	items := make([]RecentLearnItem, 0, len(progressItems))
	for _, progress := range progressItems {
		lesson, course, err := service.lessonCourse(ctx, progress.LessonID)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, RecentLearnItem{
			Course: *course, Lesson: *lesson, Progress: progress,
			LastLearnedAt: progress.UpdatedAt,
		})
	}
	return items, total, nil
}

func (service *ProgressService) lessonCourse(
	ctx context.Context, lessonID string,
) (*domain.CourseLesson, *domain.Course, error) {
	lesson, err := service.lessons.FindByID(ctx, lessonID)
	if err != nil {
		return nil, nil, mapNotFound(err, "lesson not found")
	}
	chapter, err := service.chapters.FindByID(ctx, lesson.ChapterID)
	if err != nil {
		return nil, nil, mapNotFound(err, "chapter not found")
	}
	course, err := service.courses.FindByID(ctx, chapter.CourseID)
	if err != nil {
		return nil, nil, mapNotFound(err, "course not found")
	}
	return lesson, course, nil
}

func learnerIdentity(ctx context.Context) (string, string, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || userID == "" || tenantID == "" {
		return "", "", errorsx.Forbidden("permission denied")
	}
	return userID, tenantID, nil
}

func progressStatus(percent int) int {
	if percent == 0 {
		return 0
	}
	if percent == 100 {
		return 2
	}
	return 1
}
