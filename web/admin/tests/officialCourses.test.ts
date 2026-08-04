import assert from 'node:assert/strict'
import test from 'node:test'
import type { Course } from '../src/api/course.ts'
import { updateOfficialCourseEnabled } from '../src/utils/officialCourses.ts'

test('updates one official course without mutating the existing list', () => {
  const courses: Course[] = [
    {
      id: 'official-1',
      title: '公共认知',
      status: 1,
      is_official: true,
      enabled: false,
    },
    {
      id: 'official-2',
      title: '成交与交付课',
      status: 1,
      is_official: true,
      enabled: true,
    },
  ]

  const updated = updateOfficialCourseEnabled(courses, 'official-1', true)

  assert.notEqual(updated, courses)
  assert.equal(courses[0].enabled, false)
  assert.equal(updated[0].enabled, true)
  assert.equal(updated[1], courses[1])
})
