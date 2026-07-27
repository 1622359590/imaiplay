import client from './client'
import type { ListParams, PageResult } from './types'

export interface Tenant {
  id: string
  name: string
  code: string
  status: number
  created_at?: string
}

export type TenantInput = Pick<Tenant, 'name' | 'code' | 'status'>

export const tenantApi = {
  list: (params: ListParams) => client.get<PageResult<Tenant> | Tenant[]>('/backend/v1/tenants', {
    params: {
      offset: Math.max(0, ((params.page ?? 1) - 1) * (params.page_size ?? 20)),
      limit: params.page_size ?? 20,
    },
  }),
  create: (data: TenantInput) => {
    const { status: _status, ...createData } = data
    return client.post<Tenant>('/backend/v1/tenants', createData)
  },
  update: (id: string, data: TenantInput) => client.put<Tenant>(`/backend/v1/tenants/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/tenants/${id}`),
}
