import client from './client'
import type { PageResult } from './types'
import type { AxiosProgressEvent } from 'axios'

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
  listPlatform: (params?: { offset?: number; limit?: number }) =>
    client.get<PageResult<Resource>>('/backend/v1/admin/resources', { params }),
  listAll: (params?: { offset?: number; limit?: number }) =>
    client.get<PageResult<Resource>>('/backend/v1/admin/resources', { params }),
  upload: (file: File, onProgress?: (percent: number) => void) => {
    const data = new FormData()
    data.append('file', file)
    return client.post<Resource>('/backend/v1/resources/upload', data, {
      timeout: 0,
      onUploadProgress: (event: AxiosProgressEvent) => {
        if (event.total) onProgress?.(Math.round((event.loaded / event.total) * 100))
      },
    })
  },
  uploadPlatform: (file: File, onProgress?: (percent: number) => void) => {
    const data = new FormData()
    data.append('file', file)
    return client.post<Resource>('/backend/v1/admin/resources/upload', data, {
      timeout: 0,
      onUploadProgress: (event: AxiosProgressEvent) => {
        if (event.total) onProgress?.(Math.round((event.loaded / event.total) * 100))
      },
    })
  },
  file: (id: string) =>
    client.get<Blob>(`/backend/v1/resources/${id}/file`, { responseType: 'blob' }),
  remove: (id: string) => client.delete(`/backend/v1/resources/${id}`),
  platformFile: (id: string) =>
    client.get<Blob>(`/backend/v1/admin/resources/${id}/file`, { responseType: 'blob' }),
  removePlatform: (id: string) => client.delete(`/backend/v1/admin/resources/${id}`),
}

export const resourceCategoryApi = {
  list: () => client.get<ResourceCategory[]>('/backend/v1/resource-categories'),
  create: (data: { name: string; parent_id?: string }) =>
    client.post<ResourceCategory>('/backend/v1/resource-categories', data),
  update: (id: string, data: { name: string; parent_id?: string }) =>
    client.put<ResourceCategory>(`/backend/v1/resource-categories/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/resource-categories/${id}`),
}
