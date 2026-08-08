package service

import (
	"context"
	"errors"
	"strings"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type ProgressService struct {
	progress     repository.LessonProgressRepository
	enrollments  repository.CourseEnrollmentRepository
	lessons      repository.CourseLessonRepository
	chapters     repository.CourseChapterRepository
	courses      repository.CourseRepository
	learningTime repository.LearningTimeRepository
	now          func() time.Time
}

func NewProgressService(
	progress repository.LessonProgressRepository,
	enrollments repository.CourseEnrollmentRepository,
	lessons repository.CourseLessonRepository,
	chapters repository.CourseChapterRepository,
	courses repository.CourseRepository,
	learningTime ...repository.LearningTimeRepository,
) *ProgressService {
	service := &ProgressService{
		progress: progress, enrollments: enrollments, lessons: lessons,
		chapters: chapters, courses: courses, now: time.Now,
	}
	if len(learningTime) > 0 {
		service.learningTime = learningTime[0]
	}
	return service
}

func (service *ProgressService) Report(
	ctx context.Context, lessonID string, positionSeconds, percent int,
	watchedSecondsDelta int, reportID string,
) (*domain.LessonProgress, error) {
	userID, tenantID, err := learnerIdentity(ctx)
	if err != nil {
		return nil, err
	}
	legacyReport := watchedSecondsDelta == 0 && strings.TrimSpace(reportID) == ""
	if positionSeconds < 0 || percent < 0 || percent > 100 ||
		(!legacyReport && (watchedSecondsDelta < 1 || watchedSecondsDelta > 60 || strings.TrimSpace(reportID) == "")) {
		return nil, errorsx.BadRequest("invalid progress")
	}
	_, course, err := service.lessonCourse(ctx, tenantID, lessonID)
	if err != nil {
		return nil, err
	}
	enrollment, err := service.enrollments.FindByCourseAndUser(
		ctx, course.ID, userID,
	)
	if errors.Is(err, gorm.ErrRecordNotFound) && course.IsOfficial {
		if err := service.startCourse(ctx, course, userID, tenantID); err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err == nil && enrollment.Status != 1) {
		return nil, errorsx.Forbidden("not enrolled in this course")
	} else if err != nil {
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
	reportedAt := service.now()
	progress.ProgressPercent = percent
	progress.LastPositionSeconds = positionSeconds
	progress.Status = progressStatus(percent)
	progress.CompletedAt = nil
	if percent == 100 {
		now := reportedAt.UTC()
		progress.CompletedAt = &now
	}
	if err := service.progress.Upsert(ctx, progress); err != nil {
		return nil, errorsx.Internal("save progress failed")
	}
	if !legacyReport {
		if service.learningTime == nil {
			return nil, errorsx.Internal("record learning time failed")
		}
		_, err := service.learningTime.Record(ctx, &domain.LearningTimeReport{
			LessonID: lessonID, ReportID: reportID,
			WatchedSecondsDelta: watchedSecondsDelta,
		}, shanghaiStudyDate(reportedAt))
		if err != nil {
			if errors.Is(err, repository.ErrInvalidLearningTimeReport) {
				return nil, errorsx.BadRequest("invalid progress")
			}
			return nil, errorsx.Internal("record learning time failed")
		}
	}
	return progress, nil
}

func (service *ProgressService) Get(
	ctx context.Context, lessonID string,
) (*domain.LessonProgress, error) {
	userID, tenantID, err := learnerIdentity(ctx)
	if err != nil {
		return nil, err
	}
	_, course, err := service.lessonCourse(ctx, tenantID, lessonID)
	if err != nil {
		return nil, err
	}
	if err := service.startCourse(ctx, course, userID, tenantID); err != nil {
		return nil, err
	}
	progress, err := service.progress.FindByUserAndLesson(ctx, userID, lessonID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &domain.LessonProgress{
			BaseModel: domain.BaseModel{TenantID: tenantID},
			UserID:    userID, LessonID: lessonID,
		}, nil
	}
	if err != nil {
		return nil, errorsx.Internal("find progress failed")
	}
	return progress, nil
}

func (service *ProgressService) startCourse(
	ctx context.Context, course *domain.Course, userID, tenantID string,
) error {
	enrollment, err := service.enrollments.FindByCourseAndUser(ctx, course.ID, userID)
	if err == nil {
		if enrollment.Status != 1 {
			return errorsx.Forbidden("not enrolled in this course")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errorsx.Internal("find enrollment failed")
	}
	assignmentType := course.CourseType
	if assignmentType == "" {
		assignmentType = domain.AssignmentRequired
	}
	createErr := service.enrollments.Create(ctx, &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		CourseID:  course.ID, UserID: userID, Status: 1,
		AssignmentType: assignmentType,
	})
	if createErr == nil {
		return nil
	}

	// A concurrent first read may have created the same enrollment after our
	// initial lookup. Re-read before classifying the create error so that this
	// normal race remains idempotent.
	enrollment, err = service.enrollments.FindByCourseAndUser(ctx, course.ID, userID)
	if err == nil {
		if enrollment.Status != 1 {
			return errorsx.Forbidden("not enrolled in this course")
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errorsx.Internal("find enrollment failed")
	}
	if errors.Is(createErr, gorm.ErrRecordNotFound) {
		return errorsx.NotFound("course not found")
	}
	return errorsx.Internal("start course failed")
}

func (service *ProgressService) lessonCourse(
	ctx context.Context, tenantID, lessonID string,
) (*domain.CourseLesson, *domain.Course, error) {
	lesson, err := service.lessons.FindByID(ctx, lessonID)
	if err != nil {
		return nil, nil, mapNotFound(err, "lesson not found")
	}
	chapter, err := service.chapters.FindByID(ctx, lesson.ChapterID)
	if err != nil {
		return nil, nil, mapNotFound(err, "chapter not found")
	}
	course, err := service.courses.FindPublishedByID(ctx, tenantID, chapter.CourseID)
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
