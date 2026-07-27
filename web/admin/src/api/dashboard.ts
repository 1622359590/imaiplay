import client from './client'

export interface DashboardStats {
  user_count: number
  course_count: number
  published_course_count: number
  today_new_user_count: number
  today_learning_user_count: number
  total_learning_seconds: number
  course_completion_rate: number
}

export const dashboardApi = {
  get: () => client.get<DashboardStats>('/backend/v1/dashboard'),
}
