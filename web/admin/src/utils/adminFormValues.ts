export interface CourseEditSource {
  title: string
  description?: string
  status: 0 | 1
  category_id?: string | null
}

export function normalizeCourseEditValues(course: CourseEditSource) {
  return {
    title: course.title,
    description: course.description,
    status: course.status,
    category_id: course.category_id || undefined,
  }
}

export function categoryIDForPayload(value: string | null | undefined): string | null {
  const normalized = value?.trim()
  return normalized || null
}

export type AssignmentType = 'required' | 'optional'

export function normalizeEnrollmentEditValues(enrollment: {
  user_id: string
  assignment_type?: AssignmentType
}) {
  return {
    user_id: enrollment.user_id,
    assignment_type: enrollment.assignment_type || 'required' as AssignmentType,
  }
}
