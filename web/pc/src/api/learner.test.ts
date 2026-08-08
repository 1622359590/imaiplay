import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import { getLearnerOverview, getRecentLearning } from './learner';

vi.mock('./client', () => ({
  apiClient: { get: vi.fn(), post: vi.fn() },
}));

afterEach(() => {
  vi.restoreAllMocks();
});

describe('learner API boundary normalization', () => {
  it('maps the exact snake_case overview DTO into one camelCase view model', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        required_completed: 1,
        required_total: 2,
        today_learning_seconds: 90,
        total_learning_seconds: 3600,
        categories: [{ id: 'sales', name: '销售' }],
        courses: [{
          course: {
            id: 'course-1',
            title: '销售基础',
            description: '课程介绍',
            cover_image: 'https://cdn.example.com/course.png',
            category: { id: 'sales', name: '销售' },
			course_type: 'optional',
          },
          assignment_type: 'required',
          lesson_count: 2,
          completed_lesson_count: 1,
          progress_percent: 150,
          last_learned_at: '2026-08-05T12:00:00+08:00',
          recent_lesson: {
            id: 'lesson-1',
            title: '开场',
            duration_seconds: 120,
            last_position_seconds: 42,
          },
        }],
      },
    });

    await expect(getLearnerOverview()).resolves.toEqual({
      requiredCompleted: 1,
      requiredTotal: 2,
      todayLearningSeconds: 90,
      totalLearningSeconds: 3600,
      categories: [{ id: 'sales', name: '销售' }],
      courses: [{
        id: 'course-1',
        title: '销售基础',
        description: '课程介绍',
        coverImage: 'https://cdn.example.com/course.png',
		courseType: 'optional',
        category: { id: 'sales', name: '销售' },
        lessonCount: 2,
        completedLessonCount: 1,
        progressPercent: 100,
        lastLearnedAt: '2026-08-05T12:00:00+08:00',
        recentLesson: {
          id: 'lesson-1',
          title: '开场',
          durationSeconds: 120,
          lastPositionSeconds: 42,
        },
      }],
    });
  });

  it('maps the recent-learning page without leaking snake_case fields', async () => {
    const request = vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        items: [{
          course: {
            id: 'course-1',
            title: '销售基础',
            description: '',
            cover_image: '',
            category: { id: 'sales', name: '销售' },
			course_type: 'required',
          },
          recent_lesson: {
            id: 'lesson-1',
            title: '开场',
            duration_seconds: 120,
            last_position_seconds: 42,
          },
          progress_percent: -10,
          last_position_seconds: 42,
          last_learned_at: '2026-08-05T12:00:00+08:00',
        }],
        total: 1,
      },
    });

    await expect(getRecentLearning(20, 5)).resolves.toEqual({
      items: [{
        course: {
          id: 'course-1',
          title: '销售基础',
          description: '',
          coverImage: '',
          category: { id: 'sales', name: '销售' },
		  courseType: 'required',
        },
        recentLesson: {
          id: 'lesson-1',
          title: '开场',
          durationSeconds: 120,
          lastPositionSeconds: 42,
        },
        progressPercent: 0,
        lastPositionSeconds: 42,
        lastLearnedAt: '2026-08-05T12:00:00+08:00',
      }],
      total: 1,
    });
    expect(request).toHaveBeenCalledWith('/api/v1/recent-learning', {
      params: { offset: 20, limit: 5 },
    });
  });

  it('normalizes absent or null collections to empty arrays', async () => {
    vi.spyOn(apiClient, 'get')
      .mockResolvedValueOnce({
        data: {
          required_completed: 0,
          required_total: 0,
          today_learning_seconds: 0,
          total_learning_seconds: 0,
          categories: null,
        },
      })
      .mockResolvedValueOnce({ data: { items: null, total: 0 } });

    await expect(getLearnerOverview()).resolves.toMatchObject({
      categories: [],
      courses: [],
    });
    await expect(getRecentLearning()).resolves.toEqual({ items: [], total: 0 });
  });

  it('rejects an invalid assignment discriminant with an explicit protocol error', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        required_completed: 0,
        required_total: 1,
        today_learning_seconds: 0,
        total_learning_seconds: 0,
        categories: [],
        courses: [{
		  course: { id: 'course-1', title: '课程', category: null, course_type: 'mandatory' },
		  assignment_type: 'required',
          lesson_count: 1,
          completed_lesson_count: 0,
          progress_percent: 0,
        }],
      },
    });

    await expect(getLearnerOverview()).rejects.toThrow(
      'Invalid learner assignment type',
    );
  });
});
