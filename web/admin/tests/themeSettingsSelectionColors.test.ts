import assert from 'node:assert/strict'
import test from 'node:test'
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
