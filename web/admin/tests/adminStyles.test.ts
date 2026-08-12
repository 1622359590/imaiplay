import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const stylesPath = new URL('../src/styles.css', import.meta.url)

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
