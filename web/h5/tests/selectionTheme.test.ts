import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { applyLearnerPalette } from '../src/theme/learnerPalette.ts'

test('H5 palette exposes tenant primary and independent selected colors', () => {
  const properties = new Map<string, string>()
  applyLearnerPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement, '#3582E1', {
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  })
  assert.equal(properties.get('--learner-accent'), '#3582E1')
  assert.equal(properties.get('--tenant-selected-background'), '#FFF1F0')
  assert.equal(properties.get('--tenant-selected-text'), '#C5221F')
  assert.equal(properties.get('--tenant-selected-icon'), '#8C1D18')
})

test('H5 TabBar uses selected colors while primary buttons keep the tenant accent', () => {
  const stylesheet = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8')
  assert.match(stylesheet, /\.app-tabbar\s+\.adm-tab-bar-item-active\s*\{[^}]*background:\s*var\(--tenant-selected-background\)[^}]*color:\s*var\(--tenant-selected-text\)/s)
  assert.match(stylesheet, /\.app-tabbar\s+\.adm-tab-bar-item-active\s+\.adm-tab-bar-item-icon\s*\{[^}]*color:\s*var\(--tenant-selected-icon\)/s)
  assert.match(stylesheet, /\.adm-button-primary[^}]*\{[^}]*background:\s*var\(--learner-accent\)\s*!important/s)
})
