import client from './client'
import type { Course } from './course'

export interface OfficialCourseInput {
  title: string
  description?: string
  cover_image?: string
  status: 0 | 1
  category_id?: string | null
}

export const officialCourseApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    client.get<{ items: Course[]; total: number }>(
      '/backend/v1/official-courses',
      {
        params: {
          offset: Math.max(0, ((params?.page ?? 1) - 1) * (params?.page_size ?? 20)),
          limit: params?.page_size ?? 20,
        },
      },
    ).then((response) => response.data),
  create: (data: OfficialCourseInput) =>
    client.post<Course>('/backend/v1/official-courses', data)
      .then((response) => response.data),
  update: (id: string, data: OfficialCourseInput) =>
    client.put<Course>(`/backend/v1/courses/${id}`, data)
      .then((response) => response.data),
  detail: (id: string) =>
    client.get<{ course: Course }>(`/backend/v1/courses/${id}/detail`)
      .then((response) => response.data),
  remove: (id: string) => client.delete(`/backend/v1/courses/${id}`),
  enable: (id: string, enabled: boolean) =>
    client.put(`/backend/v1/official-courses/${id}/enabled`, { enabled })
      .then((response) => response.data),
}
