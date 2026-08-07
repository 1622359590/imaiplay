import client from './client'
import type { ListParams, PageResult } from './types'
import { createUserImportFormData, type UserImportResult } from '../utils/userImport'

export type { UserImportError, UserImportResult } from '../utils/userImport'

export interface User {
  id: string
  tenant_id?: string
  tenant_name?: string
  tenant_code?: string
  name: string
  email: string
  phone?: string
  role: string
  status: number
  created_at?: string
}

export interface UserInput {
  name: string
  email: string
  phone?: string
  role: string
  status: User['status']
  password?: string
}

export const userApi = {
  list: (params: ListParams) => client.get<PageResult<User> | User[]>('/backend/v1/users', {
    params: {
      offset: Math.max(0, ((params.page ?? 1) - 1) * (params.page_size ?? 20)),
      limit: params.page_size ?? 20,
    },
  }),
  create: (data: UserInput) => {
    const { status: _status, ...createData } = data
    return client.post<User>('/backend/v1/users', createData)
  },
  import: (file: File) => client.post<UserImportResult>(
    '/backend/v1/users/import',
    createUserImportFormData(file),
  ),
  update: (id: string, data: UserInput) => client.put<User>(`/backend/v1/users/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/users/${id}`),
  resetTenantAdminPassword: (id: string, password: string) => client.put(`/backend/v1/users/${id}/password`, { password }),
}
