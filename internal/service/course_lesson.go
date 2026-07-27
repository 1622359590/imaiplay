package service

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type CourseLessonService struct {
	lessons  repository.CourseLessonRepository
	chapters repository.CourseChapterRepository
	courses  repository.CourseRepository
}

func NewCourseLessonService(
	lessons repository.CourseLessonRepository,
	chapters repository.CourseChapterRepository,
	courses repository.CourseRepository,
) *CourseLessonService {
	return &CourseLessonService{
		lessons: lessons, chapters: chapters, courses: courses,
	}
}

func (service *CourseLessonService) Create(
	ctx context.Context,
	chapterID, title, contentType, contentURL string,
	durationSeconds, sortOrder int,
) (*domain.CourseLesson, error) {
	return service.CreateWithResource(ctx, chapterID, title, contentType, "", contentURL, durationSeconds, sortOrder)
}

func (service *CourseLessonService) CreateWithResource(
	ctx context.Context,
	chapterID, title, contentType, resourceID, contentURL string,
	durationSeconds, sortOrder int,
) (*domain.CourseLesson, error) {
	_, tenantID, err := service.authorizeChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	if !validContentType(contentType) {
		return nil, errorsx.BadRequest("invalid content type")
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		ChapterID: chapterID, Title: title, ContentType: contentType,
		ResourceID: nullableResourceID(resourceID),
		ContentURL: contentURL, DurationSeconds: durationSeconds,
		SortOrder: sortOrder,
	}
	if err := service.lessons.Create(ctx, lesson); err != nil {
		return nil, errorsx.Internal("create lesson failed")
	}
	return lesson, nil
}

func (service *CourseLessonService) List(
	ctx context.Context, chapterID string,
) ([]domain.CourseLesson, error) {
	if _, _, err := service.authorizeChapter(ctx, chapterID); err != nil {
		return nil, err
	}
	items, err := service.lessons.FindByChapter(ctx, chapterID)
	if err != nil {
		return nil, errorsx.Internal("list lessons failed")
	}
	return items, nil
}

func (service *CourseLessonService) Update(
	ctx context.Context,
	id, title, contentType, contentURL string,
	durationSeconds, sortOrder int,
) (*domain.CourseLesson, error) {
	return service.UpdateWithResource(ctx, id, title, contentType, "", contentURL, durationSeconds, sortOrder)
}

func (service *CourseLessonService) UpdateWithResource(
	ctx context.Context,
	id, title, contentType, resourceID, contentURL string,
	durationSeconds, sortOrder int,
) (*domain.CourseLesson, error) {
	if !validContentType(contentType) {
		return nil, errorsx.BadRequest("invalid content type")
	}
	lesson, err := service.get(ctx, id)
	if err != nil {
		return nil, err
	}
	lesson.Title, lesson.ContentType = title, contentType
	lesson.ResourceID = nullableResourceID(resourceID)
	lesson.ContentURL, lesson.DurationSeconds = contentURL, durationSeconds
	lesson.SortOrder = sortOrder
	if err := service.lessons.Update(ctx, lesson); err != nil {
		return nil, mapNotFound(err, "lesson not found")
	}
	return lesson, nil
}

func nullableResourceID(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (service *CourseLessonService) Delete(ctx context.Context, id string) error {
	if _, err := service.get(ctx, id); err != nil {
		return err
	}
	return mapNotFound(service.lessons.Delete(ctx, id), "lesson not found")
}

func (service *CourseLessonService) get(
	ctx context.Context, id string,
) (*domain.CourseLesson, error) {
	if _, _, _, err := courseManager(ctx); err != nil {
		return nil, err
	}
	lesson, err := service.lessons.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "lesson not found")
	}
	if _, _, err := service.authorizeChapter(ctx, lesson.ChapterID); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (service *CourseLessonService) authorizeChapter(
	ctx context.Context, chapterID string,
) (*domain.CourseChapter, string, error) {
	_, tenantID, _, err := courseManager(ctx)
	if err != nil {
		return nil, "", err
	}
	chapter, err := service.chapters.FindByID(ctx, chapterID)
	if err != nil {
		return nil, "", mapNotFound(err, "chapter not found")
	}
	if _, err := service.courses.FindByID(ctx, chapter.CourseID); err != nil {
		return nil, "", mapNotFound(err, "course not found")
	}
	return chapter, tenantID, nil
}

func validContentType(contentType string) bool {
	return contentType == "video" ||
		contentType == "document" ||
		contentType == "text"
}
