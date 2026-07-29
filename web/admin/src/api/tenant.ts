import client from './client'
import type { ListParams, PageResult } from './types'

export interface Tenant {
  id: string
  name: string
  code: string
  status: number
  plan_id?: string
  lifecycle_status?: 'trial' | 'active' | 'suspended' | 'deleted'
  trial_ends_at?: string
  created_at?: string
  custom_domain?: string
}

export type TenantInput = Pick<Tenant, 'name' | 'code' | 'status'>

export interface RegisterTenantPayload {
  organization_name: string
  admin_email: string
	admin_name: string
	phone?: string
  password: string
}

export interface RegisterTenantResponse {
  tenant: { id: string; code: string; name: string }
  user: { id: string; email: string; name: string; role: string }
  token: string
}

export const tenantApi = {
  register: (data: RegisterTenantPayload) =>
    client.post<RegisterTenantResponse>('/api/v1/tenants/register', data),
  clearDemoData: () => client.delete('/backend/v1/tenants/demo-data'),
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
  update: (id: string, data: TenantInput & { lifecycle_status?: string; trial_ends_at?: string; custom_domain?: string }) => client.put<Tenant>(`/backend/v1/tenants/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/tenants/${id}`),
}
