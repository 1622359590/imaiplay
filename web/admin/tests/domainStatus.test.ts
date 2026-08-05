import assert from 'node:assert/strict'
import test from 'node:test'
import {
  mergeDomainStatus,
  portalURLAfterRegistration,
} from '../src/utils/domainStatus.ts'
import type { DomainBindStatus } from '../src/api/domain.ts'

const initial: DomainBindStatus = {
  state: 'none',
  current_step: 0,
  total_steps: 5,
  cname_target: 'play.imai.work',
  tenant_code: 'acme',
  default_portal_url: 'https://academy.example.com/t/acme',
}

test('keeps default portal metadata when an action returns only workflow state', () => {
  const next = mergeDomainStatus(initial, {
    state: 'verified',
    domain: 'learn.acme.com',
    current_step: 1,
    total_steps: 5,
    cname_target: 'play.imai.work',
  })

  assert.equal(next.tenant_code, 'acme')
  assert.equal(next.default_portal_url, 'https://academy.example.com/t/acme')
  assert.equal(next.state, 'verified')
})

test('uses backend portal URL after registration and a production fallback otherwise', () => {
  assert.equal(
    portalURLAfterRegistration(initial, 'ignored'),
    'https://academy.example.com/t/acme',
  )
  assert.equal(
    portalURLAfterRegistration(undefined, 'new tenant'),
    'https://play.imai.work/t/new%20tenant',
  )
})
