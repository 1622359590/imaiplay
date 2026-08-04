import { describe, expect, it, vi } from 'vitest';
import { countLessons, enrichLessonCounts, type Course } from './course';

describe('PC learner course presentation data', () => {
  it('counts lessons across chapters', () => {
    expect(countLessons([
      {
        id: 'chapter-1',
        title: '第一章',
        lessons: [
          { id: 'lesson-1', title: '课时一' },
          { id: 'lesson-2', title: '课时二' },
        ],
      },
      {
        id: 'chapter-2',
        title: '第二章',
        lessons: [{ id: 'lesson-3', title: '课时三' }],
      },
    ])).toBe(3);
  });

  it('keeps a course usable when detail enrichment fails', async () => {
    const courses: Course[] = [
      { id: 'ok', title: '可用课程' },
      { id: 'failed', title: '详情失败课程' },
    ];
    const loadDetail = vi.fn(async (id: string): Promise<Course> => {
      if (id === 'failed') throw new Error('detail failed');
      return {
        ...courses[0],
        chapters: [
          {
            id: 'chapter',
            title: '章节',
            lessons: [{ id: 'lesson', title: '课时' }],
          },
        ],
      };
    });

    await expect(enrichLessonCounts(courses, loadDetail)).resolves.toEqual([
      { id: 'ok', title: '可用课程', lesson_count: 1 },
      { id: 'failed', title: '详情失败课程' },
    ]);
  });
});
