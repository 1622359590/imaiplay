import { describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  apiClient: { get: vi.fn() },
  unwrap: vi.fn(),
}))

import { countLessons, courseCoverStyle, enrichLessonCounts } from './course'
import type { Course } from '../types/course'

describe('H5 learner course presentation data', () => {
  it('leaves a missing cover unset so tenant brand CSS can render the fallback', () => {
    expect(courseCoverStyle()).toBeUndefined()
    expect(courseCoverStyle('/covers/course.png')).toBe('url("/covers/course.png") center/cover')
  })

  it('counts lessons across chapters', () => {
    expect(countLessons([
      {
        id: 'chapter-1',
        title: '第一章',
        lessons: [
          { id: 'lesson-1', title: '课时一', duration: 2 },
          { id: 'lesson-2', title: '课时二', duration: 3 },
        ],
      },
    ])).toBe(2)
  })

  it('does not invent a count when detail loading fails', async () => {
    const courses: Course[] = [
      { id: 'ok', title: '可用课程', description: '', cover: '', instructor: '', progress: 0, lessonCount: 0, duration: 0, category: '' },
      { id: 'failed', title: '详情失败课程', description: '', cover: '', instructor: '', progress: 0, lessonCount: 0, duration: 0, category: '' },
    ]
    const loadDetail = vi.fn(async (id: string): Promise<Course> => {
      if (id === 'failed') throw new Error('detail failed')
      return {
        ...courses[0],
        chapters: [
          { id: 'chapter', title: '章节', lessons: [{ id: 'lesson', title: '课时', duration: 1 }] },
        ],
      }
    })

    const result = await enrichLessonCounts(courses, loadDetail)
    expect(result[0].lessonCount).toBe(1)
    expect(result[1].lessonCount).toBeUndefined()
  })
})
