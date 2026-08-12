import assert from 'node:assert/strict'
import test from 'node:test'

import { manualChunks, vendorChunkFor } from './vendorChunks.js'

test('assigns dependency families to stable vendor chunks', () => {
  assert.equal(vendorChunkFor('/repo/node_modules/react/index.js'), 'react-vendor')
  assert.equal(vendorChunkFor('/repo/node_modules/react-router-dom/dist/index.js'), 'router-vendor')
  assert.equal(vendorChunkFor('/repo/node_modules/@ant-design/icons/es/icons/UserOutlined.js'), 'antd-icons')
  assert.equal(vendorChunkFor('/repo/node_modules/@rc-component/trigger/es/index.js'), 'antd-primitives')
  assert.equal(vendorChunkFor('/repo/node_modules/rc-table/es/index.js'), 'antd-primitives')
  assert.equal(vendorChunkFor('/repo/node_modules/antd/es/table/style/index.js'), 'antd-styles')
  assert.equal(vendorChunkFor('/repo/node_modules/antd/es/table/index.js'), 'antd-framework')
  assert.equal(vendorChunkFor('/repo/node_modules/@reduxjs/toolkit/dist/index.js'), 'state-vendor')
  assert.equal(vendorChunkFor('/repo/node_modules/axios/index.js'), 'transport-vendor')
  assert.equal(vendorChunkFor('/repo/src/App.tsx'), undefined)
})

test('matches specialized Ant Design families before broad framework matching', () => {
  assert.equal(vendorChunkFor('/repo/node_modules/@ant-design/icons-svg/es/asn/UserOutlined.js'), 'antd-icons')
  assert.equal(vendorChunkFor('/repo/node_modules/@rc-component/color-picker/es/index.js'), 'antd-primitives')
  assert.equal(vendorChunkFor('/repo/node_modules/@ant-design/colors/es/index.js'), 'antd-framework')
  assert.equal(vendorChunkFor('/repo/node_modules/antd/es/button/style/index.js'), 'antd-styles')
})

test('normalizes Windows module paths before classifying dependencies', () => {
  assert.equal(vendorChunkFor('C:\\repo\\node_modules\\rc-trigger\\es\\index.js'), 'antd-primitives')
  assert.equal(manualChunks('C:\\repo\\node_modules\\react-dom\\client.js'), 'react-vendor')
  assert.equal(manualChunks('C:\\repo\\src\\App.tsx'), undefined)
})
