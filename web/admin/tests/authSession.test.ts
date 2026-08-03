import assert from 'node:assert/strict'
import test from 'node:test'
import {
  ADMIN_ACCESS_TOKEN_KEY,
  ADMIN_REFRESH_TOKEN_KEY,
  AUTH_SESSION_EXPIRED_EVENT,
  LEGACY_ACCESS_TOKEN_KEY,
  LEGACY_REFRESH_TOKEN_KEY,
  createAdminLogoutRequest,
  createSessionRefresher,
  clearAuthSession,
  migrateLegacyAdminSession,
  writeAdminSession,
} from '../src/api/authSession.ts'

function token(claims: Record<string, unknown>) {
  const encode = (value: unknown) => Buffer.from(JSON.stringify(value)).toString('base64url')
  return `${encode({ alg: 'none' })}.${encode(claims)}.signature`
}

const tenantAdminToken = token({
  user_id: 'admin-1', tenant_id: 'tenant-1', role: 'tenant_admin', exp: 4_102_444_800,
})
const anotherTenantAdminToken = token({
  user_id: 'admin-2', tenant_id: 'tenant-2', role: 'tenant_admin', exp: 4_102_444_800,
})
const learnerToken = token({
  user_id: 'learner-1', tenant_id: 'tenant-1', role: 'learner', exp: 4_102_444_800,
})

function memoryStorage(values: Record<string, string> = {}) {
  const entries = new Map(Object.entries(values))
  return {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => entries.set(key, value),
    removeItem: (key: string) => entries.delete(key),
    value: (key: string) => entries.get(key),
  }
}

test('renews an expired access token once for concurrent requests', async () => {
  const storage = memoryStorage({ [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-old' })
  let calls = 0
  let finish!: (pair: { token: string; refresh_token: string }) => void
  const response = new Promise<{ token: string; refresh_token: string }>((resolve) => {
    finish = resolve
  })
  const refresh = createSessionRefresher(async (refreshToken) => {
    calls += 1
    assert.equal(refreshToken, 'refresh-old')
    return response
  }, storage)

  const first = refresh()
  const second = refresh()
  assert.equal(calls, 1)

  finish({ token: tenantAdminToken, refresh_token: 'refresh-new' })
  assert.deepEqual(await Promise.all([first, second]), [tenantAdminToken, tenantAdminToken])
  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), tenantAdminToken)
  assert.equal(storage.value(ADMIN_REFRESH_TOKEN_KEY), 'refresh-new')
})

test('does not restore an old session when logout happens during refresh', async () => {
  const storage = memoryStorage({
    [ADMIN_ACCESS_TOKEN_KEY]: tenantAdminToken,
    [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-old',
  })
  let finish!: (pair: { token: string; refresh_token: string }) => void
  const response = new Promise<{ token: string; refresh_token: string }>((resolve) => {
    finish = resolve
  })
  const refresh = createSessionRefresher(async () => response, storage)

  const pending = refresh()
  clearAuthSession(storage, { dispatchEvent: () => undefined })
  finish({ token: tenantAdminToken, refresh_token: 'refresh-restored' })

  await assert.rejects(pending, /refresh superseded/)
  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), undefined)
  assert.equal(storage.value(ADMIN_REFRESH_TOKEN_KEY), undefined)
})

test('does not overwrite a new login when an old refresh finishes later', async () => {
  const storage = memoryStorage({
    [ADMIN_ACCESS_TOKEN_KEY]: tenantAdminToken,
    [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-old',
  })
  let finish!: (pair: { token: string; refresh_token: string }) => void
  const response = new Promise<{ token: string; refresh_token: string }>((resolve) => {
    finish = resolve
  })
  const refresh = createSessionRefresher(async () => response, storage)

  const pending = refresh()
  // A backend may preserve a refresh token across a fast account switch, so
  // generation must protect the session even when the token text is unchanged.
  writeAdminSession({ token: anotherTenantAdminToken, refresh_token: 'refresh-old' }, storage)
  finish({ token: tenantAdminToken, refresh_token: 'refresh-old-rotated' })

  await assert.rejects(pending, /refresh superseded/)
  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), anotherTenantAdminToken)
  assert.equal(storage.value(ADMIN_REFRESH_TOKEN_KEY), 'refresh-old')
})

test('clears both tokens and notifies the app after refresh can no longer recover', () => {
  const storage = memoryStorage({
    [ADMIN_ACCESS_TOKEN_KEY]: 'access-old',
    [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-old',
  })
  const events: string[] = []

  clearAuthSession(storage, { dispatchEvent: (event) => events.push(event.type) })

  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), undefined)
  assert.equal(storage.value(ADMIN_REFRESH_TOKEN_KEY), undefined)
  assert.deepEqual(events, [AUTH_SESSION_EXPIRED_EVENT])
})

test('migrates a valid legacy staff token into the admin key', () => {
  const storage = memoryStorage({ imaiplay_token: tenantAdminToken })

  migrateLegacyAdminSession(storage)

  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), tenantAdminToken)
  assert.equal(storage.value('imaiplay_token'), undefined)
})

test('leaves learner legacy tokens for portal migration', () => {
  const storage = memoryStorage({ imaiplay_token: learnerToken })

  migrateLegacyAdminSession(storage)

  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), undefined)
  assert.equal(storage.value('imaiplay_token'), learnerToken)
})

test('removes a valid legacy staff session even when a scoped admin session already exists', () => {
  const storage = memoryStorage({
    [ADMIN_ACCESS_TOKEN_KEY]: anotherTenantAdminToken,
    [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-scoped',
    [LEGACY_ACCESS_TOKEN_KEY]: tenantAdminToken,
    [LEGACY_REFRESH_TOKEN_KEY]: 'refresh-legacy-staff',
  })

  migrateLegacyAdminSession(storage)

  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), anotherTenantAdminToken)
  assert.equal(storage.value(ADMIN_REFRESH_TOKEN_KEY), 'refresh-scoped')
  assert.equal(storage.value(LEGACY_ACCESS_TOKEN_KEY), undefined)
  assert.equal(storage.value(LEGACY_REFRESH_TOKEN_KEY), undefined)

  clearAuthSession(storage, { dispatchEvent: () => undefined })
  migrateLegacyAdminSession(storage)
  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), undefined)
})

test('keeps a learner legacy session when a scoped admin session exists', () => {
  const storage = memoryStorage({
    [ADMIN_ACCESS_TOKEN_KEY]: tenantAdminToken,
    [LEGACY_ACCESS_TOKEN_KEY]: learnerToken,
    [LEGACY_REFRESH_TOKEN_KEY]: 'refresh-learner',
  })

  migrateLegacyAdminSession(storage)

  assert.equal(storage.value(LEGACY_ACCESS_TOKEN_KEY), learnerToken)
  assert.equal(storage.value(LEGACY_REFRESH_TOKEN_KEY), 'refresh-learner')
})

test('replaces a stale refresh token when a new session has no refresh token', () => {
  const storage = memoryStorage({ [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-from-old-account' })

  writeAdminSession({ token: tenantAdminToken }, storage)

  assert.equal(storage.value(ADMIN_ACCESS_TOKEN_KEY), tenantAdminToken)
  assert.equal(storage.value(ADMIN_REFRESH_TOKEN_KEY), undefined)
})

test('captures the access token for an authenticated logout request', () => {
  const storage = memoryStorage({
    [ADMIN_ACCESS_TOKEN_KEY]: tenantAdminToken,
    [ADMIN_REFRESH_TOKEN_KEY]: 'refresh-current',
  })

  assert.deepEqual(createAdminLogoutRequest(storage), {
    refreshToken: 'refresh-current',
    authorization: `Bearer ${tenantAdminToken}`,
  })
})
