package repository

import (
	"context"
	"sort"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

type learnerOverviewGORMRepository struct {
	database *gorm.DB
}

func NewLearnerOverviewRepository(database *gorm.DB) LearnerOverviewRepository {
	return &learnerOverviewGORMRepository{database: database}
}

func (repo *learnerOverviewGORMRepository) Get(
	ctx context.Context,
	studyDate string,
) (LearnerOverviewData, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "learner" || userID == "" || tenantID == "" {
		return LearnerOverviewData{}, gorm.ErrRecordNotFound
	}
	database := repo.database.WithContext(ctx)
	result := LearnerOverviewData{Courses: []LearnerOverviewCourse{}}

	var enrollments []domain.CourseEnrollment
	if err := database.Where(
		"tenant_id = ? AND user_id = ? AND status = ?",
		tenantID, userID, 1,
	).Find(&enrollments).Error; err != nil {
		return LearnerOverviewData{}, err
	}
	for _, enrollment := range enrollments {
		course, err := repo.visibleCourse(database, tenantID, enrollment.CourseID)
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			return LearnerOverviewData{}, err
		}
		courseData, err := repo.aggregateCourse(
			database, tenantID, userID, course.CourseType, course,
		)
		if err != nil {
			return LearnerOverviewData{}, err
		}
		result.Courses = append(result.Courses, courseData)
	}
	sort.Slice(result.Courses, func(left, right int) bool {
		if result.Courses[left].Course.Title == result.Courses[right].Course.Title {
			return result.Courses[left].Course.ID < result.Courses[right].Course.ID
		}
		return result.Courses[left].Course.Title < result.Courses[right].Course.Title
	})

	if err := database.Model(&domain.LearningDailyStat{}).
		Where("tenant_id = ? AND user_id = ? AND study_date = ?", tenantID, userID, studyDate).
		Select("COALESCE(SUM(duration_seconds), 0)").
		Scan(&result.TodayLearningSeconds).Error; err != nil {
		return LearnerOverviewData{}, err
	}
	if err := database.Model(&domain.LearningDailyStat{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Select("COALESCE(SUM(duration_seconds), 0)").
		Scan(&result.TotalLearningSeconds).Error; err != nil {
		return LearnerOverviewData{}, err
	}
	return result, nil
}

func (repo *learnerOverviewGORMRepository) visibleCourse(
	database *gorm.DB,
	tenantID, courseID string,
) (*domain.Course, error) {
	var course domain.Course
	err := database.Where(`
		id = ? AND status = ? AND (
			(tenant_id = ? AND is_official = ?) OR
			(tenant_id = ? AND is_official = ? AND EXISTS (
				SELECT 1 FROM tenant_official_courses toc
				WHERE toc.tenant_id = ? AND toc.course_id = courses.id AND toc.enabled = ?
			))
		)
	`, courseID, 1, tenantID, false, "", true, tenantID, true).
		First(&course).Error
	return &course, err
}

func (repo *learnerOverviewGORMRepository) aggregateCourse(
	database *gorm.DB,
	tenantID, userID, assignmentType string,
	course *domain.Course,
) (LearnerOverviewCourse, error) {
	result := LearnerOverviewCourse{
		Course: *course, AssignmentType: assignmentType,
	}
	if course.CategoryID != nil {
		var category domain.CourseCategory
		err := database.Where(
			"id = ? AND tenant_id = ? AND status = ?",
			*course.CategoryID, course.TenantID, 1,
		).First(&category).Error
		if err == nil {
			result.Category = &category
		} else if err != gorm.ErrRecordNotFound {
			return LearnerOverviewCourse{}, err
		}
	}

	var lessons []domain.CourseLesson
	if err := database.Table("course_lessons AS cl").
		Select("cl.*").
		Joins("JOIN course_chapters AS cc ON cc.id = cl.chapter_id AND cc.tenant_id = cl.tenant_id").
		Where("cc.course_id = ? AND cc.tenant_id = ?", course.ID, course.TenantID).
		Order("cc.sort_order ASC, cc.id ASC, cl.sort_order ASC, cl.id ASC").
		Scan(&lessons).Error; err != nil {
		return LearnerOverviewCourse{}, err
	}
	result.LessonCount = len(lessons)
	if len(lessons) == 0 {
		return result, nil
	}

	lessonIDs := make([]string, 0, len(lessons))
	lessonsByID := make(map[string]domain.CourseLesson, len(lessons))
	for _, lesson := range lessons {
		lessonIDs = append(lessonIDs, lesson.ID)
		lessonsByID[lesson.ID] = lesson
	}
	var progressItems []domain.LessonProgress
	if err := database.Where(
		"tenant_id = ? AND user_id = ? AND lesson_id IN ?",
		tenantID, userID, lessonIDs,
	).Find(&progressItems).Error; err != nil {
		return LearnerOverviewCourse{}, err
	}
	progressByLesson := make(map[string]domain.LessonProgress, len(progressItems))
	for _, progress := range progressItems {
		progressByLesson[progress.LessonID] = progress
	}
	progressSum := 0
	for _, lesson := range lessons {
		progress, exists := progressByLesson[lesson.ID]
		if !exists {
			continue
		}
		percent := progress.ProgressPercent
		if percent < 0 {
			percent = 0
		} else if percent > 100 {
			percent = 100
		}
		progressSum += percent
		if progress.Status == 2 {
			result.CompletedLessonCount++
		}
		if result.LastLearnedAt == nil || progress.UpdatedAt.After(*result.LastLearnedAt) ||
			(progress.UpdatedAt.Equal(*result.LastLearnedAt) && progress.LessonID < result.RecentLesson.ID) {
			updatedAt := progress.UpdatedAt
			recentLesson := lessonsByID[progress.LessonID]
			result.LastLearnedAt = &updatedAt
			result.RecentLesson = &recentLesson
			result.LastPositionSeconds = progress.LastPositionSeconds
		}
	}
	result.ProgressPercent = (progressSum + result.LessonCount/2) / result.LessonCount
	return result, nil
}
