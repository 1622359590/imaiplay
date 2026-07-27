import { useEffect, useRef, useState } from 'react'
import { ErrorBlock, NavBar, ProgressBar } from 'antd-mobile'
import { FileOutline } from 'antd-mobile-icons'
import { useNavigate, useParams } from 'react-router-dom'
import { getCourse } from '../api/course'
import { getLessonProgress, reportLessonProgress } from '../api/progress'
import type { Lesson } from '../types/course'

export function LessonPlayerPage() {
  const { courseId = '', lessonId = '' } = useParams()
  const navigate = useNavigate()
  const lastReported = useRef(-1)
  const [lesson, setLesson] = useState<Lesson>()
  const [position, setPosition] = useState(0)
  const [percent, setPercent] = useState(0)

  useEffect(() => {
    Promise.all([getCourse(courseId), getLessonProgress(lessonId).catch(() => null)]).then(([course, progress]) => {
      setLesson(course.chapters?.flatMap((chapter) => chapter.lessons).find((item) => item.id === lessonId))
      if (progress) {
        setPosition(progress.last_position_seconds)
        setPercent(progress.progress_percent)
      }
    })
  }, [courseId, lessonId])

  const report = (video: HTMLVideoElement, force = false) => {
    if (!Number.isFinite(video.duration) || video.duration <= 0) return
    const next = Math.min(100, Math.floor((video.currentTime / video.duration) * 100))
    setPercent(next)
    if (!force && next < lastReported.current + 5) return
    lastReported.current = next
    void reportLessonProgress(lessonId, video.currentTime, next)
  }

  return (
    <div className="player-page">
      <NavBar onBack={() => navigate(-1)}>{lesson?.title || '课时学习'}</NavBar>
      {!lesson ? (
        <ErrorBlock status="empty" title="课时不可访问" />
      ) : lesson.contentType === 'video' && lesson.contentUrl ? (
        <video
          className="mobile-video"
          controls
          playsInline
          src={lesson.contentUrl}
          onLoadedMetadata={(event) => { event.currentTarget.currentTime = position }}
          onTimeUpdate={(event) => report(event.currentTarget)}
          onPause={(event) => report(event.currentTarget, true)}
          onEnded={(event) => {
            setPercent(100)
            void reportLessonProgress(lessonId, event.currentTarget.duration, 100)
          }}
        />
      ) : lesson.contentUrl ? (
        <a className="mobile-document" href={lesson.contentUrl} target="_blank" rel="noreferrer"><FileOutline />打开课时资料</a>
      ) : (
        <ErrorBlock status="empty" title="该课时尚未配置学习资源" />
      )}
      <section className="mobile-player-progress">
        <div><strong>学习进度</strong><span>{percent}%</span></div>
        <ProgressBar percent={percent} />
      </section>
    </div>
  )
}
