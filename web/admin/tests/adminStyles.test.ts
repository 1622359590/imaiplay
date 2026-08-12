import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const stylesPath = new URL('../src/styles.css', import.meta.url)
const layoutPath = new URL('../src/layout/AdminLayout.tsx', import.meta.url)

test('admin CSS defines the clay action and dashboard contracts while tables and forms stay flat', async () => {
  const styles = await readFile(stylesPath, 'utf8')

  assert.match(styles, /--admin-clay-surface:/)
  assert.match(styles, /--admin-clay-shadow:/)
  assert.match(styles, /--admin-clay-atmosphere:/)
  assert.match(styles, /\.ant-btn-primary[^}]*box-shadow:\s*0 4px 0 var\(--admin-clay-shadow\)/s)
  assert.match(styles, /\.stat-card[^}]*box-shadow:\s*0 4px 0 var\(--admin-clay-white-shadow\)/s)
  assert.match(styles, /\.ant-table-wrapper[^}]*box-shadow:\s*none/s)
  assert.match(styles, /\.ant-form[^}]*box-shadow:\s*none/s)
})

test('admin CSS no longer owns literal theme colors outside neutral fallbacks', async () => {
  const styles = await readFile(stylesPath, 'utf8')
  const withoutVariableDefaults = styles.replace(/--admin-[\w-]+:\s*(?:#[0-9a-fA-F]{3,8}|rgba?\([^)]*\));?/g, '')
  assert.doesNotMatch(withoutVariableDefaults, /#[0-9a-fA-F]{3,8}|rgba?\([^)]*\)/)
})

test('admin shell exposes active page context and the required responsive navigation widths', async () => {
  const [styles, layout] = await Promise.all([
    readFile(stylesPath, 'utf8'),
    readFile(layoutPath, 'utf8'),
  ])

  assert.match(layout, /<Sider width=\{220\} collapsedWidth=\{76\}/)
  assert.match(layout, /const activeMenuItem =/)
  assert.match(layout, /className="header-context"/)
  assert.match(styles, /@media \(max-width: 959px\)/)
  assert.match(styles, /\.admin-nav-drawer/)
})

test('admin shared surfaces include semantic upload, modal, picker, error, focus, and motion contracts', async () => {
  const styles = await readFile(stylesPath, 'utf8')

  for (const selector of [
    '.media-uploader',
    '.import-modal-stack',
    '.course-material-admin-row',
    '.official-course-picker-row',
    '.route-error-surface',
  ]) {
    assert.match(styles, new RegExp(selector.replace('.', '\\\.')))
  }
  assert.match(styles, /:focus-visible/)
  assert.match(styles, /@media \(prefers-reduced-motion: reduce\)/)
})

test('focus rings and focused fields use the readable foreground token for bright tenant colors', async () => {
  const styles = await readFile(stylesPath, 'utf8')

  assert.match(styles, /\.ant-input-affix-wrapper:focus,[\s\S]*?border-color:\s*var\(--admin-accent-foreground\)\s*!important/)
  assert.match(styles, /:focus-visible,[\s\S]*?outline:\s*3px solid var\(--admin-accent-foreground\)\s*!important/)
  assert.doesNotMatch(styles, /outline:\s*[^;]*var\(--tenant-focus\)/)
})

test('all three real dashboard metric surfaces use shallow white clay depth', async () => {
  const styles = await readFile(stylesPath, 'utf8')

  assert.match(
    styles,
    /\.stat-card,\s*\.instructor-metric-grid > \.ant-card,\s*\.station-metrics-card\s*\{[^}]*box-shadow:\s*0 4px 0 var\(--admin-clay-white-shadow\)/s,
  )
  assert.match(styles, /\.ant-table-wrapper[^}]*box-shadow:\s*none/s)
  assert.match(styles, /\.ant-form[^}]*box-shadow:\s*none/s)
})

test('admin authentication brand panel keeps restrained desktop and mobile clay depth', async () => {
  const styles = await readFile(stylesPath, 'utf8')

  assert.match(styles, /\.auth-brand-panel\s*\{[^}]*box-shadow:[^;]*4px 0 0 var\(--admin-clay-shadow\), 10px 0 20px var\(--admin-clay-atmosphere\)/s)
  assert.match(styles, /@media \(max-width: 959px\)[\s\S]*?\.auth-brand-panel\s*\{[^}]*box-shadow:\s*0 2px 0 var\(--admin-clay-shadow\), 0 7px 12px var\(--admin-clay-atmosphere\)/)
})
