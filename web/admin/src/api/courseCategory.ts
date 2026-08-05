import client from './client'

export interface CourseCategory {
  id: string
  tenant_id: string
  name: string
  sort_order: number
  status: 0 | 1
  created_at: string
  updated_at: string
}

export interface CourseCategoryInput {
  name: string
  sort_order: number
  status: 0 | 1
}

const collectionPath = (platform: boolean) => platform
  ? '/backend/v1/admin/course-categories'
  : '/backend/v1/course-categories'

export const courseCategoryApi = {
  list: (platform = false) => client.get<CourseCategory[]>(collectionPath(platform)),
  create: (data: CourseCategoryInput, platform = false) => client.post<CourseCategory>(collectionPath(platform), data),
  update: (id: string, data: CourseCategoryInput, platform = false) => client.put<CourseCategory>(`${collectionPath(platform)}/${id}`, data),
  remove: (id: string, platform = false) => client.delete(`${collectionPath(platform)}/${id}`),
}
