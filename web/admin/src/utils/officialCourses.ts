import type { Course } from '../api/course'

export function updateOfficialCourseEnabled(
  courses: Course[],
  courseId: string,
  enabled: boolean,
): Course[] {
  return courses.map((course) => (
    course.id === courseId ? { ...course, enabled } : course
  ))
}
