import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveAdminBrandName } from '../src/utils/adminBrandName.ts'

test('station header prefers its configured welcome text', () => {
  assert.equal(
    resolveAdminBrandName(' 欢迎来到测试站 ', ' Acme Academy ', 'Acme Ltd'),
    '欢迎来到测试站',
  )
  assert.equal(resolveAdminBrandName('   ', ' Acme Academy ', 'Acme Ltd'), 'Acme Academy')
  assert.equal(resolveAdminBrandName(undefined, '   ', ' Acme Ltd '), 'Acme Ltd')
  assert.equal(resolveAdminBrandName(undefined, undefined, undefined), 'ImaiPlay')
})
