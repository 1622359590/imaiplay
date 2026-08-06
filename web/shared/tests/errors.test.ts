import assert from 'node:assert/strict'
import test from 'node:test'
import { responseMessage, responseStatus } from '../src/api/errors.ts'

test('reads an Axios-compatible response', () => {
  const error = { response: { status: 403, data: { message: '租户已暂停' } } }
  assert.equal(responseStatus(error), 403)
  assert.equal(responseMessage(error), '租户已暂停')
})

test('does not invent fields for unknown errors', () => {
  assert.equal(responseStatus(new Error('network')), undefined)
  assert.equal(responseMessage(new Error('network')), undefined)
})
