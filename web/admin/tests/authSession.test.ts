import assert from 'node:assert/strict'
import test from 'node:test'
import {
  AUTH_SESSION_EXPIRED_EVENT,
  createSessionRefresher,
  clearAuthSession,
  REFRESH_TOKEN_KEY,
  TOKEN_KEY,
} from '../src/api/authSession.ts'

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
  const storage = memoryStorage({ [REFRESH_TOKEN_KEY]: 'refresh-old' })
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

  finish({ token: 'access-new', refresh_token: 'refresh-new' })
  assert.deepEqual(await Promise.all([first, second]), ['access-new', 'access-new'])
  assert.equal(storage.value(TOKEN_KEY), 'access-new')
  assert.equal(storage.value(REFRESH_TOKEN_KEY), 'refresh-new')
})

test('clears both tokens and notifies the app after refresh can no longer recover', () => {
  const storage = memoryStorage({
    [TOKEN_KEY]: 'access-old',
    [REFRESH_TOKEN_KEY]: 'refresh-old',
  })
  const events: string[] = []

  clearAuthSession(storage, { dispatchEvent: (event) => events.push(event.type) })

  assert.equal(storage.value(TOKEN_KEY), undefined)
  assert.equal(storage.value(REFRESH_TOKEN_KEY), undefined)
  assert.deepEqual(events, [AUTH_SESSION_EXPIRED_EVENT])
})
