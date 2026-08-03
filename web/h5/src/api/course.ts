import { apiClient, unwrap, type ApiEnvelope } from './client'
import type { Chapter, Course, CourseList } from '../types/course'

interface RawLesson {
  id: string
  title: string
  content_type?: 'video' | 'document' | 'text'
  content_url?: string
  duration_seconds?: number
  resource_id?: string
}

interface RawChapter {
  id: string
  title: string
  lessons?: RawLesson[]
}

interface RawCourse {
  id: string
  title: string
  description?: string
  cover_image?: string
}

interface RawCourseDetail {
  course: RawCourse
  chapters: RawChapter[]
}

function mapCourse(course: RawCourse): Course {
  return {
    id: course.id,
    title: course.title,
    description: course.description ?? '',
    cover: course.cover_image ? `url("${course.cover_image}") center/cover` : 'linear-gradient(135deg, #0e55ce, #47a2ff)',
    instructor: '企业讲师',
    progress: 0,
    duration: 0,
    category: '企业课程',
  }
}

export function countLessons(chapters: Chapter[] = []): number {
  return chapters.reduce((total, chapter) => total + chapter.lessons.length, 0)
}

export async function enrichLessonCounts(
  courses: Course[],
  loadDetail: (id: string) => Promise<Course>,
): Promise<Course[]> {
  return Promise.all(courses.map(async (course) => {
    try {
      const detail = await loadDetail(course.id)
      return { ...course, lessonCount: countLessons(detail.chapters) }
    } catch {
      const withoutCount = { ...course }
      delete withoutCount.lessonCount
      return withoutCount
    }
  }))
}

export async function getCourses(): Promise<CourseList> {
  const response = await apiClient.get<ApiEnvelope<{ items: RawCourse[]; total: number }>>('/api/v1/courses')
  const payload = unwrap(response)
  return { items: await enrichLessonCounts(payload.items.map(mapCourse), getCourse), total: payload.total }
}

export async function getCourse(id: string): Promise<Course> {
  const response = await apiClient.get<ApiEnvelope<RawCourseDetail>>(`/api/v1/courses/${id}`)
  const payload = unwrap(response)
  const chapters = payload.chapters.map((chapter) => ({
    id: chapter.id,
    title: chapter.title,
    lessons: (chapter.lessons ?? []).map((lesson) => ({
      id: lesson.id,
      title: lesson.title,
      contentType: lesson.content_type,
      contentUrl: lesson.content_url,
      resourceId: lesson.resource_id,
      duration: Math.ceil((lesson.duration_seconds ?? 0) / 60),
    })),
  }))
  return {
    ...mapCourse(payload.course),
    chapters,
    lessonCount: countLessons(chapters),
    duration: chapters.reduce(
      (total, chapter) => total + chapter.lessons.reduce((sum, lesson) => sum + lesson.duration, 0),
      0,
    ),
  }
}

export async function getResourceFile(id: string): Promise<Blob> {
  const response = await apiClient.get<Blob>(`/api/v1/resources/${id}/file`, { responseType: 'blob' })
  return response.data
}
