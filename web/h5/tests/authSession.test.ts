import assert from 'node:assert/strict'
import test from 'node:test'
import * as authSession from '../src/api/authSession.ts'
import {
  portalErrorContent,
  resolvePortalLocation,
  shouldRestoreSessionPortal,
} from '../src/api/portalResolution.ts'

interface StorageLike {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
}

function memoryStorage(): StorageLike {
  const entries = new Map<string, string>()
  return {
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => entries.set(key, value),
    removeItem: (key) => entries.delete(key),
  }
}

function token(payload: Record<string, unknown>): string {
  return `header.${Buffer.from(JSON.stringify(payload)).toString('base64url')}.signature`
}

test('uses the shared portal-scoped token keys', () => {
  assert.equal(authSession.PORTAL_ACCESS_TOKEN_KEY, 'imaiplay_portal_access_token')
  assert.equal(authSession.PORTAL_REFRESH_TOKEN_KEY, 'imaiplay_portal_refresh_token')
})

test('accepts only an unexpired learner session for the resolved tenant', () => {
  const nowSeconds = 2_000_000_000
  const learner = token({
    user_id: 'learner-1',
    tenant_id: 'tenant-acme',
    role: 'learner',
    exp: nowSeconds + 60,
  })

  assert.equal(authSession.isValidPortalSession(learner, 'tenant-acme', nowSeconds * 1000), true)
  assert.equal(authSession.isValidPortalSession(learner, 'tenant-bravo', nowSeconds * 1000), false)
  assert.equal(
    authSession.isValidPortalSession(
      token({ user_id: 'staff-1', tenant_id: 'tenant-acme', role: 'instructor', exp: nowSeconds + 60 }),
      'tenant-acme',
      nowSeconds * 1000,
    ),
    false,
  )
  assert.equal(
    authSession.isValidPortalSession(
      token({ user_id: 'learner-1', tenant_id: 'tenant-acme', role: 'learner', exp: nowSeconds }),
      'tenant-acme',
      nowSeconds * 1000,
    ),
    false,
  )
})

test('migrates only a valid learner legacy token and leaves staff for admin', () => {
  const nowSeconds = 2_000_000_000
  const learnerStorage = memoryStorage()
  const learner = token({
    user_id: 'learner-1',
    tenant_id: 'tenant-acme',
    role: 'learner',
    exp: nowSeconds + 60,
  })
  learnerStorage.setItem(authSession.LEGACY_TOKEN_KEY, learner)

  authSession.migrateLegacyPortalSession(learnerStorage, nowSeconds * 1000)

  assert.equal(learnerStorage.getItem(authSession.PORTAL_ACCESS_TOKEN_KEY), learner)
  assert.equal(learnerStorage.getItem(authSession.LEGACY_TOKEN_KEY), null)

  const staffStorage = memoryStorage()
  const staff = token({
    user_id: 'staff-1',
    tenant_id: 'tenant-acme',
    role: 'tenant_admin',
    exp: nowSeconds + 60,
  })
  staffStorage.setItem(authSession.LEGACY_TOKEN_KEY, staff)

  authSession.migrateLegacyPortalSession(staffStorage, nowSeconds * 1000)

  assert.equal(staffStorage.getItem(authSession.PORTAL_ACCESS_TOKEN_KEY), null)
  assert.equal(staffStorage.getItem(authSession.LEGACY_TOKEN_KEY), staff)
})

test('removes a valid learner legacy token when a portal session already exists', () => {
  const nowSeconds = 2_000_000_000
  const storage = memoryStorage()
  const currentPortalToken = token({
    user_id: 'learner-current',
    tenant_id: 'tenant-acme',
    role: 'learner',
    exp: nowSeconds + 60,
  })
  const legacyLearnerToken = token({
    user_id: 'learner-legacy',
    tenant_id: 'tenant-acme',
    role: 'learner',
    exp: nowSeconds + 60,
  })
  storage.setItem(authSession.PORTAL_ACCESS_TOKEN_KEY, currentPortalToken)
  storage.setItem(authSession.LEGACY_TOKEN_KEY, legacyLearnerToken)

  authSession.migrateLegacyPortalSession(storage, nowSeconds * 1000)

  assert.equal(storage.getItem(authSession.PORTAL_ACCESS_TOKEN_KEY), currentPortalToken)
  assert.equal(storage.getItem(authSession.LEGACY_TOKEN_KEY), null)
})

test('binds the portal session only to the resolved tenant', () => {
  const nowSeconds = 2_000_000_000
  const storage = memoryStorage()
  storage.setItem(
    authSession.PORTAL_ACCESS_TOKEN_KEY,
    token({
      user_id: 'learner-1',
      tenant_id: 'tenant-bravo',
      role: 'learner',
      exp: nowSeconds + 60,
    }),
  )

  assert.equal(
    authSession.bindPortalSession(
      { code: 'acme', tenant_id: 'tenant-acme' },
      storage,
      nowSeconds * 1000,
    ),
    false,
  )
  assert.equal(storage.getItem(authSession.PORTAL_ACCESS_TOKEN_KEY), null)
  assert.equal(storage.getItem(authSession.PORTAL_REFRESH_TOKEN_KEY), null)
})

test('clears both portal tokens and dispatches the H5 expiry event', () => {
  const storage = memoryStorage()
  const events: string[] = []
  storage.setItem(authSession.PORTAL_ACCESS_TOKEN_KEY, 'access')
  storage.setItem(authSession.PORTAL_REFRESH_TOKEN_KEY, 'refresh')

  authSession.clearPortalSession(storage, {
    dispatchEvent: (event) => events.push(event.type),
  })

  assert.equal(storage.getItem(authSession.PORTAL_ACCESS_TOKEN_KEY), null)
  assert.equal(storage.getItem(authSession.PORTAL_REFRESH_TOKEN_KEY), null)
  assert.deepEqual(events, ['imaiplay:portal-session-expired'])
})

test('invalidates refreshes that started before the portal session was cleared', () => {
  const storage = memoryStorage()
  storage.setItem(authSession.PORTAL_REFRESH_TOKEN_KEY, 'refresh-before-logout')
  const generation = authSession.getPortalSessionGeneration()

  assert.equal(
    authSession.isPortalSessionCurrent(generation, 'refresh-before-logout', storage),
    true,
  )

  authSession.clearPortalSession(storage, { dispatchEvent: () => undefined })

  assert.equal(
    authSession.isPortalSessionCurrent(generation, 'refresh-before-logout', storage),
    false,
  )
})

test('does not expire a newer session when an older refresh finishes late', () => {
  assert.equal(
    authSession.shouldExpirePortalSessionAfterRefresh(
      new authSession.PortalSessionChangedError(),
    ),
    false,
  )
  assert.equal(
    authSession.shouldExpirePortalSessionAfterRefresh(new Error('refresh failed')),
    true,
  )
})

test('classifies every failed refresh against the session that started it', () => {
  const storage = memoryStorage()
  const networkError = new Error('network failed')
  storage.setItem(authSession.PORTAL_REFRESH_TOKEN_KEY, 'refresh-before-logout')
  const generation = authSession.getPortalSessionGeneration()

  assert.equal(
    authSession.classifyPortalRefreshFailure(
      networkError,
      generation,
      'refresh-before-logout',
      storage,
    ),
    networkError,
  )

  authSession.clearPortalSession(storage, { dispatchEvent: () => undefined })

  assert.ok(
    authSession.classifyPortalRefreshFailure(
      networkError,
      generation,
      'refresh-before-logout',
      storage,
    ) instanceof authSession.PortalSessionChangedError,
  )
})

test('replaces the session without retaining a previous refresh token', () => {
  const nowSeconds = Math.floor(Date.now() / 1000)
  const storage = memoryStorage()
  storage.setItem(authSession.PORTAL_REFRESH_TOKEN_KEY, 'stale-refresh')
  const access = token({
    user_id: 'learner-1',
    tenant_id: 'tenant-acme',
    role: 'learner',
    exp: nowSeconds + 60,
  })

  authSession.writePortalSession({ token: access }, 'acme', storage)

  assert.equal(storage.getItem(authSession.PORTAL_ACCESS_TOKEN_KEY), access)
  assert.equal(storage.getItem(authSession.PORTAL_REFRESH_TOKEN_KEY), null)
})

test('builds a tenant-aware H5 login path', () => {
  assert.equal(authSession.portalLoginPath('acme', 'play.imai.work'), '/h5/t/acme/login')
  assert.equal(authSession.portalLoginPath('acme', 'learn.acme.com'), '/h5/login')
})

test('keeps navigation inside the active default or custom-domain portal', () => {
  assert.equal(
    authSession.portalRoutePath('acme', 'play.imai.work', '/courses/42'),
    '/t/acme/courses/42',
  )
  assert.equal(
    authSession.portalRoutePath('acme', 'learn.acme.com', '/courses/42'),
    '/courses/42',
  )
})

test('coalesces concurrent refreshes into one request', async () => {
  const run = authSession.createSingleFlight<string>()
  let requests = 0
  let resolve!: (value: string) => void
  const refresh = () => {
    requests += 1
    return new Promise<string>((done) => { resolve = done })
  }

  const first = run(refresh)
  const second = run(refresh)

  assert.equal(first, second)
  assert.equal(requests, 1)
  resolve('refreshed')
  assert.equal(await first, 'refreshed')

  assert.equal(await run(async () => {
    requests += 1
    return 'again'
  }), 'again')
  assert.equal(requests, 2)
})

test('tracks the active portal identity and clears a stale tenant id while resolving', () => {
  authSession.setActivePortalIdentity({ code: 'Acme', tenant_id: 'tenant-acme' })
  assert.equal(authSession.getActivePortalCode(), 'acme')
  assert.equal(authSession.getActivePortalTenantId(), 'tenant-acme')

  authSession.setActivePortalCode('bravo')
  assert.equal(authSession.getActivePortalCode(), 'bravo')
  assert.equal(authSession.getActivePortalTenantId(), undefined)
})

test('custom domain takes priority over a conflicting tenant path', () => {
  assert.deepEqual(
    resolvePortalLocation('/t/bravo/courses', 'Learn.Acme.Test'),
    {
      tenantCode: undefined,
      mode: 'custom-domain',
      shouldResolve: true,
      key: 'custom-domain:learn.acme.test',
    },
  )
  assert.deepEqual(
    resolvePortalLocation('/t/acme/courses', 'play.imai.work'),
    {
      tenantCode: 'acme',
      mode: 'default',
      shouldResolve: true,
      key: 'default:acme',
    },
  )
})

test('restores a valid learner portal session only on the platform route', () => {
  const nowSeconds = 2_000_000_000
  const learner = token({
    user_id: 'learner-1',
    tenant_id: 'tenant-acme',
    role: 'learner',
    exp: nowSeconds + 60,
  })
  const platform = resolvePortalLocation('/', 'play.imai.work')
  const tenantPath = resolvePortalLocation('/t/acme', 'play.imai.work')

  assert.equal(
    shouldRestoreSessionPortal(platform, learner, nowSeconds * 1000),
    true,
  )
  assert.equal(
    shouldRestoreSessionPortal(tenantPath, learner, nowSeconds * 1000),
    false,
  )
})

test('maps portal API errors to specific tenant-facing states', () => {
  assert.equal(
    portalErrorContent({ response: { status: 404, data: { error: 'portal_not_found' } } }).title,
    '门户不存在',
  )
  assert.equal(
    portalErrorContent({ response: { status: 403, data: { error: 'portal_suspended' } } }).title,
    '租户已暂停',
  )
  assert.equal(
    portalErrorContent({ response: { status: 403, data: { error: 'portal_trial_expired' } } }).title,
    '试用已到期',
  )
  assert.equal(
    portalErrorContent({ response: { status: 403, data: { message: '请求失败，请稍后重试' } } }).title,
    '企业门户不可访问',
  )
  assert.equal(portalErrorContent(new Error('network')).title, '企业门户不可访问')
})
