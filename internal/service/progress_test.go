package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

func TestProgressServiceRecordsIdempotentLearningTimeAcrossShanghaiMidnight(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID, domain.AssignmentRequired,
	); err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}

	fixture.progress.now = func() time.Time {
		return time.Date(2026, 8, 5, 15, 59, 0, 0, time.UTC)
	}
	if _, err := fixture.progress.Report(learner, fixture.lesson.ID, 10, 10, 0, ""); err != nil {
		t.Fatalf("legacy Report() error = %v", err)
	}
	if _, err := fixture.progress.Report(learner, fixture.lesson.ID, 20, 20, 15, "report-15"); err != nil {
		t.Fatalf("Report(15) error = %v", err)
	}
	if _, err := fixture.progress.Report(learner, fixture.lesson.ID, 21, 21, 1, "report-1"); err != nil {
		t.Fatalf("Report(1) error = %v", err)
	}
	if _, err := fixture.progress.Report(learner, fixture.lesson.ID, 22, 22, 15, "report-15"); err != nil {
		t.Fatalf("Report(duplicate) error = %v", err)
	}
	fixture.progress.now = func() time.Time {
		return time.Date(2026, 8, 5, 16, 1, 0, 0, time.UTC)
	}
	if _, err := fixture.progress.Report(learner, fixture.lesson.ID, 30, 30, 60, "report-60"); err != nil {
		t.Fatalf("Report(60) error = %v", err)
	}

	for _, want := range []struct {
		date    string
		seconds int64
	}{{"2026-08-05", 16}, {"2026-08-06", 60}} {
		var stat domain.LearningDailyStat
		err := fixture.database.Where(
			"tenant_id = ? AND user_id = ? AND study_date = ?",
			fixture.tenant.ID, fixture.learner.ID, want.date,
		).First(&stat).Error
		if err != nil || stat.DurationSeconds != want.seconds {
			t.Fatalf("daily stat %s = %#v, %v", want.date, stat, err)
		}
	}
	var reports int64
	if err := fixture.database.Model(&domain.LearningTimeReport{}).Count(&reports).Error; err != nil || reports != 3 {
		t.Fatalf("report count = %d, %v", reports, err)
	}

	for _, input := range []struct {
		delta    int
		reportID string
	}{
		{delta: 0, reportID: "zero-is-not-legacy"},
		{delta: -1, reportID: "negative"},
		{delta: 61, reportID: "large"},
		{delta: 1, reportID: ""},
	} {
		if _, err := fixture.progress.Report(
			learner, fixture.lesson.ID, 40, 40, input.delta, input.reportID,
		); errorCode(err) != 40000 {
			t.Fatalf("Report(%#v) error = %#v", input, err)
		}
	}
}

func TestProgressServiceUsesOneTimestampForCompletionAndStudyDate(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID, domain.AssignmentRequired,
	); err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}
	beforeMidnight := time.Date(2026, 8, 5, 15, 59, 59, 0, time.UTC)
	afterMidnight := beforeMidnight.Add(2 * time.Second)
	calls := 0
	fixture.progress.now = func() time.Time {
		calls++
		if calls == 1 {
			return beforeMidnight
		}
		return afterMidnight
	}

	progress, err := fixture.progress.Report(
		learner, fixture.lesson.ID, 100, 100, 15, "single-clock",
	)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if calls != 1 || progress.CompletedAt == nil || !progress.CompletedAt.Equal(beforeMidnight) {
		t.Fatalf("clock calls=%d completed_at=%v", calls, progress.CompletedAt)
	}
	var stat domain.LearningDailyStat
	if err := fixture.database.Where("study_date = ?", "2026-08-05").First(&stat).Error; err != nil || stat.DurationSeconds != 15 {
		t.Fatalf("daily stat = %#v, %v", stat, err)
	}
}

func TestProgressServiceRequiresActiveEnrollment(t *testing.T) {
	fixture := newLearningFixture(t)
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")

	_, err := fixture.progress.Report(learner, fixture.lesson.ID, 10, 10, 0, "")
	var appErr *errorsx.AppError
	if !errors.As(err, &appErr) || appErr.Code != 40300 ||
		appErr.Message != "not enrolled in this course" {
		t.Fatalf("Report(not enrolled) error = %#v", err)
	}
	inactive := &domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: fixture.tenant.ID},
		CourseID:  fixture.course.ID, UserID: fixture.learner.ID, Status: 1,
		AssignmentType: domain.AssignmentRequired,
	}
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	if err := fixture.enrollmentRepo.Create(admin, inactive); err != nil {
		t.Fatalf("create inactive enrollment: %v", err)
	}
	if err := fixture.database.Model(&domain.CourseEnrollment{}).
		Where("id = ?", inactive.ID).Update("status", 0).Error; err != nil {
		t.Fatalf("disable enrollment: %v", err)
	}
	if _, err := fixture.progress.Report(
		learner, fixture.lesson.ID, 10, 10, 0, "",
	); errorCode(err) != 40300 {
		t.Fatalf("Report(inactive enrollment) error = %#v", err)
	}
}

func TestProgressServiceReportAndGet(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID, domain.AssignmentRequired,
	); err != nil {
		t.Fatalf("Enroll() error = %v", err)
	}
	progress, err := fixture.progress.Report(learner, fixture.lesson.ID, 30, 30, 0, "")
	if err != nil || progress.Status != 1 || progress.CompletedAt != nil {
		t.Fatalf("Report(in progress) = %#v, %v", progress, err)
	}
	progress, err = fixture.progress.Report(learner, fixture.lesson.ID, 100, 100, 0, "")
	if err != nil || progress.Status != 2 || progress.CompletedAt == nil {
		t.Fatalf("Report(completed) = %#v, %v", progress, err)
	}
	got, err := fixture.progress.Get(learner, fixture.lesson.ID)
	if err != nil || got.ID != progress.ID || got.UserID != fixture.learner.ID {
		t.Fatalf("Get() = %#v, %v", got, err)
	}
}

func TestProgressServiceGetStartsPublishedCourseAtZero(t *testing.T) {
	fixture := newLearningFixture(t)
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")

	progress, err := fixture.progress.Get(learner, fixture.lesson.ID)
	if err != nil || progress.UserID != fixture.learner.ID ||
		progress.LessonID != fixture.lesson.ID || progress.ProgressPercent != 0 {
		t.Fatalf("Get(new learner) = %#v, %v", progress, err)
	}
	enrollment, err := fixture.enrollmentRepo.FindByCourseAndUser(
		learner, fixture.course.ID, fixture.learner.ID,
	)
	if err != nil || enrollment.Status != 1 || enrollment.AssignmentType != domain.AssignmentRequired {
		t.Fatalf("auto enrollment = %#v, %v", enrollment, err)
	}
}

func TestProgressServiceGetStartsEnabledOfficialCourseAsOptional(t *testing.T) {
	fixture := newLearningFixture(t)
	official := &domain.Course{
		Title: "Official", Status: 1, CreatedBy: "root", IsOfficial: true,
		CourseType: domain.CourseTypeOptional,
	}
	if err := fixture.database.Create(official).Error; err != nil {
		t.Fatalf("create official course: %v", err)
	}
	chapter := &domain.CourseChapter{CourseID: official.ID, Title: "Official chapter"}
	if err := fixture.database.Create(chapter).Error; err != nil {
		t.Fatalf("create official chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		ChapterID: chapter.ID, Title: "Official lesson", ContentType: "video",
	}
	if err := fixture.database.Create(lesson).Error; err != nil {
		t.Fatalf("create official lesson: %v", err)
	}
	if err := fixture.database.Create(&domain.TenantOfficialCourse{
		TenantID: fixture.tenant.ID, CourseID: official.ID, Enabled: true,
	}).Error; err != nil {
		t.Fatalf("enable official course: %v", err)
	}
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")

	progress, err := fixture.progress.Get(learner, lesson.ID)
	if err != nil || progress.LessonID != lesson.ID || progress.ProgressPercent != 0 {
		t.Fatalf("Get(official lesson) = %#v, %v", progress, err)
	}
	enrollment, err := fixture.enrollmentRepo.FindByCourseAndUser(
		learner, official.ID, fixture.learner.ID,
	)
	if err != nil || enrollment.Status != 1 || enrollment.AssignmentType != domain.AssignmentOptional {
		t.Fatalf("official auto enrollment = %#v, %v", enrollment, err)
	}

	requiredCourse, requiredLesson := seedProgressCourse(
		t, fixture, "Required official", 1, true, boolPointer(true),
	)
	if err := fixture.database.Create(&domain.CourseEnrollment{
		BaseModel: domain.BaseModel{TenantID: fixture.tenant.ID},
		CourseID:  requiredCourse.ID, UserID: fixture.learner.ID, Status: 1,
		AssignmentType: domain.AssignmentRequired,
	}).Error; err != nil {
		t.Fatalf("create required official enrollment: %v", err)
	}
	if _, err := fixture.progress.Get(learner, requiredLesson.ID); err != nil {
		t.Fatalf("Get(required official lesson) error = %v", err)
	}
	requiredEnrollment, err := fixture.enrollmentRepo.FindByCourseAndUser(
		learner, requiredCourse.ID, fixture.learner.ID,
	)
	if err != nil || requiredEnrollment.AssignmentType != domain.AssignmentRequired {
		t.Fatalf("required official enrollment = %#v, %v", requiredEnrollment, err)
	}
}

func TestProgressServiceReportStartsEnabledOfficialCourseAsOptional(t *testing.T) {
	fixture := newLearningFixture(t)
	official, lesson := seedProgressCourse(
		t, fixture, "Reported official", 1, true, boolPointer(true),
	)
	if err := fixture.database.Model(official).
		Update("course_type", domain.CourseTypeOptional).Error; err != nil {
		t.Fatalf("set reported official optional course type: %v", err)
	}
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")

	progress, err := fixture.progress.Report(learner, lesson.ID, 10, 10, 0, "")
	if err != nil || progress.ProgressPercent != 10 {
		t.Fatalf("Report(enabled official) = %#v, %v", progress, err)
	}
	enrollment, err := fixture.enrollmentRepo.FindByCourseAndUser(
		learner, official.ID, fixture.learner.ID,
	)
	if err != nil || enrollment.Status != 1 || enrollment.AssignmentType != domain.AssignmentOptional {
		t.Fatalf("reported official enrollment = %#v, %v", enrollment, err)
	}
}

func boolPointer(value bool) *bool { return &value }

func TestProgressServiceHidesInaccessibleCourseStatesBeforeEnrollment(t *testing.T) {
	fixture := newLearningFixture(t)
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	disabled := false
	enabled := true
	tests := []struct {
		name      string
		status    int
		official  bool
		enabled   *bool
		preassign bool
	}{
		{name: "draft tenant course", status: 0},
		{name: "draft tenant course with preassigned enrollment", status: 0, preassign: true},
		{name: "unpublished official course", status: 0, official: true, enabled: &enabled},
		{name: "disabled official course", status: 1, official: true, enabled: &disabled},
		{name: "unactivated official course", status: 1, official: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			course, lesson := seedProgressCourse(t, fixture, test.name, test.status, test.official, test.enabled)
			if test.preassign {
				if err := fixture.enrollmentRepo.Create(admin, &domain.CourseEnrollment{
					BaseModel: domain.BaseModel{TenantID: fixture.tenant.ID},
					CourseID:  course.ID, UserID: fixture.learner.ID, Status: 1,
					AssignmentType: domain.AssignmentRequired,
				}); err != nil {
					t.Fatalf("preassign enrollment: %v", err)
				}
			}
			before := enrollmentCount(t, fixture, course.ID)
			if _, err := fixture.progress.Get(learner, lesson.ID); errorCode(err) != 40400 {
				t.Errorf("Get() error = %#v, want 40400", err)
			}
			if _, err := fixture.progress.Report(learner, lesson.ID, 10, 10, 0, ""); errorCode(err) != 40400 {
				t.Errorf("Report() error = %#v, want 40400", err)
			}
			if after := enrollmentCount(t, fixture, course.ID); after != before {
				t.Errorf("enrollment count = %d, want unchanged %d", after, before)
			}
			var progressCount int64
			if err := fixture.database.Model(&domain.LessonProgress{}).
				Where("lesson_id = ? AND user_id = ?", lesson.ID, fixture.learner.ID).
				Count(&progressCount).Error; err != nil || progressCount != 0 {
				t.Errorf("progress count = %d, error = %v", progressCount, err)
			}
		})
	}
}

func TestProgressServiceMapsEnrollmentCreateNotFoundToNotFound(t *testing.T) {
	fixture := newLearningFixture(t)
	enrollments := &createInterceptEnrollmentRepository{
		CourseEnrollmentRepository: fixture.enrollmentRepo,
		create: func(context.Context, *domain.CourseEnrollment) error {
			return gorm.ErrRecordNotFound
		},
	}
	progress := progressServiceWithEnrollments(fixture, enrollments)
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := progress.Get(learner, fixture.lesson.ID); errorCode(err) != 40400 {
		t.Fatalf("Get() error = %#v, want 40400", err)
	}
	if count := enrollmentCount(t, fixture, fixture.course.ID); count != 0 {
		t.Fatalf("enrollment count = %d, want 0", count)
	}
}

func TestProgressServiceTreatsConcurrentEnrollmentCreateAsIdempotent(t *testing.T) {
	fixture := newLearningFixture(t)
	enrollments := &createInterceptEnrollmentRepository{
		CourseEnrollmentRepository: fixture.enrollmentRepo,
	}
	enrollments.create = func(ctx context.Context, enrollment *domain.CourseEnrollment) error {
		concurrent := *enrollment
		if err := fixture.enrollmentRepo.Create(ctx, &concurrent); err != nil {
			return err
		}
		return gorm.ErrDuplicatedKey
	}
	progress := progressServiceWithEnrollments(fixture, enrollments)
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	item, err := progress.Get(learner, fixture.lesson.ID)
	if err != nil || item.LessonID != fixture.lesson.ID || item.UserID != fixture.learner.ID {
		t.Fatalf("Get() = %#v, %v", item, err)
	}
	if count := enrollmentCount(t, fixture, fixture.course.ID); count != 1 {
		t.Fatalf("enrollment count = %d, want 1", count)
	}
}

type createInterceptEnrollmentRepository struct {
	repository.CourseEnrollmentRepository
	create func(context.Context, *domain.CourseEnrollment) error
}

func (repo *createInterceptEnrollmentRepository) Create(
	ctx context.Context, enrollment *domain.CourseEnrollment,
) error {
	return repo.create(ctx, enrollment)
}

func progressServiceWithEnrollments(
	fixture learningFixture, enrollments repository.CourseEnrollmentRepository,
) *ProgressService {
	return NewProgressService(
		repository.NewLessonProgressRepository(fixture.database), enrollments,
		repository.NewCourseLessonRepository(fixture.database),
		repository.NewCourseChapterRepository(fixture.database),
		repository.NewCourseRepository(fixture.database),
	)
}

func seedProgressCourse(
	t *testing.T, fixture learningFixture, title string,
	status int, official bool, enabled *bool,
) (*domain.Course, *domain.CourseLesson) {
	t.Helper()
	tenantID := fixture.tenant.ID
	createdBy := fixture.admin.ID
	if official {
		tenantID = ""
		createdBy = "root"
	}
	course := &domain.Course{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Title:     title, Status: status, CreatedBy: createdBy, IsOfficial: official,
	}
	if err := fixture.database.Create(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	if status == 0 {
		if err := fixture.database.Model(&domain.Course{}).Where("id = ?", course.ID).
			Update("status", 0).Error; err != nil {
			t.Fatalf("persist draft status: %v", err)
		}
		course.Status = 0
	}
	chapter := &domain.CourseChapter{
		BaseModel: domain.BaseModel{TenantID: tenantID}, CourseID: course.ID,
		Title: title + " chapter",
	}
	if err := fixture.database.Create(chapter).Error; err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	lesson := &domain.CourseLesson{
		BaseModel: domain.BaseModel{TenantID: tenantID}, ChapterID: chapter.ID,
		Title: title + " lesson", ContentType: "video",
	}
	if err := fixture.database.Create(lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	if enabled != nil {
		if err := fixture.database.Create(&domain.TenantOfficialCourse{
			TenantID: fixture.tenant.ID, CourseID: course.ID, Enabled: *enabled,
		}).Error; err != nil {
			t.Fatalf("set official activation: %v", err)
		}
		if !*enabled {
			if err := fixture.database.Model(&domain.TenantOfficialCourse{}).
				Where("tenant_id = ? AND course_id = ?", fixture.tenant.ID, course.ID).
				Update("enabled", false).Error; err != nil {
				t.Fatalf("disable official course: %v", err)
			}
		}
	}
	return course, lesson
}

func enrollmentCount(t *testing.T, fixture learningFixture, courseID string) int64 {
	t.Helper()
	var count int64
	if err := fixture.database.Model(&domain.CourseEnrollment{}).
		Where("tenant_id = ? AND course_id = ? AND user_id = ?", fixture.tenant.ID, courseID, fixture.learner.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count enrollments: %v", err)
	}
	return count
}

func TestProgressServiceValidatesLearnerAndValues(t *testing.T) {
	fixture := newLearningFixture(t)
	admin := courseContext(fixture.admin.ID, fixture.tenant.ID, "tenant_admin")
	learner := courseContext(fixture.learner.ID, fixture.tenant.ID, "learner")
	if _, err := fixture.enrollments.Enroll(
		admin, fixture.course.ID, fixture.learner.ID, domain.AssignmentRequired,
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
			learner, fixture.lesson.ID, input.position, input.percent, 0, "",
		); errorCode(err) != 40000 {
			t.Fatalf("Report(%#v) error = %#v", input, err)
		}
	}
}
