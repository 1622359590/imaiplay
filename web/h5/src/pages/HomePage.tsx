import { useEffect, useState } from 'react'
import { DotLoading } from 'antd-mobile'
import { useNavigate } from 'react-router-dom'
import { logout } from '../api/auth'
import { getCourses } from '../api/course'
import { CourseCard } from '../components/CourseCard'
import type { Course } from '../types/course'
import { useTenantTheme } from '../context/TenantThemeContext'

export function HomePage() {
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()
  const theme = useTenantTheme()

  useEffect(() => {
    getCourses()
      .then((result) => setCourses(result.items))
      .catch(() => setCourses([]))
      .finally(() => setLoading(false))
  }, [])

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="home-page">
      <header className="learner-header">
        <div className="learner-brand">
          {theme.logo_url ? <img src={theme.logo_url} alt="租户 Logo" /> : <span>IM</span>}
          <strong>{theme.welcome_text || 'iMaiPlay'}</strong>
        </div>
        <button type="button" onClick={handleLogout}>退出</button>
      </header>
      <section className="course-heading">
        <div>
          <h1>我的课程</h1>
          <p>选择课程开始学习</p>
        </div>
        {!loading && <span>共 {courses.length} 门</span>}
      </section>
      {loading ? (
        <div className="loading-state"><DotLoading color="primary" /> 正在加载课程</div>
      ) : courses.length ? (
        <div className="course-list">
          {courses.map((course) => <CourseCard key={course.id} course={course} />)}
        </div>
      ) : (
        <div className="empty-state">暂无可学习课程</div>
      )}
    </div>
  )
}
