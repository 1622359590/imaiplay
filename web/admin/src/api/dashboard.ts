import client from './client'

export interface DashboardStats {
  user_count: number
  course_count: number
  published_course_count: number
  today_new_user_count: number
  today_learning_user_count: number
  total_learning_seconds: number
  course_completion_rate: number
  platform?: PlatformStats
}

export interface PlatformStats {
  tenant_count: number
  active_tenant_count: number
  learner_count: number
  course_count: number
  recent_tenants: PlatformTenant[]
}

export interface PlatformTenant {
  id: string
  name: string
  code: string
  status: number
  lifecycle_status?: string
  created_at: string
}

export const dashboardApi = {
  get: () => client.get<DashboardStats>('/backend/v1/dashboard'),
}
