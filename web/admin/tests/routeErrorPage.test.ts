import assert from 'node:assert/strict'
import test from 'node:test'
import { routeErrorPresentation } from '../src/components/routeErrorModel.ts'

test('route errors expose recovery guidance instead of technical details', () => {
  const result = routeErrorPresentation(new Error('Failed to fetch dynamically imported module: Tenants-abc.js'))

  assert.deepEqual(result, {
    title: '页面资源加载失败',
    description: '系统可能刚刚完成更新，或当前网络暂时不可用。请刷新页面后重试。',
  })
  assert.equal(JSON.stringify(result).includes('Tenants-abc.js'), false)
})
