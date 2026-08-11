import { useEffect, useRef, useState } from 'react'
import { DotLoading, ErrorBlock, NavBar, ProgressBar } from 'antd-mobile'
import { FileOutline } from 'antd-mobile-icons'
import { resolveLessonContent } from '@imaiplay/shared/learning/lessonContent'
import { PlaybackLifecycleController, restorePlaybackPosition } from '@imaiplay/shared/learning/watchHeartbeat'
import { useNavigate, useParams } from 'react-router-dom'
import { getCourse, getResourceFile } from '../api/course'
import { getLessonProgress, reportLessonProgress, reportLessonProgressOnPagehide } from '../api/progress'
import type { Lesson } from '../types/course'

export function LessonPlayerPage() {
  const { courseId = '', lessonId = '' } = useParams()
  const navigate = useNavigate()
  const lastReported = useRef(-1)
  const videoRef = useRef<HTMLVideoElement>(null)
  const lifecycleRef = useRef<PlaybackLifecycleController | null>(null)
  const [lesson, setLesson] = useState<Lesson>()
  const [position, setPosition] = useState(0)
  const [percent, setPercent] = useState(0)
  const [resourceURL, setResourceURL] = useState<string>()
  const [loading, setLoading] = useState(true)
  const [resourceLoading, setResourceLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    Promise.all([getCourse(courseId), getLessonProgress(lessonId).catch(() => null)]).then(([course, progress]) => {
      setLesson(course.chapters?.flatMap((chapter) => chapter.lessons).find((item) => item.id === lessonId))
      if (progress) {
        setPosition(progress.last_position_seconds)
        setPercent(progress.progress_percent)
      }
    }).finally(() => setLoading(false))
  }, [courseId, lessonId])

  useEffect(() => {
    let active = true
    setResourceURL(undefined)
    setResourceLoading(Boolean(lesson?.resourceId))
    if (!lesson?.resourceId) return
    void getResourceFile(lesson.resourceId).then((url) => {
      if (active) setResourceURL(url)
    }).catch(() => {
      if (active) setResourceURL(undefined)
    }).finally(() => {
      if (active) setResourceLoading(false)
    })
    return () => { active = false }
  }, [lesson])

  useEffect(() => {
    const controller = new PlaybackLifecycleController({
      now: () => performance.now(),
      read: () => ({
        positionSeconds: videoRef.current?.currentTime ?? 0,
        durationSeconds: videoRef.current?.duration ?? lesson?.duration ?? 0,
      }),
      report: async (item) => {
        setPercent(item.progressPercent)
        await reportLessonProgress(lessonId, item.positionSeconds, item.progressPercent, item.heartbeat)
      },
      terminalReport: (item) => item.heartbeat
        ? reportLessonProgressOnPagehide(lessonId, item.positionSeconds, item.progressPercent, item.heartbeat)
        : Promise.resolve(),
    })
    lifecycleRef.current = controller
    const periodic = window.setInterval(() => void controller.periodicFlush(), 15_000)
    const visibilityChanged = () => void controller.visibilityChanged(document.visibilityState === 'visible')
    const pagehide = () => void controller.pagehide()
    const pageshow = () => void controller.pageshow(Boolean(
      videoRef.current && !videoRef.current.paused && !videoRef.current.ended,
    ))
    document.addEventListener('visibilitychange', visibilityChanged)
    window.addEventListener('pagehide', pagehide)
    window.addEventListener('pageshow', pageshow)
    return () => {
      window.clearInterval(periodic)
      document.removeEventListener('visibilitychange', visibilityChanged)
      window.removeEventListener('pagehide', pagehide)
      window.removeEventListener('pageshow', pageshow)
      void controller.pagehide()
      if (lifecycleRef.current === controller) lifecycleRef.current = null
    }
  }, [lesson?.duration, lessonId])

  const report = (video: HTMLVideoElement, force = false) => {
    if (!Number.isFinite(video.duration) || video.duration <= 0) return
    const next = Math.min(100, Math.floor((video.currentTime / video.duration) * 100))
    setPercent(next)
    if (!force && next < lastReported.current + 5) return
    lastReported.current = next
    void reportLessonProgress(lessonId, video.currentTime, next)
  }

  const content = lesson
    ? resolveLessonContent(lesson.contentType ?? 'text', lesson.contentUrl, resourceURL)
    : { kind: 'empty' as const }

  return (
    <div className="player-page">
      <NavBar onBack={() => navigate(-1)}>{lesson?.title || '课时学习'}</NavBar>
      {loading ? (
        <div className="mobile-resource-loading"><DotLoading color="primary" /> 正在加载课时</div>
      ) : !lesson ? (
        <ErrorBlock status="empty" title="课时不可访问" />
      ) : resourceLoading ? (
        <div className="mobile-resource-loading"><DotLoading color="primary" /> 正在加载课时内容</div>
      ) : content.kind === 'video' ? (
        <video
          ref={videoRef}
          className="mobile-video"
          controls
          playsInline
          src={content.source}
          onLoadedMetadata={(event) => restorePlaybackPosition(event.currentTarget, position)}
          onTimeUpdate={(event) => report(event.currentTarget)}
          onPlaying={() => lifecycleRef.current?.playing()}
          onPause={(event) => { report(event.currentTarget, true); void lifecycleRef.current?.pause() }}
          onWaiting={() => void lifecycleRef.current?.waiting()}
          onSeeked={() => void lifecycleRef.current?.seeked()}
          onEnded={(event) => {
            setPercent(100)
            void lifecycleRef.current?.ended()
          }}
        />
      ) : content.kind === 'document' ? (
        <a className="mobile-document" href={content.source} target="_blank" rel="noreferrer"><FileOutline />打开 PDF 文档</a>
      ) : content.kind === 'text' ? (
        <article className="mobile-text-lesson">{content.body}</article>
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
