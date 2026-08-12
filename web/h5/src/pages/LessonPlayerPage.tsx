import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { DotLoading, ErrorBlock, NavBar, ProgressBar } from 'antd-mobile'
import { FileOutline, PlayOutline, TextOutline, VideoOutline } from 'antd-mobile-icons'
import { lessonContentLabel, resolveLessonContent } from '@imaiplay/shared/learning/lessonContent'
import { PlaybackLifecycleController, restorePlaybackPosition } from '@imaiplay/shared/learning/watchHeartbeat'
import { useNavigate, useParams } from 'react-router-dom'
import { getCourse, getResourceFile } from '../api/course'
import {
  createLessonRequestGate,
  createMediaLifecycleGate,
  getLessonProgress,
  lessonPlaybackState,
  reportPlaybackForMedia,
  reportLessonProgress,
  reportLessonProgressOnPagehide,
} from '../api/progress'
import type { Course, Lesson } from '../types/course'
import { useTenantTheme } from '../context/TenantThemeContext'

export function LessonPlayerPage() {
  const { courseId = '', lessonId = '' } = useParams()
  const navigate = useNavigate()
  const { routePath } = useTenantTheme()
  const lastReported = useRef(-1)
  const lifecycleGateRef = useRef(createMediaLifecycleGate<HTMLVideoElement, PlaybackLifecycleController>())
  lifecycleGateRef.current.setRouteLessonId(lessonId)
  const requestGateRef = useRef(createLessonRequestGate())
  const [lesson, setLesson] = useState<Lesson>()
  const [loadedLessonId, setLoadedLessonId] = useState<string>()
  const [course, setCourse] = useState<Course>()
  const [position, setPosition] = useState(0)
  const [percent, setPercent] = useState(0)
  const [resourceURL, setResourceURL] = useState<string>()
  const [loading, setLoading] = useState(true)
  const [resourceLoading, setResourceLoading] = useState(false)
  const [mediaElement, setMediaElement] = useState<HTMLVideoElement | null>(null)
  const bindVideoRef = useCallback((node: HTMLVideoElement | null) => {
    setMediaElement(node)
  }, [])

  useEffect(() => {
    const requestToken = requestGateRef.current.begin()
    const reset = lessonPlaybackState(null)
    setLoading(true)
    setCourse(undefined)
    setLesson(undefined)
    setLoadedLessonId(undefined)
    setPosition(reset.position)
    setPercent(reset.percent)
    lastReported.current = reset.lastReported
    Promise.all([getCourse(courseId), getLessonProgress(lessonId).catch(() => null)]).then(([course, progress]) => {
      if (!requestGateRef.current.isCurrent(requestToken)) return
      const playback = lessonPlaybackState(progress)
      const loadedLesson = course.chapters?.flatMap((chapter) => chapter.lessons).find((item) => item.id === lessonId)
      setCourse(course)
      setLesson(loadedLesson)
      setLoadedLessonId(loadedLesson ? lessonId : undefined)
      setPosition(playback.position)
      setPercent(playback.percent)
      lastReported.current = playback.lastReported
    }).catch(() => {
      if (!requestGateRef.current.isCurrent(requestToken)) return
      setCourse(undefined)
      setLesson(undefined)
      setLoadedLessonId(undefined)
    }).finally(() => {
      if (requestGateRef.current.isCurrent(requestToken)) setLoading(false)
    })
    return () => { requestGateRef.current.cancel(requestToken) }
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

  useLayoutEffect(() => {
    const activeLessonId = loadedLessonId
    if (!activeLessonId || activeLessonId !== lessonId || !mediaElement) return
    const lifecycleGate = lifecycleGateRef.current
    let controllerRecord: ReturnType<typeof lifecycleGate.bind>
    const controller = new PlaybackLifecycleController({
      now: () => performance.now(),
      read: () => ({
        positionSeconds: mediaElement.currentTime ?? 0,
        durationSeconds: mediaElement.duration ?? lesson?.duration ?? 0,
      }),
      report: async (item) => {
        if (!controllerRecord || !lifecycleGate.isCurrent(controllerRecord)) return
        setPercent(item.progressPercent)
        await reportLessonProgress(activeLessonId, item.positionSeconds, item.progressPercent, item.heartbeat)
      },
      terminalReport: (item) => (
        item.heartbeat
        && controllerRecord
        && lifecycleGate.isCurrent(controllerRecord)
      )
        ? reportLessonProgressOnPagehide(activeLessonId, item.positionSeconds, item.progressPercent, item.heartbeat)
        : Promise.resolve(),
    })
    controllerRecord = lifecycleGate.bind(activeLessonId, mediaElement, controller)
    if (!controllerRecord) return
    const isActiveController = () => lifecycleGate.isCurrent(controllerRecord!)
    const periodic = window.setInterval(() => {
      if (isActiveController()) void controller.periodicFlush()
    }, 15_000)
    const visibilityChanged = () => {
      if (isActiveController()) void controller.visibilityChanged(document.visibilityState === 'visible')
    }
    const pagehide = () => {
      if (isActiveController()) void controller.pagehide()
    }
    const pageshow = () => {
      if (!isActiveController()) return
      void controller.pageshow(Boolean(
        mediaElement && !mediaElement.paused && !mediaElement.ended,
      ))
    }
    document.addEventListener('visibilitychange', visibilityChanged)
    window.addEventListener('pagehide', pagehide)
    window.addEventListener('pageshow', pageshow)
    return () => {
      window.clearInterval(periodic)
      document.removeEventListener('visibilitychange', visibilityChanged)
      window.removeEventListener('pagehide', pagehide)
      window.removeEventListener('pageshow', pageshow)
      void controller.pagehide()
      lifecycleGate.unbind(controllerRecord!)
    }
  }, [lesson?.duration, lessonId, loadedLessonId, mediaElement])

  const boundControllerFor = (video: HTMLVideoElement) => lifecycleGateRef.current.currentFor(video)

  const reportBoundMedia = (video: HTMLVideoElement, force = false) => {
    const boundControllerRecord = boundControllerFor(video)
    if (!boundControllerRecord) return
    const decision = reportPlaybackForMedia({
      mediaLessonId: boundControllerRecord.loadedLessonId,
      routeLessonId: lessonId,
      currentTime: video.currentTime,
      duration: video.duration,
      lastReported: lastReported.current,
      force,
      report: (activeLessonId, positionSeconds, progressPercent) => {
        void reportLessonProgress(activeLessonId, positionSeconds, progressPercent)
      },
    })
    if (!decision) return
    setPercent(decision.percent)
    lastReported.current = decision.lastReported
  }

  const withMediaLifecycle = (
    video: HTMLVideoElement,
    action: (controller: PlaybackLifecycleController) => void,
  ) => {
    const boundControllerRecord = boundControllerFor(video)
    if (!boundControllerRecord) return false
    action(boundControllerRecord.controller)
    return true
  }

  const content = lesson
    ? resolveLessonContent(lesson.contentType ?? 'text', lesson.contentUrl, resourceURL)
    : { kind: 'empty' as const }
  const chapter = course?.chapters?.find((item) => item.lessons.some((entry) => entry.id === lessonId))
  const outline = course?.chapters?.flatMap((item, chapterIndex) => item.lessons.map((entry, lessonIndex) => ({
    chapterTitle: item.title,
    label: `${chapterIndex + 1}.${lessonIndex + 1}`,
    lesson: entry,
  }))) ?? []

  return (
    <div className="player-page">
      <NavBar onBack={() => navigate(-1)}>{lesson?.title || '课时学习'}</NavBar>
      <section className="media-stage" aria-label="课时内容">
        {loading ? (
          <div className="mobile-resource-loading"><DotLoading color="primary" /> 正在加载课时</div>
        ) : !lesson ? (
          <ErrorBlock status="empty" title="课时不可访问" />
        ) : resourceLoading ? (
          <div className="mobile-resource-loading"><DotLoading color="primary" /> 正在加载课时内容</div>
        ) : content.kind === 'video' ? (
          <video
            key={`${courseId}:${loadedLessonId}:${content.source}`}
            ref={bindVideoRef}
            className="mobile-video"
            data-lesson-id={loadedLessonId}
            controls
            playsInline
            src={content.source}
            onLoadedMetadata={(event) => {
              if (!boundControllerFor(event.currentTarget)) return
              restorePlaybackPosition(event.currentTarget, position)
            }}
            onTimeUpdate={(event) => reportBoundMedia(event.currentTarget)}
            onPlaying={(event) => {
              withMediaLifecycle(event.currentTarget, (controller) => controller.playing())
            }}
            onPause={(event) => {
              withMediaLifecycle(event.currentTarget, (controller) => {
                reportBoundMedia(event.currentTarget, true)
                void controller.pause()
              })
            }}
            onWaiting={(event) => {
              withMediaLifecycle(event.currentTarget, (controller) => { void controller.waiting() })
            }}
            onSeeked={(event) => {
              withMediaLifecycle(event.currentTarget, (controller) => { void controller.seeked() })
            }}
            onEnded={(event) => {
              withMediaLifecycle(event.currentTarget, (controller) => {
                setPercent(100)
                void controller.ended()
              })
            }}
          />
        ) : content.kind === 'document' ? (
          <a className="mobile-document" href={content.source} target="_blank" rel="noreferrer"><span><FileOutline /></span>打开 PDF 文档</a>
        ) : content.kind === 'text' ? (
          <article className="mobile-text-lesson">{content.body}</article>
        ) : (
          <ErrorBlock status="empty" title="该课时尚未配置学习资源" />
        )}
      </section>
      {lesson && (
        <section className="lesson-metadata">
          <span className="section-eyebrow">{chapter?.title || course?.title || '课时学习'}</span>
          <h1>{lesson.title}</h1>
          <p>{lessonContentLabel(lesson.contentType ?? 'text')}{lesson.duration > 0 ? ` · ${lesson.duration} 分钟` : ''}</p>
        </section>
      )}
      <section className="mobile-player-progress" aria-label={`学习进度 ${percent}%`}>
        <div><strong>本课学习进度</strong><span>{percent}%</span></div>
        <ProgressBar percent={percent} />
      </section>
      {outline.length > 0 && (
        <section className="player-outline" aria-labelledby="player-outline-title">
          <div className="player-outline-heading">
            <div><span className="section-eyebrow">OUTLINE</span><h2 id="player-outline-title">课程目录</h2></div>
            <span>{outline.length} 课时</span>
          </div>
          <div className="player-outline-list">
            {outline.map((item) => {
              const isCurrent = item.lesson.id === lessonId
              const Icon = item.lesson.contentType === 'video'
                ? VideoOutline
                : item.lesson.contentType === 'document' ? FileOutline : TextOutline
              return (
                <button
                  type="button"
                  key={item.lesson.id}
                  className={`lesson-outline-item${isCurrent ? ' is-current' : ''}`}
                  aria-current={isCurrent ? 'step' : undefined}
                  onClick={() => navigate(routePath(`/courses/${courseId}/lessons/${item.lesson.id}`))}
                >
                  <span className="outline-icon"><Icon /></span>
                  <span className="outline-copy"><small>{item.label} · {item.chapterTitle}</small><strong>{item.lesson.title}</strong></span>
                  <span className="outline-play"><PlayOutline /></span>
                </button>
              )
            })}
          </div>
        </section>
      )}
    </div>
  )
}
