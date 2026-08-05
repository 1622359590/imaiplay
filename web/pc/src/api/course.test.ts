import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import {
  countLessons,
  downloadCourseMaterial,
  enrichLessonCounts,
  getCourse,
  getCourses,
  getResourceFile,
  type Course,
} from './course';

vi.mock('./client', () => ({
  apiClient: { get: vi.fn(), post: vi.fn() },
}));

afterEach(() => {
  vi.restoreAllMocks();
});

describe('PC learner course presentation data', () => {
  it('counts lessons across chapters', () => {
    expect(countLessons([
      {
        id: 'chapter-1',
        title: '第一章',
        sortOrder: 1,
        lessons: [
          { id: 'lesson-1', title: '课时一', contentType: 'text', contentURL: '', durationSeconds: 0, sortOrder: 1 },
          { id: 'lesson-2', title: '课时二', contentType: 'text', contentURL: '', durationSeconds: 0, sortOrder: 2 },
        ],
      },
      {
        id: 'chapter-2',
        title: '第二章',
        sortOrder: 2,
        lessons: [{ id: 'lesson-3', title: '课时三', contentType: 'text', contentURL: '', durationSeconds: 0, sortOrder: 1 }],
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
            sortOrder: 1,
            lessons: [{ id: 'lesson', title: '课时', contentType: 'text', contentURL: '', durationSeconds: 0, sortOrder: 1 }],
          },
        ],
      };
    });

    await expect(enrichLessonCounts(courses, loadDetail)).resolves.toEqual([
      { id: 'ok', title: '可用课程', lessonCount: 1 },
      { id: 'failed', title: '详情失败课程' },
    ]);
  });

  it('normalizes list courses without leaking raw keys', async () => {
    vi.spyOn(apiClient, 'get')
      .mockResolvedValueOnce({
        data: {
          items: [{
            id: 'course-1',
            title: '课程',
            cover_image: 'https://cdn.example.com/course.png',
            category_id: null,
            is_official: false,
          }],
        },
      })
      .mockResolvedValueOnce({
        data: {
          course: {
            id: 'course-1',
            title: '课程',
            cover_image: 'https://cdn.example.com/course.png',
            category_id: null,
            is_official: false,
          },
          chapters: [],
          materials: [],
        },
      });

    await expect(getCourses()).resolves.toEqual([{
      id: 'course-1',
      title: '课程',
      description: undefined,
      coverImage: 'https://cdn.example.com/course.png',
      categoryId: undefined,
      isOfficial: false,
      lessonCount: 0,
    }]);
  });
});

describe('PC learner course materials', () => {
  it('maps ordered material metadata from course detail', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        course: {
          id: 'course-1',
          title: '课程',
          cover_image: 'https://cdn.example.com/course.png',
          category_id: null,
          is_official: true,
        },
        chapters: [{
          id: 'chapter-1',
          title: '第一章',
          sort_order: 2,
          lessons: [{
            id: 'lesson-1',
            title: '开场',
            content_type: 'video',
            content_url: 'https://cdn.example.com/lesson.mp4',
            resource_id: 'resource-1',
            resource_type: 'video',
            duration_seconds: 61,
            sort_order: 3,
          }],
        }],
        materials: [{
          id: 'material-1',
          display_name: '入门手册.pdf',
          resource_type: 'attachment',
          size_bytes: 4096,
        }],
      },
    });
    const course = await getCourse('course-1');
    expect(course).toMatchObject({
      coverImage: 'https://cdn.example.com/course.png',
      categoryId: undefined,
      isOfficial: true,
      chapters: [{
        id: 'chapter-1',
        title: '第一章',
        sortOrder: 2,
        lessons: [{
          id: 'lesson-1',
          title: '开场',
          contentType: 'video',
          contentURL: 'https://cdn.example.com/lesson.mp4',
          resourceID: 'resource-1',
          resourceType: 'video',
          durationSeconds: 61,
          sortOrder: 3,
        }],
      }],
      materials: [{
        id: 'material-1',
        displayName: '入门手册.pdf',
        sizeBytes: 4096,
        resourceType: 'attachment',
      }],
    });
    expect(course).not.toHaveProperty('cover_image');
    expect(course).not.toHaveProperty('category_id');
    expect(course).not.toHaveProperty('is_official');
    expect(course.chapters?.[0]).not.toHaveProperty('sort_order');
    expect(course.chapters?.[0].lessons[0]).not.toHaveProperty('content_type');
    expect(course.chapters?.[0].lessons[0]).not.toHaveProperty('content_url');
    expect(course.chapters?.[0].lessons[0]).not.toHaveProperty('resource_id');
    expect(course.chapters?.[0].lessons[0]).not.toHaveProperty('resource_type');
    expect(course.chapters?.[0].lessons[0]).not.toHaveProperty('duration_seconds');
    expect(course.chapters?.[0].lessons[0]).not.toHaveProperty('sort_order');
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

describe('PC learner video playback', () => {
  it('requests a streaming playback URL instead of buffering the full resource', async () => {
    const request = vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: { url: '/api/v1/resource-playback/resource-1?ticket=signed' },
    });

    await expect(getResourceFile('resource-1')).resolves.toBe(
      '/api/v1/resource-playback/resource-1?ticket=signed',
    );
    expect(request).toHaveBeenCalledWith(
      '/api/v1/resources/resource-1/playback-url',
    );
  });
});
