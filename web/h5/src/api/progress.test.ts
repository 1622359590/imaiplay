import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  apiClient: { post: vi.fn() },
  unwrap: vi.fn(),
}))

vi.mock('./authSession', () => ({
  getActivePortalCode: vi.fn(() => 'tenant-one'),
  readPortalAccessToken: vi.fn(() => 'access-token'),
  readPortalTenantCode: vi.fn(() => 'tenant-one'),
}))

import { apiClient, unwrap } from './client'
import {
  createMediaLifecycleGate,
  createLessonRequestGate,
  lessonPlaybackState,
  reportPlaybackForMedia,
  reportLessonProgress,
  reportLessonProgressOnPagehide,
  shouldReportPlaybackProgress,
} from './progress'

describe('H5 progress heartbeat', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.mocked(apiClient.post).mockResolvedValue({ data: {} })
    vi.mocked(unwrap).mockReturnValue({ progress_percent: 15, last_position_seconds: 15 })
  })

  it('sends the cross-device session and stable report identifiers', async () => {
    await reportLessonProgress('lesson-1', 15.8, 15.9, {
      watched_seconds_delta: 15,
      report_id: 'report-1',
      session_id: 'session-1',
    })

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/lessons/lesson-1/progress', {
      position_seconds: 15,
      progress_percent: 15,
      watched_seconds_delta: 15,
      report_id: 'report-1',
      session_id: 'session-1',
    })
  })

  it('uses a keepalive request during pagehide', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(null, { status: 200 }))

    await reportLessonProgressOnPagehide('lesson-1', 10, 20, {
      watched_seconds_delta: 10,
      report_id: 'report-2',
      session_id: 'session-1',
    })

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/lessons/lesson-1/progress', expect.objectContaining({
      method: 'POST',
      keepalive: true,
      headers: expect.objectContaining({
        Authorization: 'Bearer access-token',
        'X-Tenant-Code': 'tenant-one',
      }),
    }))
  })
})

describe('H5 lesson playback transitions', () => {
  it('clears historical progress and allows an early report when the next lesson has no history', () => {
    expect(lessonPlaybackState({ progress_percent: 74, last_position_seconds: 320 })).toEqual({
      position: 320,
      percent: 74,
      lastReported: -1,
    })

    const reset = lessonPlaybackState(null)
    expect(reset).toEqual({ position: 0, percent: 0, lastReported: -1 })
    expect(shouldReportPlaybackProgress(reset.lastReported, 1, false)).toBe(true)
  })

  it('rejects an old media event after the route has moved to another lesson', () => {
    const report = vi.fn()
    const decision = reportPlaybackForMedia({
      mediaLessonId: 'lesson-a',
      routeLessonId: 'lesson-b',
      currentTime: 12,
      duration: 100,
      lastReported: -1,
      force: true,
      report,
    })

    expect(decision).toBeUndefined()
    expect(report).not.toHaveBeenCalled()
  })

  it('lets only the latest lesson request commit when the older promise resolves last', async () => {
    const gate = createLessonRequestGate()
    const commits: string[] = []
    let resolveOld!: (value: string) => void
    let resolveCurrent!: (value: string) => void
    const oldRequest = new Promise<string>((resolve) => { resolveOld = resolve })
    const currentRequest = new Promise<string>((resolve) => { resolveCurrent = resolve })
    const oldToken = gate.begin()
    const currentToken = gate.begin()

    const oldCommit = oldRequest.then((value) => {
      if (gate.isCurrent(oldToken)) commits.push(value)
    })
    const currentCommit = currentRequest.then((value) => {
      if (gate.isCurrent(currentToken)) commits.push(value)
    })
    resolveCurrent('lesson-b')
    await currentCommit
    resolveOld('lesson-a')
    await oldCommit

    expect(commits).toEqual(['lesson-b'])
  })

  it('binds the exact video and controller after a resource URL resolves later', async () => {
    const gate = createMediaLifecycleGate<object, { id: string }>()
    const media = { id: 'resource-backed-video' }
    const controller = { id: 'controller-resource' }
    gate.setRouteLessonId('lesson-resource')

    expect(gate.currentFor(media)).toBeUndefined()
    await Promise.resolve('/api/v1/resources/resource-1/file')

    const record = gate.bind('lesson-resource', media, controller)
    expect(record).toEqual({
      loadedLessonId: 'lesson-resource',
      mediaElement: media,
      controller,
    })
    expect(gate.currentFor(media)).toBe(record)
    expect(gate.isCurrent(record!)).toBe(true)
  })

  it('rejects old nodes across A to B navigation and same-lesson controller rebuilds', () => {
    const gate = createMediaLifecycleGate<object, { id: string }>()
    const mediaA = { id: 'video-a' }
    const mediaB = { id: 'video-b' }
    const rebuiltMediaB = { id: 'video-b-rebuilt' }

    gate.setRouteLessonId('lesson-a')
    const recordA = gate.bind('lesson-a', mediaA, { id: 'controller-a' })!
    gate.setRouteLessonId('lesson-b')
    const recordB = gate.bind('lesson-b', mediaB, { id: 'controller-b' })!

    expect(gate.currentFor(mediaA)).toBeUndefined()
    expect(gate.isCurrent(recordA)).toBe(false)
    expect(gate.currentFor(mediaB)).toBe(recordB)

    const rebuiltRecordB = gate.bind('lesson-b', rebuiltMediaB, { id: 'controller-b-2' })!
    expect(gate.currentFor(mediaB)).toBeUndefined()
    expect(gate.isCurrent(recordB)).toBe(false)
    expect(gate.currentFor(rebuiltMediaB)).toBe(rebuiltRecordB)
  })
})
