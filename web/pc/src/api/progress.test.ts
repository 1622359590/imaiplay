import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import { getLessonProgress, reportLessonProgress } from './progress';

vi.mock('./client', () => ({
  apiClient: { get: vi.fn(), post: vi.fn() },
}));

afterEach(() => {
  vi.clearAllMocks();
  vi.restoreAllMocks();
});

describe('lesson progress API', () => {
  it.each([1, 60])('preserves a valid %i-second heartbeat exactly', async (delta) => {
    const request = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({
      data: {
        lesson_id: 'lesson-1',
        progress_percent: 35,
        status: 1,
        last_position_seconds: 42,
      },
    });

    await expect(reportLessonProgress('lesson-1', 42.9, 35.8, {
      watched_seconds_delta: delta,
      report_id: 'report-1',
    })).resolves.toEqual({
      lessonId: 'lesson-1',
      progressPercent: 35,
      status: 1,
      lastPositionSeconds: 42,
    });

    expect(request).toHaveBeenCalledWith('/api/v1/lessons/lesson-1/progress', {
      position_seconds: 42,
      progress_percent: 35,
      watched_seconds_delta: delta,
      report_id: 'report-1',
    });
  });

  it.each([
    { watched_seconds_delta: 0, report_id: 'report' },
    { watched_seconds_delta: -1, report_id: 'report' },
    { watched_seconds_delta: Number.NaN, report_id: 'report' },
    { watched_seconds_delta: Number.POSITIVE_INFINITY, report_id: 'report' },
    { watched_seconds_delta: Number.NEGATIVE_INFINITY, report_id: 'report' },
    { watched_seconds_delta: 61, report_id: 'report' },
    { watched_seconds_delta: 1.5, report_id: 'report' },
    { watched_seconds_delta: 1, report_id: '' },
    { watched_seconds_delta: 1, report_id: '   ' },
  ])('rejects an invalid heartbeat before POSTing: $watched_seconds_delta/$report_id', async (heartbeat) => {
    const request = vi.spyOn(apiClient, 'post');
    await expect(reportLessonProgress('lesson-1', 42, 35, heartbeat)).rejects.toThrow(
      'Invalid progress heartbeat',
    );
    expect(request).not.toHaveBeenCalled();
  });

  it.each([
    [Number.NaN, 20],
    [Number.POSITIVE_INFINITY, 20],
    [20, Number.NaN],
    [20, Number.NEGATIVE_INFINITY],
  ])('rejects non-finite position/progress before POSTing', async (position, percent) => {
    const request = vi.spyOn(apiClient, 'post');
    await expect(reportLessonProgress('lesson-1', position, percent)).rejects.toThrow(
      'Invalid lesson progress',
    );
    expect(request).not.toHaveBeenCalled();
  });

  it('normalizes GET progress into a camelCase view model', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({
      data: {
        lesson_id: 'lesson-1',
        progress_percent: 35,
        status: 1,
        last_position_seconds: 42,
      },
    });

    await expect(getLessonProgress('lesson-1')).resolves.toEqual({
      lessonId: 'lesson-1',
      progressPercent: 35,
      status: 1,
      lastPositionSeconds: 42,
    });
  });

  it('omits heartbeat fields for a position-only compatibility update', async () => {
    const request = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({
      data: {
        lesson_id: 'lesson-1',
        progress_percent: 100,
        status: 2,
        last_position_seconds: 0,
      },
    });
    await expect(reportLessonProgress('lesson-1', -1, 150)).resolves.toEqual({
      lessonId: 'lesson-1',
      progressPercent: 100,
      status: 2,
      lastPositionSeconds: 0,
    });
    expect(request).toHaveBeenCalledWith('/api/v1/lessons/lesson-1/progress', {
      position_seconds: 0,
      progress_percent: 100,
    });
  });
});
