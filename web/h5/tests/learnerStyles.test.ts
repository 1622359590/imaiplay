import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const stylesheet = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8')

test('H5 learner stylesheet uses semantic variables instead of color literals', () => {
  assert.doesNotMatch(stylesheet, /#[0-9a-f]{3,8}\b|\brgba?\(|\bhsla?\(/i)
})

test('H5 learner controls retain 44px touch targets', () => {
  assert.match(stylesheet, /\.adm-button-primary\s*\{[^}]*min-height:\s*44px/s)
  assert.match(stylesheet, /\.icon-button\s*\{[^}]*width:\s*44px[^}]*height:\s*44px/s)
})

test('H5 learner stylesheet disables Clay movement for reduced motion', () => {
  assert.match(
    stylesheet,
    /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.adm-button-primary:hover[^}]*transform:\s*none/s,
  )
})
