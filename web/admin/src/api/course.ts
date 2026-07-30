import client from './client'
import type { ListParams, PageResult } from './types'

export interface Lesson {
  id: string
  title: string
  content_type: 'video' | 'document' | 'text'
  resource_id?: string
  content_url?: string
  duration_seconds?: number
  sort_order?: number
}

export interface Chapter {
  id: string
  title: string
  sort_order?: number
  lessons?: Lesson[]
}

export interface Course {
  id: string
  title: string
  description?: string
  status: 0 | 1
  cover_image?: string
  student_count?: number
  chapters?: Chapter[]
  is_official?: boolean
  created_at?: string
}

export interface CourseInput {
  title: string
  description?: string
  status?: 0 | 1
  cover_image?: string
  is_official?: boolean
}

export const courseApi = {
  list: (params: ListParams) => client.get<PageResult<Course> | Course[]>('/backend/v1/courses', {
    params: {
      offset: Math.max(0, ((params.page ?? 1) - 1) * (params.page_size ?? 20)),
      limit: params.page_size ?? 20,
    },
  }),
  detail: async (id: string) => {
    const response = await client.get<{ course: Course; chapters: Chapter[] }>(`/backend/v1/courses/${id}/detail`)
    return { ...response, data: { ...response.data.course, chapters: response.data.chapters } }
  },
  create: (data: CourseInput) => {
    const { status: _status, ...createData } = data
    return client.post<Course>('/backend/v1/courses', createData)
  },
  update: (id: string, data: Partial<CourseInput>) => client.put<Course>(`/backend/v1/courses/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/courses/${id}`),
  createChapter: (courseId: string, data: { title: string }) =>
    client.post<Chapter>(`/backend/v1/courses/${courseId}/chapters`, data),
  updateChapter: (_courseId: string, chapterId: string, data: { title: string }) =>
    client.put<Chapter>(`/backend/v1/chapters/${chapterId}`, data),
  removeChapter: (_courseId: string, chapterId: string) =>
    client.delete(`/backend/v1/chapters/${chapterId}`),
  createLesson: (_courseId: string, chapterId: string, data: Omit<Lesson, 'id'>) =>
    client.post<Lesson>(`/backend/v1/chapters/${chapterId}/lessons`, data),
  updateLesson: (_courseId: string, _chapterId: string, lessonId: string, data: Omit<Lesson, 'id'>) =>
    client.put<Lesson>(`/backend/v1/lessons/${lessonId}`, data),
  removeLesson: (_courseId: string, _chapterId: string, lessonId: string) =>
    client.delete(`/backend/v1/lessons/${lessonId}`),
}
