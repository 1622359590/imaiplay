// @vitest-environment happy-dom

import { act, useEffect } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { MemoryRouter, Route, Routes, useNavigate, type NavigateFunction } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { getCourse, getCourses } from '../api/course'
import { getLearnerOverview } from '../api/learner'
import {
  getLessonProgress,
  reportLessonProgress,
  reportLessonProgressOnPagehide,
} from '../api/progress'
import type { Course } from '../types/course'
import { CourseDetailPage } from './CourseDetailPage'
import { HomePage } from './HomePage'
import { LessonPlayerPage } from './LessonPlayerPage'

(globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true

const lifecycleProbes = vi.hoisted(() => ({
  instances: [] as Array<{
    playing: ReturnType<typeof vi.fn>
    pause: ReturnType<typeof vi.fn>
    waiting: ReturnType<typeof vi.fn>
    seeked: ReturnType<typeof vi.fn>
    ended: ReturnType<typeof vi.fn>
    periodicFlush: ReturnType<typeof vi.fn>
    visibilityChanged: ReturnType<typeof vi.fn>
    pagehide: ReturnType<typeof vi.fn>
    pageshow: ReturnType<typeof vi.fn>
  }>,
}))

vi.mock('@imaiplay/shared/learning/watchHeartbeat', () => {
  class PlaybackLifecycleController {
    playing = vi.fn()
    pause = vi.fn()
    waiting = vi.fn()
    seeked = vi.fn()
    ended = vi.fn()
    periodicFlush = vi.fn()
    visibilityChanged = vi.fn()
    pagehide = vi.fn()
    pageshow = vi.fn()

    constructor() {
      lifecycleProbes.instances.push(this)
    }
  }

  return {
    PlaybackLifecycleController,
    restorePlaybackPosition: vi.fn(),
  }
})

vi.mock('../api/auth', () => ({ logout: vi.fn() }))

vi.mock('../api/course', async (importOriginal) => ({
  ...await importOriginal<typeof import('../api/course')>(),
  getCourse: vi.fn(),
  getCourses: vi.fn(),
  getResourceFile: vi.fn(),
}))

vi.mock('../api/learner', async (importOriginal) => ({
  ...await importOriginal<typeof import('../api/learner')>(),
  getLearnerOverview: vi.fn(),
}))

vi.mock('../api/progress', async (importOriginal) => ({
  ...await importOriginal<typeof import('../api/progress')>(),
  getLessonProgress: vi.fn(),
  reportLessonProgress: vi.fn(),
  reportLessonProgressOnPagehide: vi.fn(),
}))

vi.mock('../context/TenantThemeContext', () => ({
  useTenantTheme: () => ({
    logo_url: '',
    welcome_text: '测试学习空间',
    loginPath: '/login',
    routePath: (path: string) => path,
  }),
}))

let roots: Root[] = []

function fixtureCourse(overrides: Partial<Course> = {}): Course {
  return {
    id: 'course-1',
    title: '真实课程主体',
    description: '用于验证学习端降级行为',
    instructor: '讲师',
    progress: 87,
    duration: 30,
    category: '产品',
    courseType: 'required',
    chapters: [{
      id: 'chapter-1',
      title: '第一章',
      lessons: [
        { id: 'lesson-a', title: '课时 A', contentType: 'video', contentUrl: '/a.mp4', duration: 10 },
        { id: 'lesson-b', title: '课时 B', contentType: 'video', contentUrl: '/b.mp4', duration: 12 },
      ],
    }],
    ...overrides,
  }
}

async function renderAt(path: string, element: React.ReactNode) {
  const container = document.createElement('div')
  document.body.append(container)
  const root = createRoot(container)
  roots.push(root)
  await act(async () => {
    root.render(<MemoryRouter initialEntries={[path]}>{element}</MemoryRouter>)
  })
  return container
}

async function settle() {
  await act(async () => {
    await Promise.resolve()
    await new Promise((resolve) => window.setTimeout(resolve, 0))
  })
}

beforeEach(() => {
  lifecycleProbes.instances.length = 0
  vi.mocked(getCourse).mockReset()
  vi.mocked(getCourses).mockReset()
  vi.mocked(getLearnerOverview).mockReset()
  vi.mocked(getLessonProgress).mockReset()
  vi.mocked(reportLessonProgress).mockReset()
  vi.mocked(reportLessonProgressOnPagehide).mockReset()
})

afterEach(async () => {
  for (const root of roots) {
    await act(async () => root.unmount())
  }
  roots = []
  document.body.replaceChildren()
})

describe('learner overview degradation', () => {
  it('keeps Home courses visible at zero progress when overview fails', async () => {
    vi.mocked(getCourses).mockResolvedValue({ items: [fixtureCourse()], total: 1 })
    vi.mocked(getLearnerOverview).mockRejectedValue(new Error('overview unavailable'))

    const container = await renderAt('/', <HomePage />)
    await settle()

    expect(container.textContent).toContain('真实课程主体')
    expect(container.textContent).toContain('待开始')
    expect(container.textContent).not.toContain('87%')
  })

  it('shows the Home empty state when the required courses request fails', async () => {
    vi.mocked(getCourses).mockRejectedValue(new Error('courses unavailable'))
    vi.mocked(getLearnerOverview).mockResolvedValue({
      requiredCompleted: 0,
      requiredTotal: 0,
      todayLearningSeconds: 0,
      totalLearningSeconds: 0,
      courses: [],
    })

    const container = await renderAt('/', <HomePage />)
    await settle()

    expect(container.textContent).toContain('暂无可学习课程')
  })

  it('keeps Course Detail visible at zero progress when overview fails', async () => {
    vi.mocked(getCourse).mockResolvedValue(fixtureCourse())
    vi.mocked(getLearnerOverview).mockRejectedValue(new Error('overview unavailable'))

    const container = await renderAt(
      '/courses/course-1',
      <Routes><Route path="/courses/:id" element={<CourseDetailPage />} /></Routes>,
    )
    await settle()

    expect(container.textContent).toContain('真实课程主体')
    expect(container.textContent).toContain('0%')
    expect(container.textContent).not.toContain('87%')
  })

  it('shows Course Detail unavailable when the required course request fails', async () => {
    vi.mocked(getCourse).mockRejectedValue(new Error('course unavailable'))
    vi.mocked(getLearnerOverview).mockResolvedValue({
      requiredCompleted: 0,
      requiredTotal: 0,
      todayLearningSeconds: 0,
      totalLearningSeconds: 0,
      courses: [],
    })

    const container = await renderAt(
      '/courses/course-1',
      <Routes><Route path="/courses/:id" element={<CourseDetailPage />} /></Routes>,
    )
    await settle()

    expect(container.textContent).toContain('课程不可访问')
  })
})

describe('LessonPlayerPage route identity', () => {
  it('ignores retained A media events and stale A requests after navigating to B', async () => {
    const course = fixtureCourse({ progress: 0 })
    vi.mocked(getCourse).mockResolvedValue(course)
    vi.mocked(getLessonProgress).mockResolvedValue({
      progress_percent: 0,
      last_position_seconds: 0,
    })
    let navigateFromTest: NavigateFunction | undefined

    function NavigationProbe() {
      const navigate = useNavigate()
      useEffect(() => { navigateFromTest = navigate }, [navigate])
      return null
    }

    const container = await renderAt(
      '/courses/course-1/lessons/lesson-a',
      <>
        <NavigationProbe />
        <Routes>
          <Route path="/courses/:courseId/lessons/:lessonId" element={<LessonPlayerPage />} />
        </Routes>
      </>,
    )
    await settle()
    expect(container.querySelector('.lesson-metadata h1')?.textContent).toBe('课时 A')
    const retainedAVideo = container.querySelector('video')
    expect(retainedAVideo).not.toBeNull()
    const reactPropsKey = Object.keys(retainedAVideo!).find((key) => key.startsWith('__reactProps$'))
    expect(reactPropsKey).toBeDefined()
    const retainedAHandlers = (retainedAVideo as HTMLVideoElement & Record<string, unknown>)[reactPropsKey!] as {
      onTimeUpdate: (event: { currentTarget: HTMLVideoElement }) => void
      onPause: (event: { currentTarget: HTMLVideoElement }) => void
      onPlaying: (event: { currentTarget: HTMLVideoElement }) => void
      onWaiting: (event: { currentTarget: HTMLVideoElement }) => void
      onSeeked: (event: { currentTarget: HTMLVideoElement }) => void
      onEnded: (event: { currentTarget: HTMLVideoElement }) => void
    }

    await act(async () => navigateFromTest?.('/courses/course-1/lessons/lesson-b'))
    await settle()
    expect(container.querySelector('.lesson-metadata h1')?.textContent).toBe('课时 B')
    const controllerB = lifecycleProbes.instances.at(-1)
    expect(controllerB).toBeDefined()
    for (const method of ['playing', 'pause', 'waiting', 'seeked', 'ended'] as const) {
      controllerB?.[method].mockClear()
    }
    vi.mocked(reportLessonProgress).mockClear()

    Object.defineProperties(retainedAVideo!, {
      currentTime: { configurable: true, value: 5 },
      duration: { configurable: true, value: 10 },
    })
    await act(async () => {
      const event = { currentTarget: retainedAVideo! }
      retainedAHandlers.onTimeUpdate(event)
      retainedAHandlers.onPause(event)
      retainedAHandlers.onPlaying(event)
      retainedAHandlers.onWaiting(event)
      retainedAHandlers.onSeeked(event)
      retainedAHandlers.onEnded(event)
    })

    expect(reportLessonProgress).not.toHaveBeenCalledWith(
      'lesson-b',
      expect.any(Number),
      expect.any(Number),
      expect.anything(),
    )
    expect(reportLessonProgress).not.toHaveBeenCalledWith(
      'lesson-b',
      expect.any(Number),
      expect.any(Number),
    )
    expect(controllerB?.playing).not.toHaveBeenCalled()
    expect(controllerB?.pause).not.toHaveBeenCalled()
    expect(controllerB?.waiting).not.toHaveBeenCalled()
    expect(controllerB?.seeked).not.toHaveBeenCalled()
    expect(controllerB?.ended).not.toHaveBeenCalled()

    let resolveLateA: ((value: Course) => void) | undefined
    const lateA = new Promise<Course>((resolve) => { resolveLateA = resolve })
    vi.mocked(getCourse).mockImplementationOnce(() => lateA)
    vi.mocked(getCourse).mockResolvedValueOnce(course)

    await act(async () => navigateFromTest?.('/courses/course-1/lessons/lesson-a'))
    expect(vi.mocked(getCourse).mock.calls.length).toBe(3)
    await act(async () => navigateFromTest?.('/courses/course-1/lessons/lesson-b'))
    await settle()
    expect(container.querySelector('.lesson-metadata h1')?.textContent).toBe('课时 B')

    await act(async () => resolveLateA?.(course))
    await settle()
    expect(container.querySelector('.lesson-metadata h1')?.textContent).toBe('课时 B')
  })
})
