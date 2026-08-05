package service

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

func TestLearnerAccessAuthorizationMatrix(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	courses := repository.NewCourseRepository(database)
	enrollments := repository.NewCourseEnrollmentRepository(database)
	materials := repository.NewCourseMaterialRepository(database)
	access := NewLearnerAccess(courses, enrollments, materials)

	create := func(value any) {
		t.Helper()
		if err := database.Create(value).Error; err != nil {
			t.Fatalf("create %T: %v", value, err)
		}
	}
	course := func(id, tenantID string, status int, official bool) *domain.Course {
		value := &domain.Course{
			BaseModel: domain.BaseModel{ID: id, TenantID: tenantID},
			Title:     id, Status: status, IsOfficial: official, CreatedBy: "manager",
		}
		create(value)
		return value
	}
	enroll := func(courseID string, status int) {
		enrollment := &domain.CourseEnrollment{
			BaseModel: domain.BaseModel{TenantID: "tenant-1"},
			CourseID:  courseID, UserID: "learner-1", Status: status,
			AssignmentType: domain.AssignmentRequired,
		}
		create(enrollment)
		if status == 0 {
			if err := database.Model(enrollment).Update("status", 0).Error; err != nil {
				t.Fatalf("disable enrollment: %v", err)
			}
		}
	}
	lessonResource := func(prefix string, item *domain.Course, resourceTenant string) *domain.Resource {
		resource := &domain.Resource{
			BaseModel: domain.BaseModel{ID: prefix + "-resource", TenantID: resourceTenant},
			Name:      prefix + ".mp4", ResourceType: "video",
			URL: "/uploads/" + prefix + ".mp4", CreatedBy: "manager",
		}
		chapter := &domain.CourseChapter{
			BaseModel: domain.BaseModel{ID: prefix + "-chapter", TenantID: item.TenantID},
			CourseID:  item.ID, Title: prefix,
		}
		create(resource)
		create(chapter)
		create(&domain.CourseLesson{
			BaseModel: domain.BaseModel{ID: prefix + "-lesson", TenantID: item.TenantID},
			ChapterID: chapter.ID, Title: prefix, ContentType: "video", ResourceID: &resource.ID,
		})
		return resource
	}
	material := func(id string, item *domain.Course, resourceTenant string) *domain.CourseMaterial {
		resource := &domain.Resource{
			BaseModel: domain.BaseModel{ID: id + "-attachment", TenantID: resourceTenant},
			Name:      id + ".pdf", ResourceType: "attachment",
			URL: "/uploads/" + id + ".pdf", CreatedBy: "manager",
		}
		create(resource)
		value := &domain.CourseMaterial{
			BaseModel: domain.BaseModel{ID: id, TenantID: item.TenantID},
			CourseID:  item.ID, ResourceID: resource.ID, DisplayName: resource.Name,
			CreatedBy: "manager",
		}
		create(value)
		return value
	}

	assigned := course("assigned", "tenant-1", 1, false)
	assignedResource := lessonResource("assigned", assigned, "tenant-1")
	assignedMaterial := material("assigned-material", assigned, "tenant-1")
	enroll(assigned.ID, 1)

	inactive := course("inactive", "tenant-1", 1, false)
	inactiveResource := lessonResource("inactive", inactive, "tenant-1")
	enroll(inactive.ID, 0)

	draft := course("draft", "tenant-1", 0, false)
	draftResource := lessonResource("draft", draft, "tenant-1")
	enroll(draft.ID, 1)

	otherCourse := course("other-course", "tenant-1", 1, false)
	otherMaterial := material("other-material", otherCourse, "tenant-1")
	unrelated := &domain.Resource{
		BaseModel: domain.BaseModel{ID: "unrelated", TenantID: "tenant-1"},
		Name:      "unrelated.mp4", ResourceType: "video", URL: "/uploads/unrelated.mp4", CreatedBy: "manager",
	}
	create(unrelated)

	foreign := course("foreign", "tenant-2", 1, false)
	foreignResource := lessonResource("foreign", foreign, "tenant-2")

	disabledOfficial := course("disabled-official", "", 1, true)
	disabledOfficialResource := lessonResource("disabled-official", disabledOfficial, "")
	enroll(disabledOfficial.ID, 1)
	create(&domain.TenantOfficialCourse{TenantID: "tenant-1", CourseID: disabledOfficial.ID, Enabled: false})

	unactivatedOfficial := course("unactivated-official", "", 1, true)
	unactivatedOfficialResource := lessonResource("unactivated-official", unactivatedOfficial, "")
	enroll(unactivatedOfficial.ID, 1)
	enabledOfficial := course("enabled-official", "", 1, true)
	enabledOfficialResource := lessonResource("enabled-official", enabledOfficial, "")
	enroll(enabledOfficial.ID, 1)
	create(&domain.TenantOfficialCourse{TenantID: "tenant-1", CourseID: enabledOfficial.ID, Enabled: true})

	deletedResourceMaterial := material("deleted-resource-material", assigned, "tenant-1")
	if err := database.Where("id = ?", deletedResourceMaterial.ResourceID).Delete(&domain.Resource{}).Error; err != nil {
		t.Fatalf("delete material resource: %v", err)
	}
	deletedChapterResource := lessonResource("deleted-chapter", assigned, "tenant-1")
	if err := database.Where("id = ?", "deleted-chapter-chapter").Delete(&domain.CourseChapter{}).Error; err != nil {
		t.Fatalf("delete lesson chapter: %v", err)
	}

	learner := courseContext("learner-1", "tenant-1", "learner")
	notEnrolled := courseContext("learner-2", "tenant-1", "learner")
	tests := []struct {
		name string
		call func() error
		want int
	}{
		{"active course", func() error { _, err := access.AuthorizeCourse(learner, assigned.ID); return err }, 0},
		{"active material", func() error { _, _, err := access.AuthorizeMaterial(learner, assignedMaterial.ID); return err }, 0},
		{"active lesson resource", func() error { _, err := access.AuthorizeLessonResource(learner, assignedResource.ID); return err }, 0},
		{"not enrolled", func() error { _, err := access.AuthorizeCourse(notEnrolled, assigned.ID); return err }, 40400},
		{"inactive enrollment", func() error { _, err := access.AuthorizeLessonResource(learner, inactiveResource.ID); return err }, 40400},
		{"draft course", func() error { _, err := access.AuthorizeLessonResource(learner, draftResource.ID); return err }, 40400},
		{"other-course material", func() error { _, _, err := access.AuthorizeMaterial(learner, otherMaterial.ID); return err }, 40400},
		{"unrelated same-tenant resource", func() error { _, err := access.AuthorizeLessonResource(learner, unrelated.ID); return err }, 40400},
		{"cross-tenant resource", func() error { _, err := access.AuthorizeLessonResource(learner, foreignResource.ID); return err }, 40400},
		{"disabled official", func() error {
			_, err := access.AuthorizeLessonResource(learner, disabledOfficialResource.ID)
			return err
		}, 40400},
		{"unactivated official", func() error {
			_, err := access.AuthorizeLessonResource(learner, unactivatedOfficialResource.ID)
			return err
		}, 40400},
		{"enabled official", func() error {
			_, err := access.AuthorizeLessonResource(learner, enabledOfficialResource.ID)
			return err
		}, 0},
		{"missing resource relation", func() error { _, err := access.AuthorizeLessonResource(learner, "missing"); return err }, 40400},
		{"deleted material resource", func() error {
			_, _, err := access.AuthorizeMaterial(learner, deletedResourceMaterial.ID)
			return err
		}, 40400},
		{"deleted lesson chapter", func() error {
			_, err := access.AuthorizeLessonResource(learner, deletedChapterResource.ID)
			return err
		}, 40400},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := errorCode(test.call()); got != test.want {
				t.Fatalf("error code = %d, want %d", got, test.want)
			}
		})
	}
}

func TestLearnerAccessPreservesDatabaseFailures(t *testing.T) {
	database, _, _ := serviceRepositories(t)
	access := NewLearnerAccess(
		repository.NewCourseRepository(database),
		repository.NewCourseEnrollmentRepository(database),
		repository.NewCourseMaterialRepository(database),
	)
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("database DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	if _, err := access.AuthorizeCourse(
		courseContext("learner-1", "tenant-1", "learner"), "course-1",
	); errorCode(err) != 50000 {
		t.Fatalf("AuthorizeCourse(database failure) error = %#v", err)
	}
}
