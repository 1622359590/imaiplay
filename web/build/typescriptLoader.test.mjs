import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import test from 'node:test'
import { pathToFileURL } from 'node:url'

import { load, resolve } from './typescriptLoader.mjs'

test('load delegates a foreign node_modules TypeScript package to Node', async (t) => {
  const root = mkdtempSync(path.join(tmpdir(), 'imaiplay-loader-foreign-package-'))
  t.after(() => rmSync(root, { force: true, recursive: true }))
  const modulePath = path.join(root, 'node_modules', 'foreign-package', 'index.ts')
  mkdirSync(path.dirname(modulePath), { recursive: true })
  writeFileSync(modulePath, 'export const foreignValue: number = 1')
  const expected = { format: 'module', source: 'delegated by Node' }
  const calls = []

  const result = await load(pathToFileURL(modulePath).href, {}, async (url, context) => {
    calls.push({ context, url })
    return expected
  })

  assert.equal(result, expected)
  assert.deepEqual(calls, [{ context: {}, url: pathToFileURL(modulePath).href }])
})

test('load delegates TypeScript files outside the web workspace to Node', async (t) => {
  const root = mkdtempSync(path.join(tmpdir(), 'imaiplay-loader-outside-workspace-'))
  t.after(() => rmSync(root, { force: true, recursive: true }))
  const modulePath = path.join(root, 'external.ts')
  writeFileSync(modulePath, 'export const externalValue: number = 1')
  let delegated = false

  await load(pathToFileURL(modulePath).href, {}, async () => {
    delegated = true
    return { format: 'module', source: 'delegated by Node' }
  })

  assert.equal(delegated, true)
})

test('resolve does not append TypeScript extensions for imports from node_modules', async () => {
  const calls = []
  const parentURL = new URL('../node_modules/foreign-package/index.js', import.meta.url).href
  const missing = Object.assign(new Error('missing'), { code: 'ERR_MODULE_NOT_FOUND' })

  await assert.rejects(
    resolve('./dependency', { parentURL }, async (specifier, context) => {
      calls.push({ context, specifier })
      throw missing
    }),
    (error) => error === missing,
  )

  assert.deepEqual(calls, [{ context: { parentURL }, specifier: './dependency' }])
})
