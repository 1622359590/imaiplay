import { useEffect, useState } from 'react'
import { Button, Collapse, DotLoading, ErrorBlock, NavBar } from 'antd-mobile'
import { FileOutline, TextOutline, VideoOutline } from 'antd-mobile-icons'
import { lessonContentLabel } from '@imaiplay/shared/learning/lessonContent'
import { useNavigate, useParams } from 'react-router-dom'
import { countLessons, getCourse } from '../api/course'
import type { Course } from '../types/course'
import { useTenantTheme } from '../context/TenantThemeContext'
import { CourseMaterials } from '../components/CourseMaterials'

export function CourseDetailPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { routePath } = useTenantTheme()
  const [course, setCourse] = useState<Course | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    getCourse(id).then(setCourse).catch(() => setCourse(null)).finally(() => setLoading(false))
  }, [id])

  if (loading) return <div className="loading-state"><DotLoading color="primary" /> 正在加载课程</div>

  if (!course) {
    return (
      <div className="detail-error">
        <ErrorBlock status="empty" title="课程不可访问" description="课程不存在或尚未发布" />
        <Button onClick={() => navigate(routePath('/'), { replace: true })}>返回我的课程</Button>
      </div>
    )
  }

  const chapters = course.chapters ?? []
  const lessonCount = countLessons(chapters)

  return (
    <div className="detail-page">
      <NavBar onBack={() => navigate(routePath('/'))}>课程详情</NavBar>
      <main className="detail-content">
        <div className="detail-cover" style={{ background: course.cover }} />
        <section className="detail-summary">
		  <div className="course-card-tags">
			<span className="course-type-badge">{course.courseType === 'required' ? '必修课' : '选修课'}</span>
			<span className="course-category-badge">{course.category}</span>
			<span>{lessonCount} 课时</span>
		  </div>
          <h1>{course.title}</h1>
          <p>{course.description || '暂无课程简介'}</p>
        </section>
        <CourseMaterials materials={course.materials ?? []} />
        <section className="chapter-section">
          <h2>课程目录</h2>
          {lessonCount ? (
            <Collapse accordion={false} defaultActiveKey={chapters[0] ? [chapters[0].id] : []}>
              {chapters.map((chapter, chapterIndex) => (
                <Collapse.Panel key={chapter.id} title={`第 ${chapterIndex + 1} 章　${chapter.title}`}>
                  <div className="lesson-list">
                    {chapter.lessons.map((lesson, lessonIndex) => (
                      <button
                        type="button"
                        className="lesson-item"
                        key={lesson.id}
                        onClick={() => navigate(routePath(`/courses/${course.id}/lessons/${lesson.id}`))}
                      >
                        <span>{chapterIndex + 1}.{lessonIndex + 1}　{lesson.title}</span>
                        <small>
                          {lesson.contentType === 'video' ? <VideoOutline /> : lesson.contentType === 'document' ? <FileOutline /> : <TextOutline />}
                          {lesson.contentType === 'video' && lesson.duration > 0 ? `${lesson.duration} 分钟` : lessonContentLabel(lesson.contentType ?? 'text')}
                        </small>
                      </button>
                    ))}
                  </div>
                </Collapse.Panel>
              ))}
            </Collapse>
          ) : (
            <div className="empty-state">暂无可学习课时</div>
          )}
        </section>
      </main>
    </div>
  )
}
