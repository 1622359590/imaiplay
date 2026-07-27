package service

import "testing"

func TestCourseChapterServiceCRUDAndInstructorOwnership(t *testing.T) {
	fixture := newCourseFixture(t)
	author := courseContext("author-1", "tenant-1", "instructor")
	other := courseContext("author-2", "tenant-1", "instructor")
	course, err := fixture.courses.Create(author, "Course", "", "")
	if err != nil {
		t.Fatalf("Create(course) error = %v", err)
	}
	chapter, err := fixture.chapters.Create(author, course.ID, "Chapter", 2)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	items, err := fixture.chapters.List(author, course.ID)
	if err != nil || len(items) != 1 || items[0].ID != chapter.ID {
		t.Fatalf("List() = %#v, %v", items, err)
	}
	if _, err := fixture.chapters.List(other, course.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor List() error = %#v", err)
	}
	updated, err := fixture.chapters.Update(author, chapter.ID, "Updated", 1)
	if err != nil || updated.Title != "Updated" || updated.SortOrder != 1 {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}
	if err := fixture.chapters.Delete(other, chapter.ID); errorCode(err) != 40400 {
		t.Fatalf("other instructor Delete() error = %#v", err)
	}
	if err := fixture.chapters.Delete(author, chapter.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
