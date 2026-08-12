import { apiClient, unwrap, type ApiEnvelope } from './client'
import type { Course } from '../types/course'

export interface LearnerRecentLesson {
  id: string
  title: string
  durationSeconds: number
  lastPositionSeconds: number
}

export interface LearnerCourseProgress {
  id: string
  lessonCount: number
  completedLessonCount: number
  progressPercent: number
  recentLesson?: LearnerRecentLesson
}

export interface LearnerOverview {
  requiredCompleted: number
  requiredTotal: number
  todayLearningSeconds: number
  totalLearningSeconds: number
  courses: LearnerCourseProgress[]
}

export interface LearnerCourseView extends Course {
  completedLessonCount: number
  recentLesson?: LearnerRecentLesson
}

interface RawRecentLesson {
  id: string
  title: string
  duration_seconds: number
  last_position_seconds: number
}

interface RawLearnerCourse {
  course: { id: string }
  lesson_count: number
  completed_lesson_count: number
  progress_percent: number
  recent_lesson?: RawRecentLesson
}

interface RawLearnerOverview {
  required_completed: number
  required_total: number
  today_learning_seconds: number
  total_learning_seconds: number
  courses?: RawLearnerCourse[] | null
}

function nonNegativeInteger(value: number): number {
  return Number.isFinite(value) ? Math.max(0, Math.floor(value)) : 0
}

function boundedProgress(value: number, lessonCount: number): number {
  if (lessonCount === 0) return 0
  return Math.min(100, nonNegativeInteger(value))
}

function mapRecentLesson(lesson: RawRecentLesson): LearnerRecentLesson {
  return {
    id: lesson.id,
    title: lesson.title,
    durationSeconds: nonNegativeInteger(lesson.duration_seconds),
    lastPositionSeconds: nonNegativeInteger(lesson.last_position_seconds),
  }
}

export async function getLearnerOverview(): Promise<LearnerOverview> {
  const response = await apiClient.get<ApiEnvelope<RawLearnerOverview>>('/api/v1/learner/overview')
  const payload = unwrap(response)
  return {
    requiredCompleted: nonNegativeInteger(payload.required_completed),
    requiredTotal: nonNegativeInteger(payload.required_total),
    todayLearningSeconds: nonNegativeInteger(payload.today_learning_seconds),
    totalLearningSeconds: nonNegativeInteger(payload.total_learning_seconds),
    courses: (payload.courses ?? []).map((item) => {
      const lessonCount = nonNegativeInteger(item.lesson_count)
      return {
        id: item.course.id,
        lessonCount,
        completedLessonCount: Math.min(lessonCount, nonNegativeInteger(item.completed_lesson_count)),
        progressPercent: boundedProgress(item.progress_percent, lessonCount),
        recentLesson: item.recent_lesson ? mapRecentLesson(item.recent_lesson) : undefined,
      }
    }),
  }
}

export function mergeCourseOverview(
  courses: Course[],
  overview: LearnerOverview,
): LearnerCourseView[] {
  const progressByCourse = new Map(overview.courses.map((course) => [course.id, course]))
  return courses.map((course) => {
    const progress = progressByCourse.get(course.id)
    return {
      ...course,
      progress: progress?.progressPercent ?? course.progress,
      lessonCount: progress?.lessonCount ?? course.lessonCount,
      completedLessonCount: progress?.completedLessonCount ?? 0,
      recentLesson: progress?.recentLesson,
    }
  })
}

function emptyLearnerOverview(): LearnerOverview {
  return {
    requiredCompleted: 0,
    requiredTotal: 0,
    todayLearningSeconds: 0,
    totalLearningSeconds: 0,
    courses: [],
  }
}

export async function loadCoursesWithOptionalOverview(
  loadCourses: () => Promise<Course[]>,
  loadOverview: () => Promise<LearnerOverview> = getLearnerOverview,
): Promise<LearnerCourseView[]> {
  const [courses, overview] = await Promise.all([
    loadCourses(),
    loadOverview().catch(() => emptyLearnerOverview()),
  ])
  return mergeCourseOverview(courses, overview)
}

export async function loadCourseWithOptionalOverview(
  loadCourse: () => Promise<Course>,
  loadOverview: () => Promise<LearnerOverview> = getLearnerOverview,
): Promise<LearnerCourseView> {
  const [course] = await loadCoursesWithOptionalOverview(
    async () => [await loadCourse()],
    loadOverview,
  )
  return course
}
