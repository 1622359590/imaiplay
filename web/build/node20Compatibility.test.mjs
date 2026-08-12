import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { spawnSync } from 'node:child_process'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

import { inspectBundle } from './bundleBudget.js'
import { vendorChunkFor } from './vendorChunks.js'

test('production build tools load and execute without TypeScript loader support', (t) => {
  const root = mkdtempSync(path.join(tmpdir(), 'imaiplay-node20-build-tools-'))
  t.after(() => rmSync(root, { force: true, recursive: true }))

  for (const app of ['admin', 'pc', 'h5']) {
    mkdirSync(path.join(root, app, 'dist', 'assets'), { recursive: true })
  }

  assert.equal(vendorChunkFor('C:\\repo\\node_modules\\react\\index.js'), 'react-vendor')
  assert.deepEqual(inspectBundle(root).oversized, [])

  const result = spawnSync(
    process.execPath,
    [fileURLToPath(new URL('./bundleBudget.js', import.meta.url))],
    { cwd: root, encoding: 'utf8' },
  )

  assert.equal(result.status, 0, result.stderr)
  assert.match(result.stdout, /JavaScript bundle budget: 500,000 B raw per chunk/)
})
