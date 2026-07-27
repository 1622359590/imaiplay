import { RightOutline } from 'antd-mobile-icons'
import { ProgressBar } from 'antd-mobile'
import { useNavigate } from 'react-router-dom'
import type { Course } from '../types/course'

interface CourseCardProps {
  course: Course
}

export function CourseCard({ course }: CourseCardProps) {
  const navigate = useNavigate()

  return (
    <article className="course-card" onClick={() => navigate(`/courses/${course.id}`)}>
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
          <ProgressBar percent={course.progress} style={{ '--track-width': '5px' }} />
          <span>{course.progress ? `${course.progress}%` : '未开始'}</span>
          <RightOutline fontSize={12} />
        </div>
      </div>
    </article>
  )
}
