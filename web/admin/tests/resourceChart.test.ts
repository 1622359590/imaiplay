import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
import { loadResourceChart } from '../src/utils/resourceChart.ts'

test('resource chart loading degrades to the textual legend when the chunk fails', async () => {
  assert.equal(await loadResourceChart(async () => { throw new Error('chunk unavailable') }), null)
  const module = { init: () => undefined }
  assert.equal(await loadResourceChart(async () => module), module)
})

test('ResourceDonut rebuilds the canvas chart when the active tenant theme changes', () => {
  const source = readFileSync(new URL('../src/components/ResourceDonut.tsx', import.meta.url), 'utf8')
  assert.match(source, /const adminTheme = useAdminTheme\(\)/)
  assert.match(source, /createAdminPalette\(adminTheme\.primaryColor\)/)
  assert.match(source, /const chartThemeVersion = resourceChartThemeKey\(chartTheme\)/)
  assert.match(source, /useEffect\([\s\S]*?chart\.setOption\([\s\S]*?\}, \[chartLibrary, chartThemeVersion, data, total\]\)/)
  assert.match(source, /return \(\) => \{[\s\S]*?chart\?\.dispose\(\)/)
})
