import assert from 'node:assert/strict'
import test from 'node:test'
import { loadResourceChart } from '../src/utils/resourceChart.ts'

test('resource chart loading degrades to the textual legend when the chunk fails', async () => {
  assert.equal(await loadResourceChart(async () => { throw new Error('chunk unavailable') }), null)
  const module = { init: () => undefined }
  assert.equal(await loadResourceChart(async () => module), module)
})
