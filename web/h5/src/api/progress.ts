import { apiClient, unwrap, type ApiEnvelope } from './client'

export interface LessonProgress {
  progress_percent: number
  last_position_seconds: number
}

export async function getLessonProgress(lessonId: string): Promise<LessonProgress> {
  return unwrap(await apiClient.get<ApiEnvelope<LessonProgress>>(`/api/v1/lessons/${lessonId}/progress`))
}

export async function reportLessonProgress(
  lessonId: string,
  positionSeconds: number,
  progressPercent: number,
): Promise<LessonProgress> {
  return unwrap(await apiClient.post<ApiEnvelope<LessonProgress>>(`/api/v1/lessons/${lessonId}/progress`, {
    position_seconds: Math.max(0, Math.floor(positionSeconds)),
    progress_percent: Math.max(0, Math.min(100, Math.floor(progressPercent))),
  }))
}
