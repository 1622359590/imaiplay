package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type learnerMotivationGORMRepository struct {
	database *gorm.DB
}

func NewLearnerMotivationRepository(database *gorm.DB) LearnerMotivationRepository {
	return &learnerMotivationGORMRepository{database: database}
}

func (repository *learnerMotivationGORMRepository) MarkFirstLogin(
	ctx context.Context,
	userID string,
	at time.Time,
) (bool, error) {
	contextUserID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || tenantID == "" || userID == "" || contextUserID != userID {
		return false, gorm.ErrRecordNotFound
	}
	result := repository.database.WithContext(ctx).
		Model(&domain.LearnerEngagementState{}).
		Where("tenant_id = ? AND user_id = ? AND first_login_at IS NULL", tenantID, userID).
		Update("first_login_at", at.UTC())
	return result.RowsAffected == 1, result.Error
}

func (repository *learnerMotivationGORMRepository) Load(
	ctx context.Context,
	today, yesterday, dayBefore string,
	yesterdayStart, todayStart time.Time,
) (LearnerMotivationSnapshot, error) {
	userID, tenantID, err := learnerMotivationIdentity(ctx)
	if err != nil {
		return LearnerMotivationSnapshot{}, err
	}
	database := repository.database.WithContext(ctx)
	result := LearnerMotivationSnapshot{}
	if err := database.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		First(&result.State).Error; err != nil {
		return LearnerMotivationSnapshot{}, err
	}
	if result.YesterdaySeconds, err = motivationDailySeconds(database, tenantID, userID, yesterday); err != nil {
		return LearnerMotivationSnapshot{}, err
	}
	if result.DayBeforeSeconds, err = motivationDailySeconds(database, tenantID, userID, dayBefore); err != nil {
		return LearnerMotivationSnapshot{}, err
	}

	overview, err := (&learnerOverviewGORMRepository{database: repository.database}).Get(ctx, today)
	if err != nil {
		return LearnerMotivationSnapshot{}, err
	}
	visible := make([]motivationVisibleCourse, 0, len(overview.Courses))
	lessonIDs := make([]string, 0)

	var enrollments []domain.CourseEnrollment
	if err := database.Where(
		"tenant_id = ? AND user_id = ? AND status = ?", tenantID, userID, 1,
	).Find(&enrollments).Error; err != nil {
		return LearnerMotivationSnapshot{}, err
	}
	enrollmentByCourse := make(map[string]domain.CourseEnrollment, len(enrollments))
	for _, enrollment := range enrollments {
		enrollmentByCourse[enrollment.CourseID] = enrollment
	}

	for _, item := range overview.Courses {
		lessons, loadErr := motivationCourseLessons(database, item.Course)
		if loadErr != nil {
			return LearnerMotivationSnapshot{}, loadErr
		}
		progressByLesson := make(map[string]domain.LessonProgress, len(lessons))
		courseLessonIDs := make([]string, 0, len(lessons))
		for _, lesson := range lessons {
			courseLessonIDs = append(courseLessonIDs, lesson.ID)
			lessonIDs = append(lessonIDs, lesson.ID)
		}
		if len(courseLessonIDs) > 0 {
			var progressItems []domain.LessonProgress
			if err := database.Where(
				"tenant_id = ? AND user_id = ? AND lesson_id IN ?", tenantID, userID, courseLessonIDs,
			).Find(&progressItems).Error; err != nil {
				return LearnerMotivationSnapshot{}, err
			}
			for _, progress := range progressItems {
				progressByLesson[progress.LessonID] = progress
			}
		}
		var enrollment *domain.CourseEnrollment
		if value, exists := enrollmentByCourse[item.Course.ID]; exists {
			copy := value
			enrollment = &copy
		}
		visible = append(visible, motivationVisibleCourse{data: item, lessons: lessons, progress: progressByLesson, enrollment: enrollment})
		if item.AssignmentType == domain.AssignmentRequired {
			result.RequiredTotal++
			if len(lessons) > 0 && item.CompletedLessonCount == len(lessons) {
				result.RequiredCompleted++
			}
		}
	}

	if len(lessonIDs) > 0 {
		if err := database.Model(&domain.LearningTimeReport{}).
			Where("tenant_id = ? AND user_id = ? AND created_at >= ? AND created_at < ? AND lesson_id IN ?", tenantID, userID, yesterdayStart.UTC(), todayStart.UTC(), lessonIDs).
			Select("COUNT(DISTINCT lesson_id)").
			Scan(&result.YesterdayLessonCount).Error; err != nil {
			return LearnerMotivationSnapshot{}, err
		}
	}
	for _, course := range visible {
		if len(course.lessons) == 0 {
			continue
		}
		allComplete := true
		completedYesterday := false
		for _, lesson := range course.lessons {
			progress, exists := course.progress[lesson.ID]
			if !exists || progress.Status != 2 {
				allComplete = false
				continue
			}
			if progress.CompletedAt != nil && !progress.CompletedAt.Before(yesterdayStart.UTC()) && progress.CompletedAt.Before(todayStart.UTC()) {
				result.YesterdayCompletedLessonCount++
				completedYesterday = true
			}
		}
		if allComplete && completedYesterday {
			result.YesterdayCompletedCourseCount++
		}
	}

	if result.YesterdaySeconds > 0 {
		type cohortRow struct {
			UserID  string
			Seconds int64
		}
		var cohort []cohortRow
		if err := database.Table("learning_daily_stats AS stats").
			Select("stats.user_id, SUM(stats.duration_seconds) AS seconds").
			Joins("JOIN users AS users ON users.id = stats.user_id AND users.tenant_id = stats.tenant_id").
			Where("stats.tenant_id = ? AND stats.study_date = ? AND users.role = ? AND users.status = ?", tenantID, yesterday, "learner", 1).
			Group("stats.user_id").
			Having("SUM(stats.duration_seconds) > 0").
			Scan(&cohort).Error; err != nil {
			return LearnerMotivationSnapshot{}, err
		}
		result.ActiveLearnerCount = len(cohort)
		lower := 0
		for _, row := range cohort {
			if row.Seconds < result.YesterdaySeconds {
				lower++
			}
		}
		if result.ActiveLearnerCount > 0 {
			result.ExceededPercent = lower * 100 / result.ActiveLearnerCount
		}
	}
	result.RecommendedCourse = selectMotivationCourse(visible)
	return result, nil
}

func (repository *learnerMotivationGORMRepository) IssuePrompt(
	ctx context.Context,
	kind, promptDate string,
	expiresAt time.Time,
) (string, error) {
	userID, tenantID, err := learnerMotivationIdentity(ctx)
	if err != nil || !validMotivationPromptKind(kind) || promptDate == "" || expiresAt.IsZero() {
		return "", gorm.ErrRecordNotFound
	}
	database := repository.database.WithContext(ctx)
	for attempt := 0; attempt < 3; attempt++ {
		var state domain.LearnerEngagementState
		if err := database.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&state).Error; err != nil {
			return "", err
		}
		if state.PendingPromptKey != "" && state.PendingPromptKind == kind && state.PendingPromptDate == promptDate &&
			state.PendingPromptExpiresAt != nil && state.PendingPromptExpiresAt.After(time.Now().UTC()) {
			return state.PendingPromptKey, nil
		}
		key := uuid.NewString()
		result := database.Model(&domain.LearnerEngagementState{}).
			Where("id = ? AND tenant_id = ? AND user_id = ? AND updated_at = ?", state.ID, tenantID, userID, state.UpdatedAt).
			Updates(map[string]interface{}{
				"pending_prompt_key": key, "pending_prompt_kind": kind,
				"pending_prompt_date": promptDate, "pending_prompt_expires_at": expiresAt.UTC(),
			})
		if result.Error != nil {
			return "", result.Error
		}
		if result.RowsAffected == 1 {
			return key, nil
		}
	}
	return "", errors.New("issue learner motivation prompt conflict")
}

func (repository *learnerMotivationGORMRepository) AcknowledgePrompt(
	ctx context.Context,
	key string,
	at time.Time,
) error {
	userID, tenantID, err := learnerMotivationIdentity(ctx)
	if err != nil || key == "" {
		return gorm.ErrRecordNotFound
	}
	return repository.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var state domain.LearnerEngagementState
		if err := tx.Where(
			"tenant_id = ? AND user_id = ? AND pending_prompt_key = ?", tenantID, userID, key,
		).First(&state).Error; err != nil {
			return err
		}
		if state.PendingPromptKind == "" || state.PendingPromptDate == "" || state.PendingPromptExpiresAt == nil ||
			!state.PendingPromptExpiresAt.After(at.UTC()) {
			return gorm.ErrRecordNotFound
		}
		if state.PendingPromptKind == "welcome" {
			if state.WelcomeSeenAt != nil {
				return nil
			}
			return tx.Model(&domain.LearnerEngagementState{}).
				Where("id = ? AND pending_prompt_key = ?", state.ID, key).
				Updates(map[string]interface{}{
					"welcome_seen_at": at.UTC(), "last_daily_prompt_date": state.PendingPromptDate,
				}).Error
		}
		if state.PendingPromptKind != "daily_summary" && state.PendingPromptKind != "reengagement" {
			return gorm.ErrRecordNotFound
		}
		if state.LastDailyPromptDate == state.PendingPromptDate {
			return nil
		}
		return tx.Model(&domain.LearnerEngagementState{}).
			Where("id = ? AND pending_prompt_key = ?", state.ID, key).
			Update("last_daily_prompt_date", state.PendingPromptDate).Error
	})
}

func validMotivationPromptKind(kind string) bool {
	return kind == "welcome" || kind == "daily_summary" || kind == "reengagement"
}

func learnerMotivationIdentity(ctx context.Context) (string, string, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || userID == "" || tenantID == "" {
		return "", "", gorm.ErrRecordNotFound
	}
	return userID, tenantID, nil
}

func motivationDailySeconds(database *gorm.DB, tenantID, userID, studyDate string) (int64, error) {
	var seconds int64
	err := database.Model(&domain.LearningDailyStat{}).
		Where("tenant_id = ? AND user_id = ? AND study_date = ?", tenantID, userID, studyDate).
		Select("COALESCE(SUM(duration_seconds), 0)").Scan(&seconds).Error
	return seconds, err
}

func motivationCourseLessons(database *gorm.DB, course domain.Course) ([]domain.CourseLesson, error) {
	var lessons []domain.CourseLesson
	err := database.Table("course_lessons AS lessons").
		Select("lessons.*").
		Joins("JOIN course_chapters AS chapters ON chapters.id = lessons.chapter_id AND chapters.tenant_id = lessons.tenant_id").
		Where("chapters.course_id = ? AND chapters.tenant_id = ?", course.ID, course.TenantID).
		Order("chapters.sort_order ASC, chapters.id ASC, lessons.sort_order ASC, lessons.id ASC").
		Scan(&lessons).Error
	return lessons, err
}

type motivationVisibleCourse struct {
	data       LearnerOverviewCourse
	lessons    []domain.CourseLesson
	progress   map[string]domain.LessonProgress
	enrollment *domain.CourseEnrollment
}

func selectMotivationCourse(courses []motivationVisibleCourse) *RecommendedCourse {
	if len(courses) == 0 {
		return nil
	}
	for _, course := range courses {
		if course.data.RecentLesson == nil || course.data.LastLearnedAt == nil {
			continue
		}
		candidate := recommendedCourse(course.data, *course.data.RecentLesson, course.data.LastPositionSeconds)
		for _, other := range courses {
			if other.data.RecentLesson != nil && other.data.LastLearnedAt != nil && other.data.LastLearnedAt.After(*course.data.LastLearnedAt) {
				candidate = recommendedCourse(other.data, *other.data.RecentLesson, other.data.LastPositionSeconds)
				course = other
			}
		}
		return candidate
	}

	type candidate struct {
		course   motivationVisibleCourse
		priority int
	}
	candidates := make([]candidate, 0, len(courses))
	for _, course := range courses {
		if len(course.lessons) == 0 {
			continue
		}
		unfinished := course.data.CompletedLessonCount < len(course.lessons)
		priority := 3
		if course.enrollment != nil && course.enrollment.AssignmentType == domain.AssignmentRequired && unfinished {
			priority = 0
		} else if course.enrollment == nil && course.data.Course.IsOfficial && course.data.AssignmentType == domain.AssignmentRequired && unfinished {
			priority = 1
		} else if course.enrollment != nil && unfinished {
			priority = 2
		}
		candidates = append(candidates, candidate{course: course, priority: priority})
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(left, right int) bool {
		if candidates[left].priority != candidates[right].priority {
			return candidates[left].priority < candidates[right].priority
		}
		leftEnrollment, rightEnrollment := candidates[left].course.enrollment, candidates[right].course.enrollment
		if leftEnrollment != nil && rightEnrollment != nil && !leftEnrollment.CreatedAt.Equal(rightEnrollment.CreatedAt) {
			return leftEnrollment.CreatedAt.Before(rightEnrollment.CreatedAt)
		}
		if candidates[left].course.data.Course.Title != candidates[right].course.data.Course.Title {
			return candidates[left].course.data.Course.Title < candidates[right].course.data.Course.Title
		}
		return candidates[left].course.data.Course.ID < candidates[right].course.data.Course.ID
	})
	chosen := candidates[0].course
	lesson := chosen.lessons[0]
	position := 0
	for _, item := range chosen.lessons {
		progress, exists := chosen.progress[item.ID]
		if !exists || progress.Status != 2 {
			lesson = item
			if exists {
				position = progress.LastPositionSeconds
			}
			break
		}
	}
	return recommendedCourse(chosen.data, lesson, position)
}

func recommendedCourse(course LearnerOverviewCourse, lesson domain.CourseLesson, position int) *RecommendedCourse {
	return &RecommendedCourse{
		ID: course.Course.ID, Title: course.Course.Title, AssignmentType: course.AssignmentType,
		LessonCount: course.LessonCount, ProgressPercent: course.ProgressPercent,
		LessonID: lesson.ID, LessonTitle: lesson.Title, LastPositionSeconds: position,
	}
}
