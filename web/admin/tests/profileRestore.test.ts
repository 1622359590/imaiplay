import assert from 'node:assert/strict'
import test from 'node:test'
import { handleProfileRestoreFailure } from '../src/utils/profileRestore.ts'

test('profile restore clears persistent authentication on both 401 and 403', () => {
  for (const status of [401, 403]) {
    let cleared = 0
    assert.equal(handleProfileRestoreFailure(status, () => { cleared += 1 }), 'unauthorized')
    assert.equal(cleared, 1)
  }
})

test('profile restore preserves the session for retryable failures', () => {
  let cleared = 0
  assert.equal(handleProfileRestoreFailure(undefined, () => { cleared += 1 }), 'retry')
  assert.equal(handleProfileRestoreFailure(500, () => { cleared += 1 }), 'retry')
  assert.equal(cleared, 0)
})
