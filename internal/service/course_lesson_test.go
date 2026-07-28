package service

import "testing"

func TestCourseLessonServiceCRUDValidationAndInstructorOwnership(t *testing.T) {
	fixture := newCourseFixture(t)
	author := courseContext("author-1", "tenant-1", "instructor")
	other := courseContext("author-2", "tenant-1", "instructor")
	course, err := fixture.courses.Create(author, "Course", "", "")
	if err != nil {
		t.Fatalf("Create(course) error = %v", err)
	}
	chapter, err := fixture.chapters.Create(author, course.ID, "Chapter", 1)
	if err != nil {
		t.Fatalf("Create(chapter) error = %v", err)
	}
	if _, err := fixture.lessons.Create(
		author, chapter.ID, "Invalid", "archive", "", 0, 1,
	); errorCode(err) != 40000 {
		t.Fatalf("Create(invalid type) error = %#v", err)
	}
	lesson, err := fixture.lessons.Create(
		author, chapter.ID, "Lesson", "video", "/video.mp4", 90, 1,
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	resourceID := "resource-1"
	resourceLesson, err := fixture.lessons.CreateWithResource(
		author, chapter.ID, "Resource lesson", "video", resourceID, "", 120, 2,
	)
	if err != nil || resourceLesson.ResourceID == nil || *resourceLesson.ResourceID != resourceID {
		t.Fatalf("CreateWithResource() = %#v, %v", resourceLesson, err)
	}
	items, err := fixture.lessons.List(author, chapter.ID)
	if err != nil || len(items) != 2 || items[0].ID != lesson.ID {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	if _, err := fixture.lessons.List(other, chapter.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor List() error = %#v", err)
	}
	updated, err := fixture.lessons.Update(
		author, lesson.ID, "Updated", "text", "Lesson body", 0, 2,
	)
	if err != nil || updated.Title != "Updated" || updated.ContentType != "text" {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if err := fixture.lessons.Delete(other, lesson.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor Delete() error = %#v", err)
	}
	if err := fixture.lessons.Delete(author, lesson.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
