import client from './client'
import type { PageResult } from './types'

export interface Resource {
  id: string
  category_id?: string
  name: string
  resource_type: 'image' | 'video' | 'document'
  url: string
  size_bytes: number
  created_at: string
}

export interface ResourceCategory {
  id: string
  name: string
  parent_id?: string
}

export const resourceApi = {
  list: (offset = 0, limit = 100) =>
    client.get<PageResult<Resource>>('/backend/v1/resources', { params: { offset, limit } }),
  upload: (file: File) => {
    const data = new FormData()
    data.append('file', file)
    return client.post<Resource>('/backend/v1/resources/upload', data, { timeout: 0 })
  },
  remove: (id: string) => client.delete(`/backend/v1/resources/${id}`),
}

export const resourceCategoryApi = {
  list: () => client.get<ResourceCategory[]>('/backend/v1/resource-categories'),
  create: (data: { name: string; parent_id?: string }) =>
    client.post<ResourceCategory>('/backend/v1/resource-categories', data),
  update: (id: string, data: { name: string; parent_id?: string }) =>
    client.put<ResourceCategory>(`/backend/v1/resource-categories/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/resource-categories/${id}`),
}
