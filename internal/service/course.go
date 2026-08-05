package service

import (
	"context"
	"errors"
	"net"
	"net/url"
	"sort"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

type CourseChapterDetail struct {
	domain.CourseChapter
	Lessons []domain.CourseLesson `json:"lessons"`
}

type CourseDetail struct {
	Course    domain.Course           `json:"course"`
	Chapters  []CourseChapterDetail   `json:"chapters"`
	Materials []domain.CourseMaterial `json:"materials"`
}

type LearnerLessonDetail struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	ContentType     string  `json:"content_type"`
	ResourceID      *string `json:"resource_id,omitempty"`
	ResourceType    string  `json:"resource_type,omitempty"`
	ContentURL      string  `json:"content_url"`
	DurationSeconds int     `json:"duration_seconds"`
	SortOrder       int     `json:"sort_order"`
}

type LearnerCourseChapterDetail struct {
	ID        string                `json:"id"`
	Title     string                `json:"title"`
	SortOrder int                   `json:"sort_order"`
	Lessons   []LearnerLessonDetail `json:"lessons"`
}

type LearnerCourseMaterial struct {
	ID           string `json:"id"`
	DisplayName  string `json:"display_name"`
	ResourceType string `json:"resource_type"`
	SizeBytes    int64  `json:"size_bytes"`
}

type LearnerCourseDetail struct {
	Course    domain.Course                `json:"course"`
	Chapters  []LearnerCourseChapterDetail `json:"chapters"`
	Materials []LearnerCourseMaterial      `json:"materials"`
}

type CourseService struct {
	courses     repository.CourseRepository
	chapters    repository.CourseChapterRepository
	lessons     repository.CourseLessonRepository
	enrollments repository.CourseEnrollmentRepository
	materials   repository.CourseMaterialRepository
	access      *LearnerAccess
}

func NewCourseService(
	courses repository.CourseRepository,
	chapters repository.CourseChapterRepository,
	lessons repository.CourseLessonRepository,
	enrollments repository.CourseEnrollmentRepository,
	materials ...repository.CourseMaterialRepository,
) *CourseService {
	service := &CourseService{
		courses: courses, chapters: chapters, lessons: lessons,
		enrollments: enrollments,
	}
	if len(materials) > 0 {
		service.materials = materials[0]
	}
	service.access = NewLearnerAccess(courses, enrollments, service.materials)
	return service
}

func (service *CourseService) Create(
	ctx context.Context, title, description, coverImage string,
) (*domain.Course, error) {
	userID, tenantID, role, err := courseManager(ctx)
	if err != nil {
		return nil, err
	}
	if role == "superadmin" {
		return nil, errorsx.Forbidden("permission denied")
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

func (service *CourseService) CreateOfficial(
	ctx context.Context,
	title, description, coverImage string,
	status int,
) (*domain.Course, error) {
	userID, _, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "superadmin" {
		return nil, errorsx.Forbidden("permission denied")
	}
	if status != 0 && status != 1 {
		return nil, errorsx.BadRequest("invalid course status")
	}
	course := &domain.Course{
		Title: title, Description: description, CoverImage: coverImage,
		CreatedBy: userID, Status: status, IsOfficial: true,
	}
	if err := service.courses.Create(ctx, course); err != nil {
		return nil, errorsx.Internal("create official course failed")
	}
	return course, nil
}

func (service *CourseService) List(
	ctx context.Context, offset, limit int,
) ([]domain.Course, int64, error) {
	userID, tenantID, role, err := courseManager(ctx)
	if err != nil {
		return nil, 0, err
	}
	var items []domain.Course
	var total int64
	switch role {
	case "superadmin":
		items, total, err = service.courses.FindOfficial(ctx, offset, limit)
	case "instructor":
		items, total, err = service.courses.FindByTenantAndCreator(ctx, tenantID, userID, offset, limit)
	default:
		items, total, err = service.courses.FindByTenant(ctx, tenantID, offset, limit)
	}
	if err != nil {
		return nil, 0, errorsx.Internal("list courses failed")
	}
	return items, total, nil
}

func (service *CourseService) Get(
	ctx context.Context, id string,
) (*domain.Course, error) {
	return requireManageableCourse(ctx, service.courses, id)
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
	course.Title, course.Description = title, description
	course.CoverImage, course.Status = coverImage, status
	if err := service.courses.Update(ctx, course); err != nil {
		return nil, mapNotFound(err, "course not found")
	}
	return course, nil
}

func (service *CourseService) Delete(ctx context.Context, id string) error {
	if _, err := requireManageableCourse(ctx, service.courses, id); err != nil {
		return err
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
	userID, tenantID, err := learnerIdentity(ctx)
	if err != nil {
		return nil, 0, err
	}
	if offset < 0 || limit < 1 {
		return nil, 0, errorsx.BadRequest("invalid pagination")
	}
	if service.enrollments == nil {
		return nil, 0, errorsx.Internal("list courses failed")
	}
	enrollments, err := service.enrollments.FindByUser(ctx, userID)
	if err != nil {
		return nil, 0, errorsx.Internal("list courses failed")
	}
	items := make([]domain.Course, 0, len(enrollments))
	for _, enrollment := range enrollments {
		if enrollment.Status != 1 {
			continue
		}
		course, err := service.courses.FindPublishedByID(ctx, tenantID, enrollment.CourseID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return nil, 0, errorsx.Internal("list courses failed")
		}
		items = append(items, *course)
	}
	sort.Slice(items, func(left, right int) bool {
		if items[left].Title == items[right].Title {
			return items[left].ID < items[right].ID
		}
		return items[left].Title < items[right].Title
	})
	total := int64(len(items))
	if offset >= len(items) {
		return []domain.Course{}, total, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], total, nil
}

func (service *CourseService) GetPublishedDetail(
	ctx context.Context, id string,
) (*LearnerCourseDetail, error) {
	if service.access == nil {
		return nil, errorsx.Internal("find course failed")
	}
	course, err := service.access.AuthorizeCourse(ctx, id)
	if err != nil {
		return nil, err
	}
	detail, err := service.detail(ctx, course)
	if err != nil {
		return nil, err
	}
	return learnerCourseDetail(detail), nil
}

func learnerCourseDetail(detail *CourseDetail) *LearnerCourseDetail {
	result := &LearnerCourseDetail{
		Course:    detail.Course,
		Chapters:  make([]LearnerCourseChapterDetail, 0, len(detail.Chapters)),
		Materials: make([]LearnerCourseMaterial, 0, len(detail.Materials)),
	}
	for _, chapter := range detail.Chapters {
		learnerChapter := LearnerCourseChapterDetail{
			ID: chapter.ID, Title: chapter.Title, SortOrder: chapter.SortOrder,
			Lessons: make([]LearnerLessonDetail, 0, len(chapter.Lessons)),
		}
		for _, lesson := range chapter.Lessons {
			contentURL := learnerContentURL(lesson)
			learnerChapter.Lessons = append(learnerChapter.Lessons, LearnerLessonDetail{
				ID: lesson.ID, Title: lesson.Title, ContentType: lesson.ContentType,
				ResourceID: lesson.ResourceID, ResourceType: lesson.ResourceType,
				ContentURL: contentURL, DurationSeconds: lesson.DurationSeconds,
				SortOrder: lesson.SortOrder,
			})
		}
		result.Chapters = append(result.Chapters, learnerChapter)
	}
	for _, material := range detail.Materials {
		result.Materials = append(result.Materials, LearnerCourseMaterial{
			ID: material.ID, DisplayName: material.DisplayName,
			ResourceType: material.Resource.ResourceType,
			SizeBytes:    material.Resource.SizeBytes,
		})
	}
	return result
}

func learnerContentURL(lesson domain.CourseLesson) string {
	if lesson.ResourceID != nil {
		return ""
	}
	if lesson.ContentType == "text" {
		return lesson.ContentURL
	}
	if safePublicContentURL(lesson.ContentURL) {
		return lesson.ContentURL
	}
	return ""
}

func safePublicContentURL(raw string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || parsed.Host == "" ||
		(parsed.Scheme != "https" && parsed.Scheme != "http") {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return false
	}
	return true
}

func (service *CourseService) detail(
	ctx context.Context, course *domain.Course,
) (*CourseDetail, error) {
	chapters, err := service.chapters.FindByCourse(ctx, course.ID)
	if err != nil {
		return nil, errorsx.Internal("list chapters failed")
	}
	detail := &CourseDetail{Course: *course, Chapters: []CourseChapterDetail{}, Materials: []domain.CourseMaterial{}}
	if service.materials != nil {
		materials, err := service.materials.FindByCourse(ctx, course.ID)
		if err != nil {
			return nil, errorsx.Internal("list course materials failed")
		}
		detail.Materials = materials
	}
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

func requireManageableCourse(
	ctx context.Context, courses repository.CourseRepository, id string,
) (*domain.Course, error) {
	userID, tenantID, role, err := courseManager(ctx)
	if err != nil {
		return nil, err
	}
	course, err := courses.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.NotFound("course not found")
	}
	if err != nil {
		return nil, errorsx.Internal("find course failed")
	}
	allowed := false
	switch role {
	case "superadmin":
		allowed = course.IsOfficial && course.TenantID == ""
	case "tenant_admin":
		allowed = !course.IsOfficial && course.TenantID == tenantID
	case "instructor":
		allowed = !course.IsOfficial && course.TenantID == tenantID && course.CreatedBy == userID
	}
	if !allowed {
		return nil, errorsx.Forbidden("permission denied")
	}
	return course, nil
}
