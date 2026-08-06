import assert from 'node:assert/strict'
import test from 'node:test'
import {
  createRefreshCoordinator,
  markSessionChanged,
  type StorageLike,
} from '../src/auth/sessionCore.ts'

function memoryStorage(values: Record<string, string> = {}): StorageLike & {
  value(key: string): string | undefined
} {
  const entries = new Map(Object.entries(values))
  return {
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => entries.set(key, value),
    removeItem: (key) => entries.delete(key),
    value: (key) => entries.get(key),
  }
}

test('coalesces concurrent refreshes and stores rotated tokens', async () => {
  const storage = memoryStorage({ refresh: 'refresh-old' })
  let calls = 0
  let finish!: (value: { token: string; refresh_token: string }) => void
  const response = new Promise<{ token: string; refresh_token: string }>((resolve) => {
    finish = resolve
  })
  const refresh = createRefreshCoordinator({
    storage,
    accessTokenKey: 'access',
    refreshTokenKey: 'refresh',
    request: async () => { calls += 1; return response },
    validateAccessToken: (token) => token.startsWith('valid-'),
    supersededError: () => new Error('superseded'),
  })

  const first = refresh()
  const second = refresh()
  assert.equal(calls, 1)
  finish({ token: 'valid-access', refresh_token: 'refresh-new' })
  assert.deepEqual(await Promise.all([first, second]), ['valid-access', 'valid-access'])
  assert.equal(storage.value('access'), 'valid-access')
  assert.equal(storage.value('refresh'), 'refresh-new')
})

test('rejects a refresh superseded by logout', async () => {
  const storage = memoryStorage({ refresh: 'refresh-old' })
  let finish!: (value: { token: string }) => void
  const response = new Promise<{ token: string }>((resolve) => { finish = resolve })
  const refresh = createRefreshCoordinator({
    storage,
    accessTokenKey: 'access',
    refreshTokenKey: 'refresh',
    request: async () => response,
    validateAccessToken: () => true,
    supersededError: () => new Error('superseded'),
  })

  const pending = refresh()
  storage.removeItem('refresh')
  markSessionChanged(storage)
  finish({ token: 'valid-access' })

  await assert.rejects(pending, /superseded/)
  assert.equal(storage.value('access'), undefined)
})

test('rejects invalid access tokens without storing them', async () => {
  const storage = memoryStorage({ refresh: 'refresh-old' })
  const refresh = createRefreshCoordinator({
    storage,
    accessTokenKey: 'access',
    refreshTokenKey: 'refresh',
    request: async () => ({ token: 'invalid' }),
    validateAccessToken: (token) => token.startsWith('valid-'),
    invalidAccessTokenError: () => new Error('invalid access'),
    supersededError: () => new Error('superseded'),
  })

  await assert.rejects(refresh(), /invalid access/)
  assert.equal(storage.value('access'), undefined)
})
