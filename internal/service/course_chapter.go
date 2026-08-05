package service

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type CourseChapterService struct {
	chapters repository.CourseChapterRepository
	courses  repository.CourseRepository
}

func NewCourseChapterService(
	chapters repository.CourseChapterRepository,
	courses repository.CourseRepository,
) *CourseChapterService {
	return &CourseChapterService{chapters: chapters, courses: courses}
}

func (service *CourseChapterService) Create(
	ctx context.Context, courseID, title string, sortOrder int,
) (*domain.CourseChapter, error) {
	_, tenantID, err := service.authorizeCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		CourseID:  courseID, Title: title, SortOrder: sortOrder,
	}
	if err := service.chapters.Create(ctx, chapter); err != nil {
		return nil, errorsx.Internal("create chapter failed")
	}
	return chapter, nil
}

func (service *CourseChapterService) List(
	ctx context.Context, courseID string,
) ([]domain.CourseChapter, error) {
	if _, _, err := service.authorizeCourse(ctx, courseID); err != nil {
		return nil, err
	}
	items, err := service.chapters.FindByCourse(ctx, courseID)
	if err != nil {
		return nil, errorsx.Internal("list chapters failed")
	}
	return items, nil
}

func (service *CourseChapterService) Update(
	ctx context.Context, id, title string, sortOrder int,
) (*domain.CourseChapter, error) {
	chapter, err := service.get(ctx, id)
	if err != nil {
		return nil, err
	}
	chapter.Title, chapter.SortOrder = title, sortOrder
	if err := service.chapters.Update(ctx, chapter); err != nil {
		return nil, mapNotFound(err, "chapter not found")
	}
	return chapter, nil
}

func (service *CourseChapterService) Delete(ctx context.Context, id string) error {
	if _, err := service.get(ctx, id); err != nil {
		return err
	}
	return mapNotFound(service.chapters.Delete(ctx, id), "chapter not found")
}

func (service *CourseChapterService) get(
	ctx context.Context, id string,
) (*domain.CourseChapter, error) {
	chapter, err := service.chapters.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "chapter not found")
	}
	if _, _, err := service.authorizeCourse(ctx, chapter.CourseID); err != nil {
		return nil, err
	}
	return chapter, nil
}

func (service *CourseChapterService) authorizeCourse(
	ctx context.Context, courseID string,
) (*domain.Course, string, error) {
	course, err := requireManageableCourse(ctx, service.courses, courseID)
	if err != nil {
		return nil, "", err
	}
	return course, course.TenantID, nil
}
