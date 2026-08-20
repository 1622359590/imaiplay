import { useEffect, useState } from 'react'
import { Button, DotLoading, ProgressBar } from 'antd-mobile'
import { PlayOutline } from 'antd-mobile-icons'
import { useNavigate } from 'react-router-dom'
import { logout } from '../api/auth'
import { getCourses } from '../api/course'
import { getLearnerOverview, loadCoursesWithOptionalOverview, type LearnerCourseView } from '../api/learner'
import { CourseCard } from '../components/CourseCard'
import { LearnerMotivationPrompt } from '../components/LearnerMotivationPrompt'
import { useTenantTheme } from '../context/TenantThemeContext'

export function HomePage() {
  const [courses, setCourses] = useState<LearnerCourseView[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [filter, setFilter] = useState<'all' | 'required' | 'optional'>('all')
  const navigate = useNavigate()
  const theme = useTenantTheme()

  useEffect(() => {
    let active = true
    loadCoursesWithOptionalOverview(async () => (await getCourses()).items, getLearnerOverview)
      .then((result) => {
        if (active) {
          setCourses(result)
          setLoadError(false)
        }
      })
      .catch(() => {
        if (active) {
          setCourses([])
          setLoadError(true)
        }
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => { active = false }
  }, [])

  const handleLogout = () => {
    logout()
    navigate(theme.loginPath, { replace: true })
  }

  const continueCourse = courses.find((course) => course.recentLesson && course.progress < 100)
  const continueLesson = continueCourse?.recentLesson
  const visibleCourses = filter === 'all'
    ? courses
    : courses.filter((course) => course.courseType === filter)
  const requiredCount = courses.filter((course) => course.courseType === 'required').length
  const completedCount = courses.filter((course) => course.progress >= 100).length

  return (
    <div className="home-page">
      <LearnerMotivationPrompt enabled={!loading && !loadError} />
      <header className="learner-header">
        <div className="learner-brand">
          {theme.logo_url ? <img src={theme.logo_url} alt="租户 Logo" /> : <span>IM</span>}
          <strong>{theme.welcome_text || 'iMaiPlay'}</strong>
        </div>
        <button className="header-action" type="button" onClick={handleLogout}>退出</button>
      </header>

      <section className="home-intro" aria-labelledby="home-title">
        <span className="section-eyebrow">LEARNING SPACE</span>
        <h1 id="home-title">今天也向前一步</h1>
        <p>{theme.welcome_text || '选择一门课程，继续你的学习旅程'}</p>
      </section>

      {loading ? (
        <div className="loading-state"><DotLoading color="primary" /> 正在加载课程</div>
      ) : loadError ? (
        <div className="empty-state home-error-state" role="alert">课程加载失败，请稍后重试</div>
      ) : courses.length ? (
        <>
          {continueCourse && continueLesson && (
            <section className="continue-card" aria-labelledby="continue-title">
              <div className="continue-copy">
                <span className="section-eyebrow">继续学习</span>
                <h2 id="continue-title">{continueCourse.title}</h2>
                <p>已完成 {continueCourse.progress}%</p>
                <ProgressBar percent={continueCourse.progress} />
              </div>
              <Button
                className="continue-action"
                color="primary"
                aria-label={`继续学习：${continueCourse.title}`}
                onClick={() => navigate(theme.routePath(
                  `/courses/${continueCourse.id}/lessons/${continueLesson.id}`,
                ))}
              >
                <PlayOutline />
              </Button>
            </section>
          )}

          <section className="learning-summary" aria-label="学习概览">
            <div><strong>{courses.length}</strong><span>全部课程</span></div>
            <div><strong>{requiredCount}</strong><span>必修课程</span></div>
            <div><strong>{completedCount}</strong><span>已学完</span></div>
          </section>

          <section className="course-catalog" aria-labelledby="course-catalog-title">
            <div className="course-heading">
              <div>
                <span className="section-eyebrow">COURSES</span>
                <h2 id="course-catalog-title">我的课程</h2>
              </div>
              <span>共 {visibleCourses.length} 门</span>
            </div>
            <div className="course-filters" aria-label="课程筛选">
              {([
                ['all', '全部'],
                ['required', '必修'],
                ['optional', '选修'],
              ] as const).map(([value, label]) => (
                <button
                  key={value}
                  type="button"
                  className={filter === value ? 'is-active' : ''}
                  aria-pressed={filter === value}
                  onClick={() => setFilter(value)}
                >
                  {label}
                </button>
              ))}
            </div>
            {visibleCourses.length ? (
              <div className="course-list">
                {visibleCourses.map((course) => <CourseCard key={course.id} course={course} />)}
              </div>
            ) : (
              <div className="empty-state">当前筛选下暂无课程</div>
            )}
          </section>
        </>
      ) : (
        <div className="empty-state">暂无可学习课程</div>
      )}
    </div>
  )
}
