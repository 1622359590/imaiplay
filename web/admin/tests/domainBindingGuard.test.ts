import assert from 'node:assert/strict'
import test from 'node:test'
import {
  domainBindingProgress,
  shouldGuardDomainBinding,
} from '../src/utils/domainBindingGuard.ts'

test('guards only long-running domain provisioning states', () => {
  assert.equal(shouldGuardDomainBinding('creating_site'), true)
  assert.equal(shouldGuardDomainBinding('configuring'), true)
  assert.equal(shouldGuardDomainBinding('pending_verification'), false)
  assert.equal(shouldGuardDomainBinding('ready'), false)
  assert.equal(shouldGuardDomainBinding('setup_failed'), false)
})

test('calculates a bounded whole-number binding progress percentage', () => {
  assert.equal(domainBindingProgress({ current_step: 3, total_steps: 5 }), 60)
  assert.equal(domainBindingProgress({ current_step: 8, total_steps: 5 }), 100)
  assert.equal(domainBindingProgress({ current_step: 0, total_steps: 0 }), 0)
})
