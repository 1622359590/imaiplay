import assert from 'node:assert/strict'
import test from 'node:test'
import { consumeOneShotAction } from '../src/utils/oneShotAction.ts'
import {
  categoryIDForPayload,
  normalizeCourseEditValues,
} from '../src/utils/adminFormValues.ts'

test('consumes a one-shot action and preserves unrelated query parameters', () => {
  assert.deepEqual(consumeOneShotAction('?create=1', 'create'), {
    active: true,
    remainingSearch: '',
  })
  assert.deepEqual(consumeOneShotAction('?create=1&page=2', 'create'), {
    active: true,
    remainingSearch: '?page=2',
  })
  assert.deepEqual(consumeOneShotAction('?create=0&page=2', 'create'), {
    active: false,
    remainingSearch: '?create=0&page=2',
  })
})

test('course edit normalization preserves the category relationship', () => {
  assert.deepEqual(normalizeCourseEditValues({
    title: '安全入门',
    description: '基础课',
    status: 1,
    category_id: 'category-1',
    course_type: 'optional',
  }), {
    title: '安全入门',
    description: '基础课',
    status: 1,
    category_id: 'category-1',
    course_type: 'optional',
  })
})

test('course saves send a selected category or an explicit null when cleared', () => {
  assert.equal(categoryIDForPayload('category-1'), 'category-1')
  assert.equal(categoryIDForPayload('  '), null)
  assert.equal(categoryIDForPayload(undefined), null)
  assert.equal(normalizeCourseEditValues({
    title: '未分类课程',
    status: 0,
    category_id: null,
    course_type: 'required',
  }).category_id, undefined)
})
