import client from './client'
import type { Course } from './course'
export const officialCourseApi = {
  list: (params?: { page?: number; page_size?: number }) => client.get<{ items: Course[]; total: number }>('/backend/v1/official-courses', { params }).then((response) => response.data),
  create: (data: { title: string; description?: string; cover_image?: string }) => client.post('/backend/v1/official-courses', data).then((response) => response.data),
  enable: (id: string, enabled: boolean) => client.put(`/backend/v1/official-courses/${id}/enabled`, { enabled }).then((response) => response.data),
}
