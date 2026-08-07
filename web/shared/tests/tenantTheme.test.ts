import assert from 'node:assert/strict'
import test from 'node:test'
import {
  contrastRatio,
  normalizePrimaryColor,
  normalizeSelectionColors,
  recommendedSelectionColors,
} from '../src/theme/tenantTheme.ts'

test('normalizes valid tenant colors', () => {
  assert.equal(normalizePrimaryColor('  #ab12ef ', '#FF5156'), '#AB12EF')
  assert.equal(normalizePrimaryColor('#ABCDEF', '#FF5156'), '#ABCDEF')
})

test('uses the supplied fallback for invalid or empty colors', () => {
  assert.equal(normalizePrimaryColor('#abcd', '#FF5156'), '#FF5156')
  assert.equal(normalizePrimaryColor('', '#FF5156'), '#FF5156')
})

test('recommends a readable selected state from the primary color', () => {
  assert.deepEqual(recommendedSelectionColors('#3582E1'), {
    selected_background_color: '#3582E1',
    selected_text_color: '#000000',
    selected_icon_color: '#000000',
  })
})

test('normalizes independent selected state colors', () => {
  assert.deepEqual(normalizeSelectionColors({
    primary_color: '#3582E1',
    selected_background_color: '#fff1f0',
    selected_text_color: '#c5221f',
    selected_icon_color: '#8c1d18',
  }, '#4F46E5'), {
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  })
})

test('falls back each invalid selected color independently', () => {
  assert.deepEqual(normalizeSelectionColors({
    primary_color: '#3582E1',
    selected_background_color: 'blue',
    selected_text_color: '#c5221f',
    selected_icon_color: '#123',
  }, '#4F46E5'), {
    selected_background_color: '#3582E1',
    selected_text_color: '#C5221F',
    selected_icon_color: '#C5221F',
  })
})

test('calculates WCAG contrast ratios', () => {
  assert.equal(contrastRatio('#FFFFFF', '#FFFFFF'), 1)
  assert.equal(contrastRatio('#000000', '#FFFFFF'), 21)
})
