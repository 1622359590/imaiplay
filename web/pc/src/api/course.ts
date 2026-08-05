import { apiClient } from './client';

export interface Lesson {
  id: string;
  title: string;
  content_type?: 'video' | 'document' | 'text';
  content_url?: string;
  resource_id?: string;
  duration?: number;
  learned?: boolean;
  progress?: number;
}

export interface Chapter {
  id: string;
  title: string;
  lessons: Lesson[];
}

export interface Course {
  id: string;
  title: string;
  description?: string;
  cover?: string;
  instructor?: string;
  category?: string;
  lesson_count?: number;
  student_count?: number;
  duration?: number;
  progress?: number;
  chapters?: Chapter[];
  last_learned_at?: string;
  materials?: CourseMaterial[];
  categoryId?: string;
  isOfficial?: boolean;
}

export interface CourseMaterial {
  id: string;
  displayName: string;
  sizeBytes: number;
  resourceType: 'attachment';
}

interface CourseListPayload {
  items?: Course[];
  list?: Course[];
}

interface RawLesson {
  id: string;
  title: string;
  content_type?: 'video' | 'document' | 'text';
  content_url?: string;
  duration_seconds?: number;
  resource_id?: string;
}

interface RawChapter {
  id: string;
  title: string;
  lessons?: RawLesson[];
}

interface RawCourse {
  id: string;
  title: string;
  description?: string;
  cover_image?: string;
  category_id?: string;
  is_official?: boolean;
}

interface RawCourseDetail {
  course: RawCourse;
  chapters: RawChapter[];
  materials?: Array<{
    id: string;
    display_name: string;
    resource_type: 'attachment';
    size_bytes: number;
  }>;
}

function mapCourse(course: RawCourse): Course {
  return {
    id: course.id,
    title: course.title,
    description: course.description,
    cover: course.cover_image,
    categoryId: course.category_id,
    isOfficial: course.is_official,
  };
}

function normalizeList(payload: Course[] | CourseListPayload): Course[] {
  const items = Array.isArray(payload) ? payload : payload.items ?? payload.list ?? [];
  return items.map((course) => mapCourse(course as RawCourse));
}

export function countLessons(chapters: Chapter[] = []): number {
  return chapters.reduce((total, chapter) => total + chapter.lessons.length, 0);
}

export async function enrichLessonCounts(
  courses: Course[],
  loadDetail: (id: string) => Promise<Course>,
): Promise<Course[]> {
  return Promise.all(courses.map(async (course) => {
    try {
      const detail = await loadDetail(course.id);
      return { ...course, lesson_count: countLessons(detail.chapters) };
    } catch {
      return course;
    }
  }));
}

export async function getCourses(): Promise<Course[]> {
  const response = await apiClient.get<Course[] | CourseListPayload>('/api/v1/courses');
  return enrichLessonCounts(normalizeList(response.data), getCourse);
}

export async function getCourse(id: string): Promise<Course> {
  const response = await apiClient.get<RawCourseDetail>(`/api/v1/courses/${id}`);
  return {
    ...mapCourse(response.data.course),
    chapters: response.data.chapters.map((chapter) => ({
      id: chapter.id,
      title: chapter.title,
      lessons: (chapter.lessons ?? []).map((lesson) => ({
        id: lesson.id,
        title: lesson.title,
        content_type: lesson.content_type,
        content_url: lesson.content_url,
        resource_id: lesson.resource_id,
        duration: Math.ceil((lesson.duration_seconds ?? 0) / 60),
      })),
    })),
    materials: (response.data.materials ?? []).map((material) => ({
      id: material.id,
      displayName: material.display_name,
      sizeBytes: material.size_bytes,
      resourceType: material.resource_type,
    })),
  };
}

export async function downloadCourseMaterial(id: string): Promise<Blob> {
  const response = await apiClient.get<Blob>(
    `/api/v1/course-materials/${encodeURIComponent(id)}/download`,
    { responseType: 'blob' },
  );
  return response.data;
}

export async function getResourceFile(id: string): Promise<string> {
  const response = await apiClient.get<{ url: string }>(
    `/api/v1/resources/${id}/playback-url`,
  );
  return response.data.url;
}

export async function getRecentCourses(): Promise<Course[]> {
  const response = await apiClient.get<{
    items: Array<{
      course: RawCourse;
      lesson: RawLesson;
      progress: { progress_percent: number };
      last_learned_at: string;
    }>;
  }>('/api/v1/recent-learning');
  return response.data.items.map((item) => ({
    ...mapCourse(item.course),
    progress: item.progress.progress_percent,
    last_learned_at: item.last_learned_at,
  }));
}
