import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveAdminBrandName } from '../src/utils/adminBrandName.ts'

test('tenant brand name falls back without coupling to browser title', () => {
  assert.equal(resolveAdminBrandName(' Acme Academy ', 'Acme Ltd'), 'Acme Academy')
  assert.equal(resolveAdminBrandName('   ', ' Acme Ltd '), 'Acme Ltd')
  assert.equal(resolveAdminBrandName(undefined, undefined), 'ImaiPlay')
})
