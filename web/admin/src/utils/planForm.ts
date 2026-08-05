import type { Plan, PlanInput } from '../api/plan'

export interface PlanFormValues {
  name: string
  storage_quota_mb: number
  max_users: number
  max_courses: number
  features?: string
  is_default?: boolean
  status: 0 | 1
}

export function createPlanFormValues(plan?: Plan): PlanFormValues {
  if (!plan) {
    return {
      name: '',
      storage_quota_mb: 1024,
      max_users: 0,
      max_courses: 0,
      features: '{}',
      is_default: false,
      status: 1,
    }
  }
  return {
    name: plan.name,
    storage_quota_mb: plan.storage_quota_bytes / 1024 ** 2,
    max_users: plan.max_users,
    max_courses: plan.max_courses,
    features: plan.features,
    is_default: plan.is_default,
    status: plan.status === 1 ? 1 : 0,
  }
}

export function buildPlanInput(values: PlanFormValues): PlanInput {
  const { storage_quota_mb, ...rest } = values
  return {
    ...rest,
    storage_quota_bytes: storage_quota_mb * 1024 ** 2,
  }
}
