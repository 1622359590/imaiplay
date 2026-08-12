import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

import { inspectBundle } from './bundleBudget.js'

const applications = ['admin', 'pc', 'h5']

function createBundleRoot(t) {
  const root = mkdtempSync(path.join(tmpdir(), 'imaiplay-bundle-budget-'))
  t.after(() => rmSync(root, { force: true, recursive: true }))

  for (const app of applications) {
    mkdirSync(path.join(root, app, 'dist', 'assets'), { recursive: true })
  }

  return root
}

test('accepts JavaScript chunks at the limit and reports raw and gzip metrics', (t) => {
  const root = createBundleRoot(t)
  writeFileSync(path.join(root, 'admin', 'dist', 'assets', 'boundary.js'), Buffer.alloc(500_000))
  writeFileSync(path.join(root, 'admin', 'dist', 'assets', 'small.js'), 'export default 1')

  const report = inspectBundle(root, 500_000)
  const admin = report.applications.find((application) => application.app === 'admin')

  assert.deepEqual(report.oversized, [])
  assert.equal(admin?.chunkCount, 2)
  assert.equal(admin?.totalRawBytes, 500_016)
  assert.equal(admin?.maximum?.app, 'admin')
  assert.equal(admin?.maximum?.file, 'boundary.js')
  assert.equal(admin?.maximum?.rawBytes, 500_000)
  assert.ok((admin?.maximum?.gzipBytes ?? 0) > 0)
  assert.equal(
    admin?.totalGzipBytes,
    admin?.assets.reduce((total, asset) => total + asset.gzipBytes, 0),
  )
})

test('reports a 500001-byte JavaScript chunk with its application, file, size, and limit', (t) => {
  const root = createBundleRoot(t)
  writeFileSync(path.join(root, 'pc', 'dist', 'assets', 'oversized.js'), Buffer.alloc(500_001))

  const report = inspectBundle(root, 500_000)

  assert.deepEqual(report.oversized, [
    {
      app: 'pc',
      file: 'oversized.js',
      size: 500_001,
      limit: 500_000,
    },
  ])
})

test('ignores CSS and source-map files when calculating JavaScript metrics', (t) => {
  const root = createBundleRoot(t)
  const assets = path.join(root, 'h5', 'dist', 'assets')
  writeFileSync(path.join(assets, 'app.js'), '1234')
  writeFileSync(path.join(assets, 'app.css'), Buffer.alloc(700_000))
  writeFileSync(path.join(assets, 'app.js.map'), Buffer.alloc(700_000))

  const h5 = inspectBundle(root).applications.find((application) => application.app === 'h5')

  assert.equal(h5?.chunkCount, 1)
  assert.equal(h5?.totalRawBytes, 4)
  assert.deepEqual(h5?.assets.map((asset) => asset.file), ['app.js'])
})

test('fails with an actionable error when an application assets directory is missing', (t) => {
  const root = createBundleRoot(t)
  rmSync(path.join(root, 'pc', 'dist', 'assets'), { recursive: true })

  assert.throws(
    () => inspectBundle(root),
    /Missing bundle assets directory for pc: .*pc\/dist\/assets.*npm run build:all/,
  )
})

test('CLI exits nonzero and names every oversized application asset', (t) => {
  const root = createBundleRoot(t)
  writeFileSync(path.join(root, 'pc', 'dist', 'assets', 'oversized.js'), Buffer.alloc(500_001))

  const result = spawnSync(
    process.execPath,
    [fileURLToPath(new URL('./bundleBudget.js', import.meta.url))],
    { cwd: root, encoding: 'utf8' },
  )

  assert.equal(result.status, 1)
  assert.match(result.stderr, /Bundle budget exceeded/)
  assert.match(result.stderr, /pc\/oversized\.js: 500,001 B exceeds 500,000 B/)
})
