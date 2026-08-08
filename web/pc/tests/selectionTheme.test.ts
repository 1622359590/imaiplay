import assert from 'node:assert/strict'
import test from 'node:test'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'
import { applyLearnerPalette, createLearnerPalette } from '../src/theme/learnerPalette.ts'
import { readStyleBundle } from './styleSource.ts'

test('PC palette exposes independent persistent selection variables', () => {
  const properties = new Map<string, string>()
  const selectionColors: TenantSelectionColors = {
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  }
  applyLearnerPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement, createLearnerPalette('#3582E1'), selectionColors)
  assert.equal(properties.get('--tenant-selected-background'), '#FFF1F0')
  assert.equal(properties.get('--tenant-selected-text'), '#C5221F')
  assert.equal(properties.get('--tenant-selected-icon'), '#8C1D18')
  assert.equal(properties.get('--learner-accent'), '#3582e1')
})

test('PC persistent navigation and tabs consume selection variables', () => {
  const stylesheet = readStyleBundle(new URL('../src/styles.css', import.meta.url))
  for (const selector of [
    '.learner-top-nav-link.active',
    '.learner-filter-tabs .ant-tabs-tab-active',
    '.course-experience-tabs .ant-tabs-tab-active',
  ]) {
    const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    assert.match(stylesheet, new RegExp(`${escaped}[^\\{]*\\{[^}]*background:\\s*var\\(--tenant-selected-background\\)[^}]*color:\\s*var\\(--tenant-selected-text\\)`, 's'))
  }
  assert.match(
    stylesheet,
    /\.learner-filter-tabs \.ant-tabs-tab\.ant-tabs-tab-active \.ant-tabs-tab-btn[^\{]*\{[^}]*color:\s*var\(--tenant-selected-text\)\s*!important/s,
  )
  assert.match(
    stylesheet,
    /\.learner-filter-tabs \.ant-tabs-ink-bar[^\{]*\{[^}]*background:\s*var\(--tenant-selected-icon\)/s,
  )
  assert.match(stylesheet, /\.ant-btn-primary[^}]*\{[^}]*background:\s*var\(--learner-accent\)/s)
  assert.match(stylesheet, /\.ant-progress-bg[^}]*\{[^}]*background:\s*var\(--learner-accent\)/s)
})
