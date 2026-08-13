import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { loadConfigFromFile } from 'vite'

test('PC build delegates stable vendors while leaving mutually dependent Ant modules to Rollup', async () => {
  const configUrl = new URL('../vite.config.ts', import.meta.url)
  const source = readFileSync(configUrl, 'utf8')
  const loaded = await loadConfigFromFile(
    { command: 'build', mode: 'production' },
    configUrl.pathname,
  )

  assert.ok(loaded)
  assert.match(source, /from '\.\.\/build\/vendorChunks'/)
  assert.equal(loaded.config.base, '/pc/')
  assert.equal(loaded.config.server?.port, 5174)
  const output = loaded.config.build?.rollupOptions?.output
  assert.ok(output && !Array.isArray(output))
  const classifier = output.manualChunks
  assert.equal(typeof classifier, 'function')
  if (typeof classifier !== 'function') return
  assert.equal(classifier('/repo/node_modules/react/index.js', {} as never), 'react-vendor')
  assert.equal(classifier('/repo/node_modules/antd/es/table/index.js', {} as never), undefined)
  assert.equal(classifier('/repo/node_modules/rc-table/es/index.js', {} as never), undefined)
  assert.equal(classifier('/repo/node_modules/@ant-design/icons/es/icons/UserOutlined.js', {} as never), undefined)
  assert.equal(classifier('/repo/src/router.tsx', {} as never), undefined)
})

test('PC router retains every learner page lazy import', () => {
  const source = readFileSync(new URL('../src/router.tsx', import.meta.url), 'utf8')

  for (const page of [
    'CourseDetailPage',
    'CourseListPage',
    'HomePage',
    'LoginPage',
    'OrganizationSelectPage',
    'LessonPlayerPage',
    'RecentPage',
  ]) {
    assert.match(source, new RegExp(`const ${page} = lazy\\(\\(\\) => import\\('./pages/${page}'\\)`))
  }
})
