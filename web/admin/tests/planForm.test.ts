import assert from 'node:assert/strict'
import test from 'node:test'
import {
  buildPlanInput,
  createPlanFormValues,
} from '../src/utils/planForm.ts'

test('new plans are submitted as enabled by default', () => {
  const values = createPlanFormValues()
  values.name = '免费版'

  const input = buildPlanInput(values)

  assert.equal(input.status, 1)
})

test('editing a plan preserves an explicit enabled status', () => {
  const values = createPlanFormValues({
    id: 'plan-1',
    name: '旗舰版',
    storage_quota_bytes: 50 * 1024 ** 3,
    max_users: 0,
    max_courses: 0,
    features: '{}',
    is_default: false,
    status: 1,
  })

  const input = buildPlanInput(values)

  assert.equal(input.status, 1)
  assert.equal(input.storage_quota_bytes, 50 * 1024 ** 3)
})
