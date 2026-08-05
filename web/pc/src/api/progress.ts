import { apiClient } from './client';

export interface LessonProgress {
  lesson_id: string;
  progress_percent: number;
  status: number;
  last_position_seconds: number;
}

export interface ProgressHeartbeat {
  watched_seconds_delta: number;
  report_id: string;
}

export async function getLessonProgress(lessonId: string): Promise<LessonProgress> {
  const response = await apiClient.get<LessonProgress>(`/api/v1/lessons/${lessonId}/progress`);
  return response.data;
}

export async function reportLessonProgress(
  lessonId: string,
  positionSeconds: number,
  progressPercent: number,
  heartbeat?: ProgressHeartbeat,
): Promise<LessonProgress> {
  const response = await apiClient.post<LessonProgress>(`/api/v1/lessons/${lessonId}/progress`, {
    position_seconds: Math.max(0, Math.floor(positionSeconds)),
    progress_percent: Math.max(0, Math.min(100, Math.floor(progressPercent))),
    ...(heartbeat ? {
      watched_seconds_delta: Math.max(1, Math.min(60, Math.floor(heartbeat.watched_seconds_delta))),
      report_id: heartbeat.report_id,
    } : {}),
  });
  return response.data;
}
