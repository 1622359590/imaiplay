import assert from 'node:assert/strict'
import { readFile, readdir } from 'node:fs/promises'
import { relative } from 'node:path'
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

test('every Admin source color is supplied by palette defaults or CSS variables', async () => {
  const srcRoot = new URL('../src/', import.meta.url)
  const files: string[] = []
  const collect = async (directory: URL) => {
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      const url = new URL(entry.name + (entry.isDirectory() ? '/' : ''), directory)
      if (entry.isDirectory()) await collect(url)
      else if (/\.(?:ts|tsx|css)$/.test(entry.name)) files.push(url.pathname)
    }
  }
  await collect(srcRoot)

  const violations: string[] = []
  for (const file of files) {
    const relativePath = relative(srcRoot.pathname, file)
    if (relativePath === 'theme/adminPalette.ts') continue
    let source = await readFile(file, 'utf8')
    if (relativePath === 'styles.css') source = source.replace(/:root\s*\{[\s\S]*?\n\}/, '')
    if (/#[0-9a-fA-F]{3,8}|rgba?\(/.test(source)) violations.push(relativePath)
  }
  assert.deepEqual(violations, [])
})

test('all Task 8 data pages use the shared page width contract', async () => {
  const pages = [
    'Courses.tsx',
    'OfficialCourses.tsx',
    'CourseCategories.tsx',
    'Users.tsx',
    'Resources.tsx',
    'ResourceCategories.tsx',
    'Tenants.tsx',
    'Plans.tsx',
    'AuditLogs.tsx',
  ]
  for (const page of pages) {
    const source = await readFile(new URL(`../src/pages/${page}`, import.meta.url), 'utf8')
    assert.match(source, /className="[^"]*admin-page/)
  }
})

test('reviewed Admin surfaces keep semantic clay and responsive layout contracts', async () => {
  const styles = await readFile(stylesPath, 'utf8')
  assert.match(styles, /\.domain-settings-grid\s*\{[^}]*grid-template-columns:\s*repeat\(2,/s)
  assert.match(styles, /\.quick-action-icon\s*\{[^}]*background:\s*var\(--admin-accent-light\)[^}]*0 2px 0 var\(--admin-clay-shadow\)/s)
  assert.match(styles, /@media \(max-width: 959px\)[\s\S]*?\.domain-settings-grid[^}]*grid-template-columns:\s*1fr/)
  assert.match(styles, /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.theme-preview-primary-button[^}]*transform:\s*none\s*!important/)
})

test('plan editor grid resolves to two desktop columns and one mobile column', async () => {
  const [styles, plans] = await Promise.all([
    readFile(stylesPath, 'utf8'),
    readFile(new URL('../src/pages/Plans.tsx', import.meta.url), 'utf8'),
  ])
  const mobileStart = styles.indexOf('@media (max-width: 620px)')
  assert.ok(mobileStart > 0)
  const desktopStyles = styles.slice(0, mobileStart)
  const mobileStyles = styles.slice(mobileStart, styles.indexOf('@media', mobileStart + 1))

  assert.match(plans, /className="admin-modal-form plan-editor-form"[\s\S]*?className="form-grid form-grid-two"/)
  assert.match(desktopStyles, /\.form-grid-two\s*\{[^}]*grid-template-columns:\s*repeat\(2,/s)
  assert.doesNotMatch(desktopStyles, /\.plan-editor-form\s+\.form-grid-two\s*\{/)
  assert.match(mobileStyles, /\.form-grid-two\s*\{[^}]*grid-template-columns:\s*1fr/s)
})
