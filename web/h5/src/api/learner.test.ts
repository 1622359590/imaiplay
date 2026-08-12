import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  apiClient: { get: vi.fn() },
  unwrap: vi.fn(),
}))

import { apiClient, unwrap } from './client'
import { getLearnerOverview, mergeCourseOverview } from './learner'
import type { Course } from '../types/course'

describe('H5 learner overview', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('maps real course progress and recent lesson from learner overview', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: {} })
    vi.mocked(unwrap).mockReturnValueOnce({
      required_completed: 1,
      required_total: 2,
      today_learning_seconds: 180,
      total_learning_seconds: 7200,
      courses: [{
        course: {
          id: 'course-1',
          title: '销售基础',
          course_type: 'required',
          category: { id: 'sales', name: '销售' },
        },
        assignment_type: 'required',
        lesson_count: 8,
        completed_lesson_count: 3,
        progress_percent: 38,
        recent_lesson: {
          id: 'lesson-3',
          title: '异议处理',
          duration_seconds: 600,
          last_position_seconds: 125,
        },
      }],
    })

    await expect(getLearnerOverview()).resolves.toMatchObject({
      requiredCompleted: 1,
      requiredTotal: 2,
      courses: [{
        id: 'course-1',
        progressPercent: 38,
        completedLessonCount: 3,
        recentLesson: {
          id: 'lesson-3',
          lastPositionSeconds: 125,
        },
      }],
    })
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/learner/overview')
  })

  it('merges overview progress into matching course data without replacing course detail', () => {
    const course: Course = {
      id: 'course-1',
      title: '详情标题',
      description: '详情描述',
      instructor: '企业讲师',
      progress: 0,
      lessonCount: 2,
      duration: 12,
      category: '销售',
      courseType: 'required',
    }

    const [merged] = mergeCourseOverview([course], {
      requiredCompleted: 0,
      requiredTotal: 1,
      todayLearningSeconds: 0,
      totalLearningSeconds: 0,
      courses: [{
        id: 'course-1',
        progressPercent: 62,
        lessonCount: 8,
        completedLessonCount: 5,
        recentLesson: {
          id: 'lesson-5',
          title: '成交推进',
          durationSeconds: 300,
          lastPositionSeconds: 80,
        },
      }],
    })

    expect(merged).toMatchObject({
      title: '详情标题',
      description: '详情描述',
      progress: 62,
      lessonCount: 8,
      completedLessonCount: 5,
      recentLesson: { id: 'lesson-5' },
    })
  })
})
