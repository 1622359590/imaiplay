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
	category?: { id: string; name: string } | null
	course_type: string
}

interface RawCourseDetail {
  course: RawCourse
  chapters: RawChapter[]
  materials?: Array<{
    id: string
    display_name: string
    resource: {
      resource_type: 'attachment'
      size_bytes: number
    }
  }>
}

function mapCourse(course: RawCourse): Course {
	if (course.course_type !== 'required' && course.course_type !== 'optional') {
		throw new Error('Invalid learner course type')
	}
  return {
    id: course.id,
    title: course.title,
    description: course.description ?? '',
    cover: courseCoverStyle(course.cover_image),
    instructor: '企业讲师',
    progress: 0,
    duration: 0,
	category: course.category?.name ?? '未分类',
	courseType: course.course_type,
    materials: [],
  }
}

export function courseCoverStyle(coverImage?: string): string | undefined {
  return coverImage ? `url("${coverImage}") center/cover` : undefined
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
    materials: (payload.materials ?? []).map((material) => ({
      id: material.id,
      displayName: material.display_name,
      sizeBytes: material.resource.size_bytes,
      resourceType: material.resource.resource_type,
    })),
  }
}

export async function downloadCourseMaterial(id: string): Promise<Blob> {
  const response = await apiClient.get<Blob>(
    `/api/v1/course-materials/${encodeURIComponent(id)}/download`,
    { responseType: 'blob' },
  )
  return response.data
}

export async function getResourceFile(id: string): Promise<string> {
  const response = await apiClient.get<ApiEnvelope<{ url: string }>>(
    `/api/v1/resources/${id}/playback-url`,
  )
  return unwrap(response).url
}
