package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/storage"
	"gorm.io/gorm"
)

type courseFixture struct {
	database    *gorm.DB
	courses     *CourseService
	chapters    *CourseChapterService
	lessons     *CourseLessonService
	enrollments repository.CourseEnrollmentRepository
	resources   repository.ResourceRepository
}

func TestCourseServicePublishedEndpointsRequireActiveVisibleAssignment(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	materialRepo := repository.NewCourseMaterialRepository(database)
	service := NewCourseService(
		courseRepo, chapterRepo, lessonRepo, enrollmentRepo, materialRepo,
	)
	ctx := courseContext("learner-1", "tenant-1", "learner")

	createCourse := func(id, title, tenantID string, status int, official bool) *domain.Course {
		course := &domain.Course{
			BaseModel: domain.BaseModel{ID: id, TenantID: tenantID},
			Title:     title, Status: status, CreatedBy: "admin", IsOfficial: official,
		}
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course %s: %v", title, err)
		}
		return course
	}
	assigned := createCourse("course-assigned", "A Assigned", "tenant-1", 1, false)
	official := createCourse("course-official", "B Official", "", 1, true)
	inactive := createCourse("course-inactive", "C Inactive", "tenant-1", 1, false)
	unassigned := createCourse("course-unassigned", "D Unassigned", "tenant-1", 1, false)
	draft := createCourse("course-draft", "E Draft", "tenant-1", 0, false)
	disabledOfficial := createCourse("course-official-disabled", "F Disabled", "", 1, true)
	foreign := createCourse("course-foreign", "G Foreign", "tenant-2", 1, false)

	activeEnrollment := func(course *domain.Course) *domain.CourseEnrollment {
		return &domain.CourseEnrollment{
			BaseModel: domain.BaseModel{TenantID: "tenant-1"},
			CourseID:  course.ID, UserID: "learner-1", Status: 1,
			AssignmentType: domain.AssignmentRequired,
		}
	}
	inactiveEnrollment := activeEnrollment(inactive)
	for _, enrollment := range []*domain.CourseEnrollment{
		activeEnrollment(assigned), activeEnrollment(official), inactiveEnrollment,
		activeEnrollment(draft), activeEnrollment(disabledOfficial), activeEnrollment(foreign),
	} {
		if err := database.Create(enrollment).Error; err != nil {
			t.Fatalf("create enrollment: %v", err)
		}
	}
	if err := database.Model(inactiveEnrollment).Update("status", 0).Error; err != nil {
		t.Fatalf("disable enrollment: %v", err)
	}
	for _, activation := range []*domain.TenantOfficialCourse{
		{TenantID: "tenant-1", CourseID: official.ID, Enabled: true},
		{TenantID: "tenant-1", CourseID: disabledOfficial.ID, Enabled: true},
	} {
		if err := database.Create(activation).Error; err != nil {
			t.Fatalf("create activation: %v", err)
		}
	}
	if err := database.Model(&domain.TenantOfficialCourse{}).
		Where("tenant_id = ? AND course_id = ?", "tenant-1", disabledOfficial.ID).
		Update("enabled", false).Error; err != nil {
		t.Fatalf("disable official: %v", err)
	}

	items, total, err := service.ListPublished(ctx, 0, 1)
	if err != nil || total != 2 || len(items) != 1 || items[0].ID != assigned.ID {
		t.Fatalf("ListPublished(first page) = %#v, %d, %v", items, total, err)
	}
	items, total, err = service.ListPublished(ctx, 1, 10)
	if err != nil || total != 2 || len(items) != 1 || items[0].ID != official.ID {
		t.Fatalf("ListPublished(second page) = %#v, %d, %v", items, total, err)
	}
	if _, err := service.GetPublishedDetail(ctx, assigned.ID); err != nil {
		t.Fatalf("GetPublishedDetail(assigned) error = %v", err)
	}
	for _, hidden := range []*domain.Course{inactive, unassigned, draft, disabledOfficial, foreign} {
		if _, err := service.GetPublishedDetail(ctx, hidden.ID); errorCode(err) != 40400 {
			t.Errorf("GetPublishedDetail(%s) error = %#v", hidden.Title, err)
		}
	}
}

func TestCourseServiceDetailIncludesOrderedMaterialsAndEmptySlice(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	materialRepo := repository.NewCourseMaterialRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	service := NewCourseService(
		courseRepo, chapterRepo, lessonRepo,
		repository.NewCourseEnrollmentRepository(database), materialRepo,
	)
	ctx := courseContext("admin", "tenant-1", "tenant_admin")
	course := &domain.Course{BaseModel: domain.BaseModel{ID: "course", TenantID: "tenant-1"}, Title: "Course", Status: 1, CreatedBy: "admin"}
	empty := &domain.Course{BaseModel: domain.BaseModel{ID: "empty", TenantID: "tenant-1"}, Title: "Empty", Status: 1, CreatedBy: "admin"}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	if err := database.Create(empty).Error; err != nil {
		t.Fatalf("create empty course: %v", err)
	}
	for _, resource := range []*domain.Resource{
		{BaseModel: domain.BaseModel{ID: "resource-1", TenantID: "tenant-1"}, Name: "one.pdf", ResourceType: "attachment", URL: "/uploads/one.pdf", CreatedBy: "admin"},
		{BaseModel: domain.BaseModel{ID: "resource-2", TenantID: "tenant-1"}, Name: "two.pdf", ResourceType: "attachment", URL: "/uploads/two.pdf", CreatedBy: "admin"},
	} {
		if err := resourceRepo.Create(context.Background(), resource); err != nil {
			t.Fatalf("create resource: %v", err)
		}
	}
	for _, material := range []*domain.CourseMaterial{
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: course.ID, ResourceID: "resource-1", DisplayName: "第二份.pdf", SortOrder: 2, CreatedBy: "admin"},
		{BaseModel: domain.BaseModel{TenantID: "tenant-1"}, CourseID: course.ID, ResourceID: "resource-2", DisplayName: "第一份.pdf", SortOrder: 1, CreatedBy: "admin"},
	} {
		if err := materialRepo.Create(ctx, material); err != nil {
			t.Fatalf("create material: %v", err)
		}
	}
	detail, err := service.GetDetail(ctx, course.ID)
	if err != nil || len(detail.Materials) != 2 || detail.Materials[0].DisplayName != "第一份.pdf" {
		t.Fatalf("GetDetail() = %#v, %v", detail, err)
	}
	emptyDetail, err := service.GetDetail(ctx, empty.ID)
	if err != nil || emptyDetail.Materials == nil || len(emptyDetail.Materials) != 0 {
		t.Fatalf("GetDetail(empty) = %#v, %v", emptyDetail, err)
	}
}

func TestCourseMaterialServiceOpenForLearnerStreamsAuthorizedFile(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	local, err := storage.NewLocal(storage.LocalConfig{Root: t.TempDir(), URL: "/uploads"})
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	courseRepo := repository.NewCourseRepository(database)
	materialRepo := repository.NewCourseMaterialRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	resourceService := NewResourceService(resourceRepo, local)
	materialService := NewCourseMaterialService(courseRepo, materialRepo, resourceRepo, resourceService)
	admin := courseContext("admin", "tenant-1", "tenant_admin")
	learner := courseContext("learner", "tenant-1", "learner")
	foreign := courseContext("foreign", "tenant-2", "learner")
	course := &domain.Course{BaseModel: domain.BaseModel{ID: "course", TenantID: "tenant-1"}, Title: "Course", Status: 1, CreatedBy: "admin"}
	if err := database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	content := []byte("%PDF-1.7\ncourse guide")
	resource, err := resourceService.UploadAttachment(admin, "guide.pdf", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("UploadAttachment() error = %v", err)
	}
	material, err := materialService.Add(admin, course.ID, CourseMaterialInput{ResourceID: resource.ID, DisplayName: "../\"guide\r\n.pdf"})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	body, contentType, fileName, err := materialService.OpenForLearner(learner, material.ID)
	if err != nil {
		t.Fatalf("OpenForLearner() error = %v", err)
	}
	got, readErr := io.ReadAll(body)
	_ = body.Close()
	if readErr != nil || !bytes.Equal(got, content) || contentType != "application/pdf" || fileName != material.DisplayName {
		t.Fatalf("OpenForLearner() = %q, %q, %q, %v", got, contentType, fileName, readErr)
	}
	if _, _, _, err := materialService.OpenForLearner(foreign, material.ID); errorCode(err) != 40400 {
		t.Fatalf("OpenForLearner(foreign) error = %#v", err)
	}
	course.Status = 0
	if err := database.Model(course).Update("status", 0).Error; err != nil {
		t.Fatalf("unpublish course: %v", err)
	}
	if _, _, _, err := materialService.OpenForLearner(learner, material.ID); errorCode(err) != 40400 {
		t.Fatalf("OpenForLearner(unpublished) error = %#v", err)
	}
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
	if _, err := fixture.courses.Get(otherAuthor, draft.ID); errorCode(err) != 40300 {
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
	if err := fixture.database.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: "tenant-1"},
		CourseID:  published.ID, UserID: "learner", Status: 1,
		AssignmentType: domain.AssignmentRequired,
	}).Error; err != nil {
		t.Fatalf("assign published course: %v", err)
	}
	items, total, err = fixture.courses.ListPublished(learner, 0, 20)
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != published.ID {
		t.Fatalf("ListPublished() = %#v, %d, %v", items, total, err)
	}
	if _, err := fixture.courses.GetPublishedDetail(learner, draft.ID); errorCode(err) != 40400 {
		t.Fatalf("GetPublishedDetail(draft) error = %#v", err)
	}
}

func TestCourseManagerInstructorForeignDirectIDPathsAreForbidden(t *testing.T) {
	fixture := newCourseFixture(t)
	owner := courseContext("owner", "tenant-1", "instructor")
	other := courseContext("other", "tenant-1", "instructor")
	foreignCourse, err := fixture.courses.Create(owner, "Foreign course", "", "")
	if err != nil {
		t.Fatalf("Create(course) error = %v", err)
	}
	foreignChapter, err := fixture.chapters.Create(owner, foreignCourse.ID, "Foreign chapter", 1)
	if err != nil {
		t.Fatalf("Create(chapter) error = %v", err)
	}
	foreignLesson, err := fixture.lessons.Create(
		owner, foreignChapter.ID, "Foreign lesson", "text", "body", 0, 1,
	)
	if err != nil {
		t.Fatalf("Create(lesson) error = %v", err)
	}

	tests := []struct {
		name string
		call func() error
	}{
		{"course get", func() error { _, err := fixture.courses.Get(other, foreignCourse.ID); return err }},
		{"course detail", func() error { _, err := fixture.courses.GetDetail(other, foreignCourse.ID); return err }},
		{"course update", func() error {
			_, err := fixture.courses.Update(other, foreignCourse.ID, "changed", "", "", 0)
			return err
		}},
		{"course delete", func() error { return fixture.courses.Delete(other, foreignCourse.ID) }},
		{"chapter list", func() error { _, err := fixture.chapters.List(other, foreignCourse.ID); return err }},
		{"chapter create", func() error { _, err := fixture.chapters.Create(other, foreignCourse.ID, "changed", 2); return err }},
		{"chapter update", func() error { _, err := fixture.chapters.Update(other, foreignChapter.ID, "changed", 2); return err }},
		{"chapter delete", func() error { return fixture.chapters.Delete(other, foreignChapter.ID) }},
		{"lesson list", func() error { _, err := fixture.lessons.List(other, foreignChapter.ID); return err }},
		{"lesson create", func() error {
			_, err := fixture.lessons.Create(other, foreignChapter.ID, "changed", "text", "body", 0, 2)
			return err
		}},
		{"lesson update", func() error {
			_, err := fixture.lessons.Update(other, foreignLesson.ID, "changed", "text", "body", 0, 2)
			return err
		}},
		{"lesson delete", func() error { return fixture.lessons.Delete(other, foreignLesson.ID) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); errorCode(err) != 40300 {
				t.Fatalf("error = %#v, want 40300", err)
			}
		})
	}
}

func TestCourseServiceRejectsTenantCreateForSuperadmin(t *testing.T) {
	fixture := newCourseFixture(t)
	if _, err := fixture.courses.Create(
		courseContext("root", "", "superadmin"), "Not official", "", "",
	); errorCode(err) != 40300 {
		t.Fatalf("Create(superadmin tenant course) error = %#v", err)
	}
}

func TestCourseManagerPolicyMatrix(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	courseService := NewCourseService(courseRepo, chapterRepo, lessonRepo, nil)
	tenantCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "tenant-course", TenantID: "tenant-1"}, Title: "Tenant", CreatedBy: "admin"}
	instructorCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "instructor-course", TenantID: "tenant-1"}, Title: "Instructor", CreatedBy: "instructor"}
	officialCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "official-course"}, Title: "Official", Status: 1, CreatedBy: "root", IsOfficial: true}
	nonOfficialPlatformCourse := &domain.Course{BaseModel: domain.BaseModel{ID: "invalid-platform-course"}, Title: "Invalid", CreatedBy: "root"}
	for _, course := range []*domain.Course{tenantCourse, instructorCourse, officialCourse, nonOfficialPlatformCourse} {
		if err := database.Create(course).Error; err != nil {
			t.Fatalf("create course: %v", err)
		}
	}
	if err := courseRepo.ActivateOfficial(context.Background(), "tenant-1", officialCourse.ID, true); err != nil {
		t.Fatalf("activate official course: %v", err)
	}
	tests := []struct {
		name string
		ctx  context.Context
		id   string
		code int
	}{
		{"tenant admin tenant course", courseContext("admin", "tenant-1", "tenant_admin"), tenantCourse.ID, 0},
		{"tenant admin instructor course", courseContext("admin", "tenant-1", "tenant_admin"), instructorCourse.ID, 0},
		{"tenant admin official course", courseContext("admin", "tenant-1", "tenant_admin"), officialCourse.ID, 40300},
		{"instructor owned course", courseContext("instructor", "tenant-1", "instructor"), instructorCourse.ID, 0},
		{"instructor foreign course", courseContext("instructor", "tenant-1", "instructor"), tenantCourse.ID, 40300},
		{"superadmin official course", courseContext("root", "", "superadmin"), officialCourse.ID, 0},
		{"superadmin non-official empty-scope course", courseContext("root", "", "superadmin"), nonOfficialPlatformCourse.ID, 40300},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := courseService.Get(test.ctx, test.id)
			if got := errorCode(err); got != test.code {
				t.Fatalf("Get() code = %d, error = %#v, want %d", got, err, test.code)
			}
		})
	}
}

func newCourseFixture(t *testing.T) courseFixture {
	t.Helper()
	database, _, _ := serviceRepositories(t)
	courseRepo := repository.NewCourseRepository(database)
	chapterRepo := repository.NewCourseChapterRepository(database)
	lessonRepo := repository.NewCourseLessonRepository(database)
	resourceRepo := repository.NewResourceRepository(database)
	materialRepo := repository.NewCourseMaterialRepository(database)
	enrollmentRepo := repository.NewCourseEnrollmentRepository(database)
	courses := NewCourseService(courseRepo, chapterRepo, lessonRepo, enrollmentRepo, materialRepo)
	return courseFixture{
		database: database,
		courses:  courses,
		chapters: NewCourseChapterService(chapterRepo, courseRepo),
		lessons: NewCourseLessonService(
			lessonRepo, chapterRepo, courseRepo, resourceRepo,
		),
		enrollments: enrollmentRepo,
		resources:   resourceRepo,
	}
}

func courseContext(userID, tenantID, role string) context.Context {
	return usercontext.WithUser(context.Background(), userID, tenantID, "", role)
}
