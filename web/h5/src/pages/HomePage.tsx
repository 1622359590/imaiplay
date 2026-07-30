import { useEffect, useState } from 'react'
import { DotLoading } from 'antd-mobile'
import { BellOutline, RightOutline } from 'antd-mobile-icons'
import { getCourses } from '../api/course'
import { CourseCard } from '../components/CourseCard'
import type { Course } from '../types/course'
import { useTenantTheme } from '../context/TenantThemeContext'

export function HomePage() {
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(true)
  const theme = useTenantTheme()

  useEffect(() => {
    getCourses()
      .then((result) => setCourses(result.items))
      .catch(() => setCourses([]))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="home-page">
      <header className="home-header reveal">
        <div>
          <p>周一好，学习者</p>
          <h1>{theme.welcome_text || '今天也要持续成长'}</h1>
        </div>
        <button className="icon-button" aria-label="通知">
          <BellOutline />
          <i />
        </button>
      </header>
      <section className="hero-card">
        <div>
          <span className="eyebrow">本周学习计划</span>
          <h2>保持节奏，离目标更近一步</h2>
          <p>本周已学习 2.5 小时</p>
        </div>
        <div className="hero-score">
          <strong>72</strong>
          <span>%</span>
        </div>
      </section>
      <section className="quick-stats stagger-group reveal">
        <div><strong>6</strong><span>在学课程</span></div>
        <div><strong>18</strong><span>完成课时</span></div>
        <div><strong>3</strong><span>获得证书</span></div>
      </section>
      <section className="section-block">
        <div className="section-heading">
          <div><span>CONTINUE</span><h2>继续学习</h2></div>
          <button>全部课程 <RightOutline /></button>
        </div>
        {loading ? (
          <div className="loading-state"><DotLoading color="primary" /> 正在加载课程</div>
        ) : (
          <div className="course-list stagger-group">
            {courses.map((course) => <CourseCard key={course.id} course={course} />)}
          </div>
        )}
      </section>
    </div>
  )
}
