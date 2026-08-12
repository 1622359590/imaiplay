import { useNavigate } from 'react-router-dom'
import { ProgressBar } from 'antd-mobile'
import { RightOutline } from 'antd-mobile-icons'
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
        <span className="course-card-topline">
          <span className="course-card-tags">
            <small className="course-type-badge">{course.courseType === 'required' ? '必修课' : '选修课'}</small>
            <small className="course-category-badge">{course.category}</small>
          </span>
          <span className="course-card-arrow"><RightOutline /></span>
        </span>
        <strong>{course.title}</strong>
        <span className="course-card-meta">
          <small>{course.lessonCount !== undefined ? `${course.lessonCount} 课时` : course.instructor}</small>
          <small>{course.progress > 0 ? `${course.progress}%` : '待开始'}</small>
        </span>
        <ProgressBar percent={course.progress} />
      </span>
    </button>
  )
}
