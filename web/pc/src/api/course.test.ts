import { describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import {
  countLessons,
  downloadCourseMaterial,
  enrichLessonCounts,
  getCourse,
  type Course,
} from './course';

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

describe('PC learner course materials', () => {
  it('maps ordered material metadata from course detail', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        course: { id: 'course-1', title: '课程' },
        chapters: [],
        materials: [{
          id: 'material-1',
          display_name: '入门手册.pdf',
          resource: { resource_type: 'attachment', size_bytes: 4096 },
        }],
      },
    });
    await expect(getCourse('course-1')).resolves.toMatchObject({
      materials: [{
        id: 'material-1',
        displayName: '入门手册.pdf',
        sizeBytes: 4096,
        resourceType: 'attachment',
      }],
    });
  });

  it('downloads a material as a blob through its protected route', async () => {
    const blob = new Blob(['guide']);
    const request = vi.spyOn(apiClient, 'get').mockResolvedValueOnce({ data: blob });
    await expect(downloadCourseMaterial('material-1')).resolves.toBe(blob);
    expect(request).toHaveBeenCalledWith(
      '/api/v1/course-materials/material-1/download',
      { responseType: 'blob' },
    );
  });
});
