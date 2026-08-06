import assert from 'node:assert/strict'
import test from 'node:test'
import { lessonPayload } from '../src/pages/course-detail/courseDetailModel.ts'

test('resource lessons discard stale text content', () => {
  assert.deepEqual(lessonPayload({
    title: '视频课', content_type: 'video', content_url: 'stale',
    resource_id: 'resource-1', duration_seconds: 0, sort_order: 0,
  }), {
    title: '视频课', content_type: 'video', content_url: '',
    resource_id: 'resource-1', duration_seconds: 0, sort_order: 0,
  })
})

test('text lessons preserve content and discard stale resources', () => {
  assert.deepEqual(lessonPayload({
    title: '文本课', content_type: 'text', content_url: '正文',
    resource_id: 'resource-1', duration_seconds: 0, sort_order: 0,
  }), {
    title: '文本课', content_type: 'text', content_url: '正文',
    resource_id: undefined, duration_seconds: 0, sort_order: 0,
  })
})
