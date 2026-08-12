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
  lessonPlaybackState,
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
})
