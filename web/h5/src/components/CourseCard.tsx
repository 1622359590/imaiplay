import { useNavigate } from 'react-router-dom'
import { useTenantTheme } from '../context/TenantThemeContext'
import type { Course } from '../types/course'

interface CourseCardProps {
  course: Course
}

export function CourseCard({ course }: CourseCardProps) {
  const navigate = useNavigate()
  const { routePath } = useTenantTheme()

  return (
    <button
      type="button"
      className="course-card"
      aria-label={`查看课程：${course.title}`}
      onClick={() => navigate(routePath(`/courses/${course.id}`))}
    >
      <span className="course-cover" style={{ background: course.cover }} />
      <span className="course-card-body">
        <strong>{course.title}</strong>
        {course.lessonCount !== undefined && <small>{course.lessonCount} 课时</small>}
      </span>
    </button>
  )
}
