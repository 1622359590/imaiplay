export interface CourseEditSource {
  title: string
  description?: string
  status: 0 | 1
  category_id?: string
}

export function normalizeCourseEditValues(course: CourseEditSource) {
  return {
    title: course.title,
    description: course.description,
    status: course.status,
    category_id: course.category_id,
  }
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
