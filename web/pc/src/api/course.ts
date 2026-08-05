import { apiClient } from './client';

export interface Lesson {
  id: string;
  title: string;
  contentType: 'video' | 'document' | 'text';
  contentURL: string;
  resourceID?: string;
  resourceType?: 'video' | 'document';
  durationSeconds: number;
  sortOrder: number;
}

export interface Chapter {
  id: string;
  title: string;
  sortOrder: number;
  lessons: Lesson[];
}

export interface Course {
  id: string;
  title: string;
  description?: string;
  coverImage?: string;
  instructor?: string;
  category?: string;
  lessonCount?: number;
  studentCount?: number;
  durationSeconds?: number;
  progressPercent?: number;
  chapters?: Chapter[];
  lastLearnedAt?: string;
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

interface RawLesson {
  id: string;
  title: string;
  content_type?: 'video' | 'document' | 'text';
  content_url?: string;
  duration_seconds?: number;
  resource_id?: string | null;
  resource_type?: 'video' | 'document';
  sort_order?: number;
}

interface RawChapter {
  id: string;
  title: string;
  sort_order?: number;
  lessons?: RawLesson[] | null;
}

interface RawCourse {
  id: string;
  title: string;
  description?: string;
  cover_image?: string;
  instructor?: string;
  category?: string;
  lesson_count?: number;
  student_count?: number;
  duration_seconds?: number;
  progress_percent?: number;
  last_learned_at?: string;
  category_id?: string | null;
  is_official?: boolean;
}

interface RawCourseListPayload {
  items?: RawCourse[] | null;
  list?: RawCourse[] | null;
}

interface RawCourseDetail {
  course: RawCourse;
  chapters?: RawChapter[] | null;
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
    coverImage: course.cover_image,
    instructor: course.instructor,
    category: course.category,
    lessonCount: course.lesson_count,
    studentCount: course.student_count,
    durationSeconds: course.duration_seconds,
    progressPercent: course.progress_percent,
    lastLearnedAt: course.last_learned_at,
    categoryId: course.category_id ?? undefined,
    isOfficial: course.is_official,
  };
}

function normalizeList(payload: RawCourse[] | RawCourseListPayload): Course[] {
  const items = Array.isArray(payload) ? payload : payload.items ?? payload.list ?? [];
  return items.map(mapCourse);
}

function mapLesson(lesson: RawLesson): Lesson {
  return {
    id: lesson.id,
    title: lesson.title,
    contentType: lesson.content_type ?? 'text',
    contentURL: lesson.content_url ?? '',
    resourceID: lesson.resource_id ?? undefined,
    resourceType: lesson.resource_type,
    durationSeconds: Math.max(0, Math.floor(lesson.duration_seconds ?? 0)),
    sortOrder: Math.max(0, Math.floor(lesson.sort_order ?? 0)),
  };
}

function mapChapter(chapter: RawChapter): Chapter {
  return {
    id: chapter.id,
    title: chapter.title,
    sortOrder: Math.max(0, Math.floor(chapter.sort_order ?? 0)),
    lessons: (chapter.lessons ?? []).map(mapLesson),
  };
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
      return { ...course, lessonCount: countLessons(detail.chapters) };
    } catch {
      return course;
    }
  }));
}

export async function getCourses(): Promise<Course[]> {
  const response = await apiClient.get<RawCourse[] | RawCourseListPayload>('/api/v1/courses');
  return enrichLessonCounts(normalizeList(response.data), getCourse);
}

export async function getCourse(id: string): Promise<Course> {
  const response = await apiClient.get<RawCourseDetail>(`/api/v1/courses/${id}`);
  return {
    ...mapCourse(response.data.course),
    chapters: (response.data.chapters ?? []).map(mapChapter),
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
