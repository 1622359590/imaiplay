package service

import (
	"context"
	"errors"
	"testing"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
)

func TestProgressServiceRequiresActiveEnrollment(t *testing.T) {
	fixture := newLearningFixture(t)
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")

	_, err := fixture.progress.Report(learner, fixture.lesson.ID, 10, 10)
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40300 ||
		appErr.Message != "not enrolled in this course" {
		t.Fatalf("Report(not enrolled) error = %#v", err)
	}
	inactive := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: fixture.tenant.ID},
		CourseID:  fixture.course.ID, UserID: fixture.learner.ID, Status: 0,
	}
	if err := fixture.enrollmentRepo.Create(context.Background(), inactive); err != nil {
		t.Fatalf("create inactive enrollment: %v", err)
	}
	if err := fixture.database.Model(&domain.CourseEnrollment{}).
		Where("id = ?", inactive.ID).Update("status", 0).Error; err != nil {
		t.Fatalf("disable enrollment: %v", err)
	}
	if _, err := fixture.progress.Report(
		learner, fixture.lesson.ID, 10, 10,
	); errorCode(err) != 40300 {
		t.Fatalf("Report(inactive enrollment) error = %#v", err)
	}
}

func TestProgressServiceReportGetAndRecent(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID,
	); err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}
	progress, err := fixture.progress.Report(learner, fixture.lesson.ID, 30, 30)
	if err != nil || progress.Status != 1 || progress.CompletedAt != nil {
		t.Fatalf("Report(in progress) = %#v, %v", progress, err)
	}
	progress, err = fixture.progress.Report(learner, fixture.lesson.ID, 100, 100)
	if err != nil || progress.Status != 2 || progress.CompletedAt == nil {
		t.Fatalf("Report(completed) = %#v, %v", progress, err)
	}
	got, err := fixture.progress.Get(learner, fixture.lesson.ID)
	if err != nil || got.ID != progress.ID || got.UserID != fixture.learner.ID {
		t.Fatalf("Get() = %#v, %v", got, err)
	}
	items, total, err := fixture.progress.GetRecent(learner, 0, 20)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("GetRecent() = %#v, %d, %v", items, total, err)
	}
	if items[0].Course.ID != fixture.course.ID ||
		items[0].Lesson.ID != fixture.lesson.ID ||
		items[0].Progress.ID != progress.ID ||
		items[0].LastLearnedAt.IsZero() {
		t.Fatalf("GetRecent()[0] = %#v", items[0])
	}
}

func TestProgressServiceValidatesLearnerAndValues(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID,
	); err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}
	for _, input := range []struct {
		position int
		percent  int
	}{
		{position: -1, percent: 10},
		{position: 0, percent: -1},
		{position: 0, percent: 101},
	} {
		if _, err := fixture.progress.Report(
			learner, fixture.lesson.ID, input.position, input.percent,
		); errorCode(err) != 40000 {
			t.Fatalf("Report(%#v) error = %#v", input, err)
		}
	}
	if _, _, err := fixture.progress.GetRecent(
		admin, 0, 20,
	); errorCode(err) != 40300 {
		t.Fatalf("GetRecent(admin) error = %#v", err)
	}
}
