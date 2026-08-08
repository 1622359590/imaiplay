import assert from 'node:assert/strict'
import test from 'node:test'
import {
  lessonContentLabel,
  resolveLessonContent,
} from '../src/learning/lessonContent.ts'

test('keeps plain text as readable body instead of treating it as a resource URL', () => {
  assert.deepEqual(resolveLessonContent('text', '第一条：进入车间请佩戴安全帽', 'blob:ignored'), {
    kind: 'text',
    body: '第一条：进入车间请佩戴安全帽',
  })
  assert.equal(lessonContentLabel('text'), '图文')
})

test('presents PDF lessons as documents without video duration metadata', () => {
  assert.deepEqual(resolveLessonContent('document', '', 'blob:pdf-preview'), {
    kind: 'document',
    source: 'blob:pdf-preview',
  })
  assert.equal(lessonContentLabel('document'), 'PDF 文档')
})

test('uses the resolved playback URL only for video lessons', () => {
  assert.deepEqual(resolveLessonContent('video', 'https://example.com/fallback.mp4', 'blob:video'), {
    kind: 'video',
    source: 'blob:video',
  })
  assert.equal(lessonContentLabel('video'), '视频')
})

test('reports missing document and video resources as empty content', () => {
  assert.deepEqual(resolveLessonContent('document', '', undefined), { kind: 'empty' })
  assert.deepEqual(resolveLessonContent('video', '   ', undefined), { kind: 'empty' })
})
