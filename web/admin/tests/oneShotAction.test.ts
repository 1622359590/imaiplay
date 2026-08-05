import assert from 'node:assert/strict'
import test from 'node:test'
import { consumeOneShotAction } from '../src/utils/oneShotAction.ts'
import {
  normalizeCourseEditValues,
  normalizeEnrollmentEditValues,
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
  }), {
    title: '安全入门',
    description: '基础课',
    status: 1,
    category_id: 'category-1',
  })
})

test('enrollment edit normalization preserves assignment type', () => {
  assert.deepEqual(normalizeEnrollmentEditValues({
    user_id: 'learner-1',
    assignment_type: 'optional',
  }), {
    user_id: 'learner-1',
    assignment_type: 'optional',
  })
})
