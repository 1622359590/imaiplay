import { apiClient } from './client';

export interface Lesson {
  id: string;
  title: string;
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
}

interface CourseListPayload {
  items?: Course[];
  list?: Course[];
}

interface RawLesson {
  id: string;
  title: string;
  duration_seconds?: number;
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
}

interface RawCourseDetail {
  course: RawCourse;
  chapters: RawChapter[];
}

function mapCourse(course: RawCourse): Course {
  return {
    ...course,
    cover: course.cover_image,
  };
}

function normalizeList(payload: Course[] | CourseListPayload): Course[] {
  const items = Array.isArray(payload) ? payload : payload.items ?? payload.list ?? [];
  return items.map((course) => mapCourse(course as RawCourse));
}

export async function getCourses(): Promise<Course[]> {
  const response = await apiClient.get<Course[] | CourseListPayload>('/api/v1/courses');
  return normalizeList(response.data);
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
        duration: Math.ceil((lesson.duration_seconds ?? 0) / 60),
      })),
    })),
  };
}

export async function getRecentCourses(): Promise<Course[]> {
  return [];
}
