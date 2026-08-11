import { apiClient, unwrap, type ApiEnvelope } from './client'
import { getActivePortalCode, readPortalAccessToken, readPortalTenantCode } from './authSession'

export interface LessonProgress {
  progress_percent: number
  last_position_seconds: number
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
