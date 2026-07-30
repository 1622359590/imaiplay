package service

import (
	"context"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/repository"
)

type courseFixture struct {
	courses   *CourseService
	chapters  *CourseChapterService
	lessons   *CourseLessonService
	resources repository.ResourceRepository
}

func TestCourseServiceCreateOfficialUsesRequestedStatus(t *testing.T) {
	fixture := newCourseFixture(t)
	root := courseContext("root", "", "superadmin")
	draft, err := fixture.courses.CreateOfficial(
		root, "Official draft", "", "", 0,
	)
	if err != nil || draft.Status != 0 || !draft.IsOfficial ||
		draft.TenantID != "" {
		t.Fatalf("CreateOfficial(draft) = %#v, %v", draft, err)
	}
	published, err := fixture.courses.CreateOfficial(
		root, "Official published", "", "", 1,
	)
	if err != nil || published.Status != 1 {
		t.Fatalf("CreateOfficial(published) = %#v, %v", published, err)
	}
	if _, err := fixture.courses.CreateOfficial(
		root, "Invalid", "", "", 2,
	); errorCode(err) != 40000 {
		t.Fatalf("CreateOfficial(invalid status) error = %#v", err)
	}
}

func TestCourseServicePermissionsDetailAndPublishedList(t *testing.T) {
	fixture := newCourseFixture(t)
	admin := courseContext("admin", "tenant-1", "tenant_admin")
	author := courseContext("author-1", "tenant-1", "instructor")
	otherAuthor := courseContext("author-2", "tenant-1", "instructor")
	learner := courseContext("learner", "tenant-1", "learner")

	draft, err := fixture.courses.Create(author, "Draft", "Desc", "")
	if err != nil {
		t.Fatalf("Create(draft) error = %v", err)
	}
	published, err := fixture.courses.Create(admin, "Published", "Desc", "")
	if err != nil {
		t.Fatalf("Create(published) error = %v", err)
	}
	published, err = fixture.courses.Update(
		admin, published.ID, "Published", "Desc", "", 1,
	)
	if err != nil {
		t.Fatalf("Update(published) error = %v", err)
	}
	items, total, err := fixture.courses.List(admin, 0, 20)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("admin List() = %#v, %d, %v", items, total, err)
	}
	items, total, err = fixture.courses.List(author, 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != draft.ID {
		t.Fatalf("instructor List() = %#v, %d, %v", items, total, err)
	}
	if _, err := fixture.courses.Get(otherAuthor, draft.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor Get() error = %#v", err)
	}
	chapter, err := fixture.chapters.Create(author, draft.ID, "Chapter", 1)
	if err != nil {
		t.Fatalf("Create(chapter) error = %v", err)
	}
	if _, err := fixture.lessons.Create(
		author, chapter.ID, "Lesson", "video", "/video.mp4", 90, 1,
	); err != nil {
		t.Fatalf("Create(lesson) error = %v", err)
	}
	detail, err := fixture.courses.GetDetail(author, draft.ID)
	if err != nil || len(detail.Chapters) != 1 ||
		len(detail.Chapters[0].Lessons) != 1 {
		t.Fatalf("GetDetail() = %#v, %v", detail, err)
	}
	items, total, err = fixture.courses.ListPublished(learner, 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != published.ID {
		t.Fatalf("ListPublished() = %#v, %d, %v", items, total, err)
	}
	if _, err := fixture.courses.GetPublishedDetail(learner, draft.ID); errorCode(err) != 40400 {
		t.Fatalf("GetPublishedDetail(draft) error = %#v", err)
	}
}

func newCourseFixture(t *testing.T) courseFixture {
	t.Helper()
	database, _, _ := serviceRepositories(t)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	courses := NewCourseService(courseRepo, chapterRepo, lessonRepo)
	return courseFixture{
		courses:  courses,
		chapters: NewCourseChapterService(chapterRepo, courseRepo),
		lessons: NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo, resourceRepo,
		),
		resources: resourceRepo,
	}
}

func courseContext(userID, tenantID, role string) context.Context {
	return usercontext.WithUser(context.Background(), userID, tenantID, "", role)
}
