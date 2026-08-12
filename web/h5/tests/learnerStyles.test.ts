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

test('H5 lesson outline applies persisted selection colors through the real current-item cascade', () => {
  assert.match(
    stylesheet,
    /\.lesson-outline-item\.is-current\s*\{[^}]*color:\s*var\(--tenant-selected-text\)[^}]*background:\s*var\(--tenant-selected-background\)/s,
  )
  assert.match(
    stylesheet,
    /\.lesson-outline-item\.is-current \.outline-copy strong,[\s\S]*?\.lesson-outline-item\.is-current \.outline-copy small\s*\{[^}]*color:\s*var\(--tenant-selected-text\)/s,
  )
  assert.match(
    stylesheet,
    /\.lesson-outline-item\.is-current \.outline-icon,[\s\S]*?\.lesson-outline-item\.is-current \.outline-play\s*\{[^}]*color:\s*var\(--tenant-selected-icon\)[^}]*background:\s*var\(--tenant-selected-background\)/s,
  )
})

test('H5 lesson outline has tactile press feedback that reduced motion cancels', () => {
  assert.match(stylesheet, /\.lesson-outline-item:active\s*\{[^}]*transform:\s*translateY\(/s)
  assert.match(
    stylesheet,
    /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.lesson-outline-item:active[^}]*transform:\s*none/s,
  )
})

test('H5 player navigation has an explicit neutral Clay contact layer', () => {
  assert.match(
    stylesheet,
    /\.player-page > \.adm-nav-bar\s*\{[^}]*0 4px 0 0 var\(--learner-clay-white-shadow\)/s,
  )
})
