import client from './client'

export interface Plan { id: string; name: string; storage_quota_bytes: number; max_users: number; max_courses: number; features: string; is_default: boolean; status: number }
export interface PlanInput { name: string; storage_quota_bytes: number; max_users: number; max_courses: number; features?: string; is_default?: boolean; status?: number }
export const planApi = {
  list: (offset = 0, limit = 20) => client.get<{ items: Plan[]; total: number }>('/backend/v1/plans', { params: { offset, limit } }),
  create: (data: PlanInput) => client.post<Plan>('/backend/v1/plans', data),
  update: (id: string, data: PlanInput) => client.put<Plan>(`/backend/v1/plans/${id}`, data),
  remove: (id: string) => client.delete(`/backend/v1/plans/${id}`),
  assign: (tenantId: string, planId: string) => client.put(`/backend/v1/tenant-plans/${tenantId}`, { plan_id: planId }),
  current: () => client.get<{ plan: Plan; used_bytes: number; quota_bytes: number }>('/backend/v1/plan/current'),
}
