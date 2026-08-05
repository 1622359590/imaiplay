import assert from 'node:assert/strict'
import test from 'node:test'
import { loadWithReload, type ReloadRuntime } from './lazyWithReload.ts'

function createRuntime(initialMarker: string | null = null) {
  let marker = initialMarker
  let reloads = 0

  const runtime: ReloadRuntime = {
    getReloadMarker: () => marker,
    setReloadMarker: () => { marker = '1' },
    clearReloadMarker: () => { marker = null },
    reload: () => { reloads += 1 },
  }

  return {
    runtime,
    marker: () => marker,
    reloads: () => reloads,
  }
}

test('successful import clears a previous recovery marker', async () => {
  const state = createRuntime('1')

  const module = await loadWithReload(async () => ({ default: 'tenant-page' }), state.runtime)

  assert.equal(module.default, 'tenant-page')
  assert.equal(state.marker(), null)
  assert.equal(state.reloads(), 0)
})

test('first dynamic import failure marks the session and reloads once', async () => {
  const state = createRuntime()

  void loadWithReload(
    async () => { throw new TypeError('Failed to fetch dynamically imported module: /assets/Tenants-old.js') },
    state.runtime,
  )
  await new Promise((resolve) => setImmediate(resolve))

  assert.equal(state.marker(), '1')
  assert.equal(state.reloads(), 1)
})

test('repeated dynamic import failure clears the marker and rejects', async () => {
  const state = createRuntime('1')
  const error = new TypeError('Failed to fetch dynamically imported module: /assets/Tenants-old.js')

  await assert.rejects(loadWithReload(async () => { throw error }, state.runtime), error)

  assert.equal(state.marker(), null)
  assert.equal(state.reloads(), 0)
})

test('unrelated import failure rejects without reloading', async () => {
  const state = createRuntime()
  const error = new Error('permission denied')

  await assert.rejects(loadWithReload(async () => { throw error }, state.runtime), error)

  assert.equal(state.marker(), null)
  assert.equal(state.reloads(), 0)
})
