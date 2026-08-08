export interface CourseEditSource {
  title: string
  description?: string
  status: 0 | 1
  category_id?: string | null
  course_type: 'required' | 'optional'
}

export function normalizeCourseEditValues(course: CourseEditSource) {
  return {
    title: course.title,
    description: course.description,
    status: course.status,
    category_id: course.category_id || undefined,
    course_type: course.course_type,
  }
}

export function categoryIDForPayload(value: string | null | undefined): string | null {
  const normalized = value?.trim()
  return normalized || null
}
