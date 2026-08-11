import { apiClient } from './client';
import {
  readPortalAccessToken,
  readPortalTenantCode,
} from './authSession';
import { getActivePortalCode } from './portalSession';

export interface LessonProgress {
  lessonId: string;
  progressPercent: number;
  status: number;
  lastPositionSeconds: number;
}

export interface ProgressHeartbeat {
  watched_seconds_delta: number;
  report_id: string;
  session_id: string;
}

interface RawLessonProgress {
  lesson_id: string;
  progress_percent: number;
  status: number;
  last_position_seconds: number;
}

function mapLessonProgress(progress: RawLessonProgress): LessonProgress {
  return {
    lessonId: progress.lesson_id,
    progressPercent: progress.progress_percent,
    status: progress.status,
    lastPositionSeconds: progress.last_position_seconds,
  };
}

function validateHeartbeat(heartbeat: ProgressHeartbeat): void {
  if (
    !Number.isFinite(heartbeat.watched_seconds_delta)
    || !Number.isInteger(heartbeat.watched_seconds_delta)
    || heartbeat.watched_seconds_delta < 1
    || heartbeat.watched_seconds_delta > 60
    || typeof heartbeat.report_id !== 'string'
    || heartbeat.report_id.trim() === ''
    || typeof heartbeat.session_id !== 'string'
    || heartbeat.session_id.trim() === ''
  ) {
    throw new Error('Invalid progress heartbeat');
  }
}

function progressPayload(
  positionSeconds: number,
  progressPercent: number,
  heartbeat?: ProgressHeartbeat,
) {
  if (!Number.isFinite(positionSeconds) || !Number.isFinite(progressPercent)) {
    throw new Error('Invalid lesson progress');
  }
  if (heartbeat) validateHeartbeat(heartbeat);
  return {
    position_seconds: Math.max(0, Math.floor(positionSeconds)),
    progress_percent: Math.max(0, Math.min(100, Math.floor(progressPercent))),
    ...(heartbeat ? {
      watched_seconds_delta: heartbeat.watched_seconds_delta,
      report_id: heartbeat.report_id,
      session_id: heartbeat.session_id,
    } : {}),
  };
}

export async function getLessonProgress(lessonId: string): Promise<LessonProgress> {
  const response = await apiClient.get<RawLessonProgress>(`/api/v1/lessons/${lessonId}/progress`);
  return mapLessonProgress(response.data);
}

export async function reportLessonProgress(
  lessonId: string,
  positionSeconds: number,
  progressPercent: number,
  heartbeat?: ProgressHeartbeat,
): Promise<LessonProgress> {
  const response = await apiClient.post<RawLessonProgress>(
    `/api/v1/lessons/${lessonId}/progress`,
    progressPayload(positionSeconds, progressPercent, heartbeat),
  );
  return mapLessonProgress(response.data);
}

type TerminalProgressFetcher = (
  input: string,
  init: RequestInit,
) => Promise<unknown>;

interface TerminalProgressOptions {
  fetcher?: TerminalProgressFetcher;
  accessToken?: string | null;
  tenantCode?: string | null;
}

export async function reportLessonProgressOnPagehide(
  lessonId: string,
  positionSeconds: number,
  progressPercent: number,
  heartbeat: ProgressHeartbeat,
  options: TerminalProgressOptions = {},
): Promise<void> {
  const token = options.accessToken ?? readPortalAccessToken();
  const tenantCode = options.tenantCode ?? getActivePortalCode() ?? readPortalTenantCode();
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;
  if (tenantCode) headers['X-Tenant-Code'] = tenantCode;
  const fetcher = options.fetcher ?? ((input, init) => globalThis.fetch(input, init));
  const response = await fetcher(`/api/v1/lessons/${lessonId}/progress`, {
    method: 'POST',
    keepalive: true,
    credentials: 'same-origin',
    headers,
    body: JSON.stringify(progressPayload(positionSeconds, progressPercent, heartbeat)),
  });
  if (typeof response === 'object' && response !== null && 'ok' in response && response.ok === false) {
    const status = 'status' in response ? response.status : 'unknown';
    throw new Error(`Progress keepalive failed: ${status}`);
  }
}
