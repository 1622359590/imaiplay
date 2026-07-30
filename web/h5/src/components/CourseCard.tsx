import { RightOutline } from 'antd-mobile-icons'
import type { CSSProperties } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Course } from '../types/course'

interface CourseCardProps {
  course: Course
}

export function CourseCard({ course }: CourseCardProps) {
  const navigate = useNavigate()

  return (
    <article className="course-card glass-card tilt-card reveal" onClick={() => navigate(`/courses/${course.id}`)}>
      <div className="course-cover" style={{ background: course.cover }}>
        <span>{course.category}</span>
        <div className="cover-mark">IMAI</div>
      </div>
      <div className="course-card-body">
        <h3>{course.title}</h3>
        <p className="course-teacher">{course.instructor}</p>
        <div className="course-meta">
          <span>{course.lessonCount} 课时</span>
          <span>{course.duration} 分钟</span>
        </div>
        <div className="progress-row">
          <div className="progress-track"><div className="progress-fill" style={{ '--progress': `${course.progress}%` } as CSSProperties} /></div>
          <span>{course.progress ? `${course.progress}%` : '未开始'}</span>
          <RightOutline fontSize={12} />
        </div>
      </div>
    </article>
  )
}
