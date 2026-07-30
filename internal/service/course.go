package service

import (
	"context"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
)

type CourseChapterDetail struct {
	domain.CourseChapter
	Lessons []domain.CourseLesson `json:"lessons"`
}

type CourseDetail struct {
	Course   domain.Course         `json:"course"`
	Chapters []CourseChapterDetail `json:"chapters"`
}

type CourseService struct {
	courses  repository.CourseRepository
	chapters repository.CourseChapterRepository
	lessons  repository.CourseLessonRepository
}

func NewCourseService(
	courses repository.CourseRepository,
	chapters repository.CourseChapterRepository,
	lessons repository.CourseLessonRepository,
) *CourseService {
	return &CourseService{courses: courses, chapters: chapters, lessons: lessons}
}

func (service *CourseService) Create(
	ctx context.Context, title, description, coverImage string,
) (*domain.Course, error) {
	userID, tenantID, _, err := courseManager(ctx)
	if err != nil {
		return nil, err
	}
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Title:     title, Description: description, CoverImage: coverImage,
		CreatedBy: userID, Status: 0,
	}
	if err := service.courses.Create(ctx, course); err != nil {
		return nil, errorsx.Internal("create course failed")
	}
	return course, nil
}

func (service *CourseService) CreateOfficial(ctx context.Context, title, description, coverImage string) (*domain.Course, error) {
	userID, _, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "superadmin" {
		return nil, errorsx.Forbidden("permission denied")
	}
	course := &domain.Course{Title: title, Description: description, CoverImage: coverImage, CreatedBy: userID, Status: 1, IsOfficial: true}
	if err := service.courses.Create(ctx, course); err != nil {
		return nil, errorsx.Internal("create official course failed")
	}
	return course, nil
}

func (service *CourseService) List(
	ctx context.Context, offset, limit int,
) ([]domain.Course, int64, error) {
	_, tenantID, _, err := courseManager(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := service.courses.FindByTenant(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, 0, errorsx.Internal("list courses failed")
	}
	return items, total, nil
}

func (service *CourseService) Get(
	ctx context.Context, id string,
) (*domain.Course, error) {
	if _, _, _, err := courseManager(ctx); err != nil {
		return nil, err
	}
	course, err := service.courses.FindByID(ctx, id)
	return course, mapNotFound(err, "course not found")
}

func (service *CourseService) Update(
	ctx context.Context,
	id, title, description, coverImage string,
	status int,
) (*domain.Course, error) {
	if status != 0 && status != 1 {
		return nil, errorsx.BadRequest("invalid course status")
	}
	course, err := service.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if course.IsOfficial && !superadmin(ctx) {
		return nil, errorsx.Forbidden("permission denied")
	}
	course.Title, course.Description = title, description
	course.CoverImage, course.Status = coverImage, status
	if err := service.courses.Update(ctx, course); err != nil {
		return nil, mapNotFound(err, "course not found")
	}
	return course, nil
}

func (service *CourseService) Delete(ctx context.Context, id string) error {
	if _, _, _, err := courseManager(ctx); err != nil {
		return err
	}
	if course, err := service.courses.FindByID(ctx, id); err != nil {
		return mapNotFound(err, "course not found")
	} else if course.IsOfficial && !superadmin(ctx) {
		return errorsx.Forbidden("permission denied")
	}
	return mapNotFound(service.courses.Delete(ctx, id), "course not found")
}

func superadmin(ctx context.Context) bool {
	_, _, _, role, ok := usercontext.UserFromContext(ctx)
	return ok && role == "superadmin"
}

func (service *CourseService) OfficialList(ctx context.Context, offset, limit int) ([]domain.Course, int64, error) {
	if _, tenantID, _, role, ok := usercontext.UserFromContext(ctx); !ok || (role != "superadmin" && (role != "tenant_admin" || tenantID == "")) {
		return nil, 0, errorsx.Forbidden("permission denied")
	}
	return service.courses.FindOfficial(ctx, offset, limit)
}

func (service *CourseService) EnableOfficial(ctx context.Context, courseID string, enabled bool) error {
	_, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" || role != "tenant_admin" {
		return errorsx.Forbidden("permission denied")
	}
	if err := service.courses.ActivateOfficial(ctx, tenantID, courseID, enabled); err != nil {
		return mapNotFound(err, "official course not found")
	}
	return nil
}

func (service *CourseService) GetDetail(
	ctx context.Context, id string,
) (*CourseDetail, error) {
	course, err := service.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return service.detail(ctx, course)
}

func (service *CourseService) ListPublished(
	ctx context.Context, offset, limit int,
) ([]domain.Course, int64, error) {
	tenantID, err := authenticatedTenant(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := service.courses.FindPublishedByTenant(
		ctx, tenantID, offset, limit,
	)
	if err != nil {
		return nil, 0, errorsx.Internal("list courses failed")
	}
	return items, total, nil
}

func (service *CourseService) GetPublishedDetail(
	ctx context.Context, id string,
) (*CourseDetail, error) {
	tenantID, err := authenticatedTenant(ctx)
	if err != nil {
		return nil, err
	}
	course, err := service.courses.FindPublishedByID(ctx, tenantID, id)
	if err != nil {
		return nil, mapNotFound(err, "course not found")
	}
	return service.detail(ctx, course)
}

func (service *CourseService) detail(
	ctx context.Context, course *domain.Course,
) (*CourseDetail, error) {
	chapters, err := service.chapters.FindByCourse(ctx, course.ID)
	if err != nil {
		return nil, errorsx.Internal("list chapters failed")
	}
	detail := &CourseDetail{Course: *course, Chapters: []CourseChapterDetail{}}
	for _, chapter := range chapters {
		lessons, err := service.lessons.FindByChapter(ctx, chapter.ID)
		if err != nil {
			return nil, errorsx.Internal("list lessons failed")
		}
		for index := range lessons {
			if lessons[index].ResourceID != nil {
				lessons[index].ResourceType = lessons[index].ContentType
			}
		}
		detail.Chapters = append(detail.Chapters, CourseChapterDetail{
			CourseChapter: chapter, Lessons: lessons,
		})
	}
	return detail, nil
}

func courseManager(ctx context.Context) (string, string, string, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || (tenantID == "" && role != "superadmin") || (role != "tenant_admin" && role != "instructor" && role != "superadmin") {
		return "", "", "", errorsx.Forbidden("permission denied")
	}
	return userID, tenantID, role, nil
}

func authenticatedTenant(ctx context.Context) (string, error) {
	_, tenantID, _, _, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" {
		return "", errorsx.Unauthorized("authentication required")
	}
	return tenantID, nil
}
