import assert from 'node:assert/strict'
import test from 'node:test'
import { loadConfigFromFile } from 'vite'

test('Admin build delegates stable vendor groups and leaves cyclic Ant runtime modules to Rollup', async () => {
  const loaded = await loadConfigFromFile(
    { command: 'build', mode: 'production' },
    new URL('../vite.config.ts', import.meta.url).pathname,
  )
  assert.ok(loaded)
  const adminConfig = loaded.config
  const classifier = adminConfig.build?.rollupOptions?.output?.manualChunks

  assert.equal(adminConfig.base, '/admin/')
  assert.equal(adminConfig.server?.port, 5173)
  assert.equal(adminConfig.build?.chunkSizeWarningLimit, 500)
  assert.equal(typeof classifier, 'function')
  assert.equal(classifier('/repo/node_modules/react/index.js'), 'react-vendor')
  assert.equal(classifier('/repo/node_modules/react-router-dom/dist/index.js'), 'router-vendor')
  assert.equal(classifier('/repo/node_modules/antd/es/table/style/index.js'), 'antd-styles')
  assert.equal(classifier('/repo/node_modules/@ant-design/icons/es/icons/UserOutlined.js'), undefined)
  assert.equal(classifier('/repo/node_modules/rc-table/es/index.js'), undefined)
  assert.equal(classifier('/repo/node_modules/antd/es/table/index.js'), undefined)
  assert.equal(classifier('/repo/src/App.tsx'), undefined)
})
