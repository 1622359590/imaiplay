import assert from 'node:assert/strict'
import test from 'node:test'
import { existsSync, readFileSync } from 'node:fs'
import { loadResourceChart } from '../src/utils/resourceChart.ts'

test('resource chart exposes the modular ECharts pie runtime to the loader boundary', async () => {
  const library = await loadResourceChart(async () => (
    await import('../src/utils/resourceChartRuntime.ts')
  ).default)

  assert.ok(library)
  assert.equal(typeof library.init, 'function')
})

test('resource chart runtime imports only the ECharts modules needed by the donut', () => {
  const loaderSource = readFileSync(new URL('../src/utils/resourceChart.ts', import.meta.url), 'utf8')
  const runtimeUrl = new URL('../src/utils/resourceChartRuntime.ts', import.meta.url)

  assert.match(loaderSource, /import\(['"]\.\/resourceChartRuntime['"]\)/)
  assert.doesNotMatch(loaderSource, /await import\(['"]echarts(?:\/[^'"]+)?['"]\)/)
  assert.equal(existsSync(runtimeUrl), true, 'the async chart runtime boundary must exist')

  const runtimeSource = readFileSync(runtimeUrl, 'utf8')
  assert.match(runtimeSource, /import \* as echarts from ['"]echarts\/core['"]/)
  assert.match(runtimeSource, /import \{ PieChart \} from ['"]echarts\/charts['"]/)
  assert.match(runtimeSource, /import \{ LegendComponent, TitleComponent, TooltipComponent \} from ['"]echarts\/components['"]/)
  assert.match(runtimeSource, /import \{ CanvasRenderer \} from ['"]echarts\/renderers['"]/)
  assert.match(runtimeSource, /echarts\.use\(\[PieChart, LegendComponent, TitleComponent, TooltipComponent, CanvasRenderer\]\)/)
  assert.doesNotMatch(runtimeSource, /from ['"]echarts['"]/)
})

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
