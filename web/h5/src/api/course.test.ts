import { describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  apiClient: { get: vi.fn() },
  unwrap: vi.fn(),
}))

import { apiClient, unwrap } from './client'
import {
  countLessons,
  courseCoverStyle,
  downloadCourseMaterial,
  enrichLessonCounts,
  getCourse,
  getResourceFile,
} from './course'
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
      { id: 'ok', title: '可用课程', description: '', cover: '', instructor: '', progress: 0, lessonCount: 0, duration: 0, category: '', materials: [] },
      { id: 'failed', title: '详情失败课程', description: '', cover: '', instructor: '', progress: 0, lessonCount: 0, duration: 0, category: '', materials: [] },
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

describe('H5 learner course materials', () => {
  it('maps materials and defaults an absent field to an empty list', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: {} })
    vi.mocked(unwrap)
      .mockReturnValueOnce({
        course: { id: 'course-1', title: '课程' },
        chapters: [],
        materials: [{
          id: 'material-1',
          display_name: '入门手册.pdf',
          resource: { resource_type: 'attachment', size_bytes: 4096 },
        }],
      })
      .mockReturnValueOnce({
        course: { id: 'course-2', title: '无资料课程' },
        chapters: [],
      })
    await expect(getCourse('course-1')).resolves.toMatchObject({
      materials: [{ id: 'material-1', displayName: '入门手册.pdf', sizeBytes: 4096 }],
    })
    await expect(getCourse('course-2')).resolves.toMatchObject({ materials: [] })
  })

  it('downloads a material through the protected blob route', async () => {
    const blob = new Blob(['guide'])
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: blob })
    await expect(downloadCourseMaterial('material-1')).resolves.toBe(blob)
    expect(apiClient.get).toHaveBeenLastCalledWith(
      '/api/v1/course-materials/material-1/download',
      { responseType: 'blob' },
    )
  })
})

describe('H5 learner video playback', () => {
  it('requests a streaming playback URL instead of buffering the full resource', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { code: 0, message: '', data: { url: '/api/v1/resource-playback/resource-1?ticket=signed' } },
    })
    vi.mocked(unwrap).mockReturnValueOnce({
      url: '/api/v1/resource-playback/resource-1?ticket=signed',
    })

    await expect(getResourceFile('resource-1')).resolves.toBe(
      '/api/v1/resource-playback/resource-1?ticket=signed',
    )
    expect(apiClient.get).toHaveBeenLastCalledWith(
      '/api/v1/resources/resource-1/playback-url',
    )
  })
})
