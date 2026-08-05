import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import { reportLessonProgress } from './progress';

vi.mock('./client', () => ({
  apiClient: { get: vi.fn(), post: vi.fn() },
}));

afterEach(() => {
  vi.restoreAllMocks();
});

describe('lesson progress API', () => {
  it('serializes a bounded heartbeat together with the position update', async () => {
    const request = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({
      data: {
        lesson_id: 'lesson-1',
        progress_percent: 35,
        status: 1,
        last_position_seconds: 42,
      },
    });

    await reportLessonProgress('lesson-1', 42.9, 35.8, {
      watched_seconds_delta: 15,
      report_id: 'report-1',
    });

    expect(request).toHaveBeenCalledWith('/api/v1/lessons/lesson-1/progress', {
      position_seconds: 42,
      progress_percent: 35,
      watched_seconds_delta: 15,
      report_id: 'report-1',
    });
  });

  it('omits heartbeat fields for a position-only compatibility update', async () => {
    const request = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({ data: {} });
    await reportLessonProgress('lesson-1', -1, 150);
    expect(request).toHaveBeenCalledWith('/api/v1/lessons/lesson-1/progress', {
      position_seconds: 0,
      progress_percent: 100,
    });
  });
});
