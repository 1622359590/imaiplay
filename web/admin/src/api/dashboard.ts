import client from './client'

export interface PlatformTenant {
  id: string
  name: string
  code: string
  status: number
  lifecycle_status?: string
  created_at: string
}

export interface TenantDashboard {
  scope: 'tenant'
  today_learning_user_count: number
  yesterday_learning_user_count: number
  today_learning_user_delta: number
  learner_count: number
  today_new_learner_count: number
  published_course_count: number
  course_count: number
  resource_category_count: number
  resource_count: number
  manager_count: number
  has_demo_data: boolean
  resource_type_counts: Record<'video' | 'image' | 'document' | 'attachment', number>
  today_learning_ranking: Array<{
    user_id: string
    display_name: string
    duration_seconds: number
  }>
}

export interface PlatformDashboard {
  scope: 'platform'
  tenant_count: number
  active_tenant_count: number
  learner_count: number
  course_count: number
  recent_tenants: PlatformTenant[]
}

export interface InstructorDashboard {
  scope: 'instructor'
  course_count: number
  published_course_count: number
  today_learning_user_count: number
  recent_courses: Array<{
    id: string
    title: string
    status: number
    updated_at: string
  }>
}

export type DashboardResponse = PlatformDashboard | TenantDashboard | InstructorDashboard

export const dashboardApi = {
  get: () => client.get<DashboardResponse>('/backend/v1/dashboard'),
}
