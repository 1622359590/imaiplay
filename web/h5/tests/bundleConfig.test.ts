import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { loadConfigFromFile } from 'vite'

test('H5 build uses the shared vendor classifier without changing its public base or port', async () => {
  const configUrl = new URL('../vite.config.ts', import.meta.url)
  const source = readFileSync(configUrl, 'utf8')
  const loaded = await loadConfigFromFile(
    { command: 'build', mode: 'production' },
    configUrl.pathname,
  )

  assert.ok(loaded)
  assert.match(source, /import \{ manualChunks \} from '\.\.\/build\/vendorChunks'/)
  assert.equal(loaded.config.base, '/h5/')
  assert.equal(loaded.config.server?.port, 5175)
  const output = loaded.config.build?.rollupOptions?.output
  assert.ok(output && !Array.isArray(output))
  const classifier = output.manualChunks
  assert.equal(typeof classifier, 'function')
  if (typeof classifier !== 'function') return
  assert.equal(classifier('/repo/node_modules/react/index.js', {} as never), 'react-vendor')
  assert.equal(classifier('/repo/node_modules/react-router-dom/dist/index.js', {} as never), 'router-vendor')
  assert.equal(classifier('/repo/src/App.tsx', {} as never), undefined)
})

test('H5 router retains every learner page lazy import', () => {
  const source = readFileSync(new URL('../src/App.tsx', import.meta.url), 'utf8')

  for (const page of [
    'CourseDetailPage',
    'ForgotPasswordPage',
    'HomePage',
    'LoginPage',
    'LessonPlayerPage',
  ]) {
    assert.match(source, new RegExp(`const ${page} = lazy\\(\\(\\) => import\\('./pages/${page}'\\)`))
  }
})

test('H5 TypeScript build includes the shared classifier imported by its Vite config', () => {
  const tsconfig = JSON.parse(
    readFileSync(new URL('../tsconfig.node.json', import.meta.url), 'utf8'),
  ) as { include?: string[] }

  assert.ok(tsconfig.include?.includes('../build/vendorChunks.d.ts'))
})
