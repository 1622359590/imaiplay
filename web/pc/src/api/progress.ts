import { apiClient } from './client';

export interface LessonProgress {
  lessonId: string;
  progressPercent: number;
  status: number;
  lastPositionSeconds: number;
}

export interface ProgressHeartbeat {
  watched_seconds_delta: number;
  report_id: string;
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
  ) {
    throw new Error('Invalid progress heartbeat');
  }
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
  if (!Number.isFinite(positionSeconds) || !Number.isFinite(progressPercent)) {
    throw new Error('Invalid lesson progress');
  }
  if (heartbeat) validateHeartbeat(heartbeat);
  const response = await apiClient.post<RawLessonProgress>(`/api/v1/lessons/${lessonId}/progress`, {
    position_seconds: Math.max(0, Math.floor(positionSeconds)),
    progress_percent: Math.max(0, Math.min(100, Math.floor(progressPercent))),
    ...(heartbeat ? {
      watched_seconds_delta: heartbeat.watched_seconds_delta,
      report_id: heartbeat.report_id,
    } : {}),
  });
  return mapLessonProgress(response.data);
}
