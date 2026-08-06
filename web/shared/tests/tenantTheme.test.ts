import assert from 'node:assert/strict'
import test from 'node:test'
import { normalizePrimaryColor } from '../src/theme/tenantTheme.ts'

test('normalizes valid tenant colors', () => {
  assert.equal(normalizePrimaryColor('  #ab12ef ', '#FF5156'), '#AB12EF')
  assert.equal(normalizePrimaryColor('#ABCDEF', '#FF5156'), '#ABCDEF')
})

test('uses the supplied fallback for invalid or empty colors', () => {
  assert.equal(normalizePrimaryColor('#abcd', '#FF5156'), '#FF5156')
  assert.equal(normalizePrimaryColor('', '#FF5156'), '#FF5156')
})
