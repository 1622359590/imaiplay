import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { deriveClayColors } from '@imaiplay/shared/theme/tenantTheme'
import {
  hasLowSelectionContrast,
  syncSelectionColorsForPrimaryChange,
} from '../src/theme/selectionSettings.ts'

test('primary changes update only selection colors that still match recommendations', () => {
  assert.deepEqual(syncSelectionColorsForPrimaryChange(
    '#4F46E5',
    '#3582E1',
    {
      selected_background_color: '#4F46E5',
      selected_text_color: '#FFFFFF',
      selected_icon_color: '#8C1D18',
    },
  ), {
    selected_background_color: '#3582E1',
    selected_text_color: '#000000',
    selected_icon_color: '#8C1D18',
  })
})

test('primary changes leave every custom selection color untouched', () => {
  const custom = {
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  }

  assert.deepEqual(
    syncSelectionColorsForPrimaryChange('#4F46E5', '#3582E1', custom),
    custom,
  )
})

test('low contrast selection colors warn without invalidating the values', () => {
  assert.equal(hasLowSelectionContrast({
    selected_background_color: '#FFFFFF',
    selected_text_color: '#F5F5F5',
    selected_icon_color: '#EEEEEE',
  }), true)
  assert.equal(hasLowSelectionContrast({
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  }), false)
})

test('theme settings exposes four semantic sections and every cross-platform preview hook', () => {
  const source = readFileSync(
    new URL('../src/pages/ThemeSettings.tsx', import.meta.url),
    'utf8',
  )

  for (const className of [
    'theme-section-brand-basics',
    'theme-section-color-system',
    'theme-section-brand-assets',
    'theme-section-live-preview',
    'theme-preview-admin-nav',
    'theme-preview-primary-button',
    'theme-preview-learner-hero',
    'theme-preview-progress',
    'theme-preview-clay-contact',
  ]) {
    assert.match(source, new RegExp(`className=["'{][^\\n]*${className}`))
  }
})

test('local primary state derives every live preview surface without reading root theme state', async () => {
  const previewModule = await import('../src/theme/themePreview.ts').catch(() => undefined)
  assert.ok(previewModule, 'theme preview style helper must be independently testable')
  const primary = '#3582E1'
  const selection = {
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  }
  const clay = deriveClayColors(primary)
  const style = previewModule.createThemePreviewStyle(primary, selection)

  assert.equal(style['--admin-preview-selected-background'], selection.selected_background_color)
  assert.equal(style['--admin-preview-selected-text'], selection.selected_text_color)
  assert.equal(style['--admin-preview-selected-icon'], selection.selected_icon_color)
  assert.equal(style['--admin-preview-button-surface'], clay.surface)
  assert.equal(style['--admin-preview-hero-surface'], clay.surface)
  assert.equal(style['--admin-preview-progress-surface'], clay.surface)
  assert.equal(style['--admin-preview-contact-shadow'], clay.shadow)
})
