import { useEffect, useState } from 'react'
import { Button, Collapse, DotLoading, ErrorBlock, NavBar } from 'antd-mobile'
import type { CSSProperties } from 'react'
import { CheckCircleFill, ClockCircleOutline, PlayOutline } from 'antd-mobile-icons'
import { useNavigate, useParams } from 'react-router-dom'
import { getCourse } from '../api/course'
import type { Course } from '../types/course'

export function CourseDetailPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const [course, setCourse] = useState<Course | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    getCourse(id).then(setCourse).catch(() => setCourse(null)).finally(() => setLoading(false))
  }, [id])

  if (loading) {
    return <div className="loading-state"><DotLoading color="primary" /> 正在加载课程</div>
  }
  if (!course) {
    return <ErrorBlock status="empty" title="课程不可访问" description="课程不存在或尚未发布" />
  }

  const chapters = course.chapters ?? []

  return (
    <div className="detail-page">
      <div className="detail-hero" style={{ background: course.cover }}>
        <NavBar onBack={() => navigate(-1)}>课程详情</NavBar>
        <div className="detail-title">
          <span>{course.category}</span>
          <h1>{course.title}</h1>
          <p>{course.instructor}</p>
        </div>
      </div>
      <section className="detail-summary glass-card reveal">
        <div className="detail-stats">
          <span>{course.lessonCount} 课时</span>
          <span><ClockCircleOutline /> {course.duration} 分钟</span>
          <span>随时学习</span>
        </div>
        <p>{course.description}</p>
        <div className="detail-progress">
          <div><strong>学习进度</strong><span>{course.progress}%</span></div>
          <div className="progress-track"><div className="progress-fill" style={{ '--progress': `${course.progress}%` } as CSSProperties} /></div>
        </div>
      </section>
      <section className="chapter-section glass-card reveal">
        <div className="section-heading">
          <div><span>COURSE OUTLINE</span><h2>课程章节</h2></div>
          <small>{chapters.length} 章</small>
        </div>
        <Collapse accordion={false} defaultActiveKey={chapters[0] ? [chapters[0].id] : []}>
          {chapters.map((chapter) => (
            <Collapse.Panel key={chapter.id} title={chapter.title}>
              <div className="lesson-list">
                {chapter.lessons.map((lesson, index) => (
                  <div className="lesson-item" key={lesson.id} onClick={() => navigate(`/courses/${course.id}/lessons/${lesson.id}`)}>
                    <div className={lesson.completed ? 'lesson-icon done' : 'lesson-icon'}>
                      {lesson.completed ? <CheckCircleFill /> : <PlayOutline />}
                    </div>
                    <div><strong>{index + 1}. {lesson.title}</strong><span>{lesson.duration} 分钟</span></div>
                  </div>
                ))}
              </div>
            </Collapse.Panel>
          ))}
        </Collapse>
      </section>
      <div className="detail-action">
        <Button
          block
          color="primary"
          size="large"
          disabled={!chapters.some((chapter) => chapter.lessons.length)}
          onClick={() => {
            const lesson = chapters.flatMap((chapter) => chapter.lessons)[0]
            if (lesson) navigate(`/courses/${course.id}/lessons/${lesson.id}`)
          }}
        >
          <PlayOutline /> {course.progress ? '继续学习' : '开始学习'}
        </Button>
      </div>
    </div>
  )
}
