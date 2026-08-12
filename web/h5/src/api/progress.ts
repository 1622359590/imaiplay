import { apiClient, unwrap, type ApiEnvelope } from './client'
import { getActivePortalCode, readPortalAccessToken, readPortalTenantCode } from './authSession'

export interface LessonProgress {
  progress_percent: number
  last_position_seconds: number
}

export interface LessonPlaybackState {
  position: number
  percent: number
  lastReported: number
}

export interface LessonRequestGate {
  begin: () => number
  isCurrent: (token: number) => boolean
  cancel: (token: number) => void
}

export interface MediaLifecycleRecord<Media, Controller> {
  loadedLessonId: string
  mediaElement: Media
  controller: Controller
}

export interface MediaLifecycleGate<Media, Controller> {
  setRouteLessonId: (lessonId: string) => void
  bind: (
    loadedLessonId: string | undefined,
    mediaElement: Media,
    controller: Controller,
  ) => MediaLifecycleRecord<Media, Controller> | undefined
  currentFor: (mediaElement: Media) => MediaLifecycleRecord<Media, Controller> | undefined
  isCurrent: (record: MediaLifecycleRecord<Media, Controller>) => boolean
  unbind: (record: MediaLifecycleRecord<Media, Controller>) => void
}

export interface PlaybackReportDecision {
  percent: number
  lastReported: number
  reported: boolean
}

export interface ProgressHeartbeat {
  watched_seconds_delta: number
  report_id: string
  session_id: string
}

function progressPayload(
  positionSeconds: number,
  progressPercent: number,
  heartbeat?: ProgressHeartbeat,
) {
  return {
    position_seconds: Math.max(0, Math.floor(positionSeconds)),
    progress_percent: Math.max(0, Math.min(100, Math.floor(progressPercent))),
    ...(heartbeat ? {
      watched_seconds_delta: heartbeat.watched_seconds_delta,
      report_id: heartbeat.report_id,
      session_id: heartbeat.session_id,
    } : {}),
  }
}

export function lessonPlaybackState(progress?: LessonProgress | null): LessonPlaybackState {
  return {
    position: Math.max(0, Math.floor(progress?.last_position_seconds ?? 0)),
    percent: Math.max(0, Math.min(100, Math.floor(progress?.progress_percent ?? 0))),
    lastReported: -1,
  }
}

export function shouldReportPlaybackProgress(
  lastReported: number,
  nextPercent: number,
  force: boolean,
): boolean {
  return force || lastReported < 0 || nextPercent >= lastReported + 5
}

export function createLessonRequestGate(): LessonRequestGate {
  let generation = 0
  return {
    begin: () => ++generation,
    isCurrent: (token) => token === generation,
    cancel: (token) => {
      if (token === generation) generation++
    },
  }
}

export function createMediaLifecycleGate<Media, Controller>(): MediaLifecycleGate<Media, Controller> {
  let routeLessonId = ''
  let current: MediaLifecycleRecord<Media, Controller> | undefined
  const isCurrent = (record: MediaLifecycleRecord<Media, Controller>) => (
    current === record && record.loadedLessonId === routeLessonId
  )

  return {
    setRouteLessonId: (lessonId) => { routeLessonId = lessonId },
    bind: (loadedLessonId, mediaElement, controller) => {
      if (!loadedLessonId || loadedLessonId !== routeLessonId) return undefined
      current = { loadedLessonId, mediaElement, controller }
      return current
    },
    currentFor: (mediaElement) => (
      current && current.mediaElement === mediaElement && isCurrent(current)
        ? current
        : undefined
    ),
    isCurrent,
    unbind: (record) => {
      if (current === record) current = undefined
    },
  }
}

export function reportPlaybackForMedia(input: {
  mediaLessonId: string | undefined
  routeLessonId: string
  currentTime: number
  duration: number
  lastReported: number
  force: boolean
  report: (lessonId: string, positionSeconds: number, percent: number) => void
}): PlaybackReportDecision | undefined {
  if (!input.mediaLessonId || input.mediaLessonId !== input.routeLessonId) return undefined
  if (!Number.isFinite(input.duration) || input.duration <= 0) return undefined
  const percent = Math.min(100, Math.floor((input.currentTime / input.duration) * 100))
  const reported = shouldReportPlaybackProgress(input.lastReported, percent, input.force)
  if (reported) input.report(input.mediaLessonId, input.currentTime, percent)
  return {
    percent,
    lastReported: reported ? percent : input.lastReported,
    reported,
  }
}

export async function getLessonProgress(lessonId: string): Promise<LessonProgress> {
  return unwrap(await apiClient.get<ApiEnvelope<LessonProgress>>(`/api/v1/lessons/${lessonId}/progress`))
}

export async function reportLessonProgress(
  lessonId: string,
  positionSeconds: number,
  progressPercent: number,
	heartbeat?: ProgressHeartbeat,
): Promise<LessonProgress> {
	return unwrap(await apiClient.post<ApiEnvelope<LessonProgress>>(
		`/api/v1/lessons/${lessonId}/progress`,
		progressPayload(positionSeconds, progressPercent, heartbeat),
	))
}

export async function reportLessonProgressOnPagehide(
  lessonId: string,
  positionSeconds: number,
  progressPercent: number,
  heartbeat: ProgressHeartbeat,
): Promise<void> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = readPortalAccessToken()
  const tenantCode = getActivePortalCode() ?? readPortalTenantCode()
  if (token) headers.Authorization = `Bearer ${token}`
  if (tenantCode) headers['X-Tenant-Code'] = tenantCode
  const response = await fetch(`/api/v1/lessons/${lessonId}/progress`, {
    method: 'POST',
    keepalive: true,
    credentials: 'same-origin',
    headers,
    body: JSON.stringify(progressPayload(positionSeconds, progressPercent, heartbeat)),
  })
  if (!response.ok) throw new Error(`Progress keepalive failed: ${response.status}`)
}
