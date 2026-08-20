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

test('H5 motivation popup is safe-area aware, touchable, themed, and motion-safe', () => {
  assert.match(stylesheet, /\.learner-motivation-popup\s*\{[^}]*env\(safe-area-inset-bottom\)/s)
  assert.match(stylesheet, /\.learner-motivation-popup \.learner-motivation-primary\s*\{[^}]*min-height:\s*44px/s)
  assert.match(stylesheet, /\.learner-motivation-metrics\s*\{[^}]*grid-template-columns:\s*repeat\(2,/s)
  assert.match(stylesheet, /@media \(prefers-reduced-motion: reduce\)[\s\S]*\.learner-motivation-popup[^}]*transform:\s*none/s)
  const promptRules = stylesheet.match(/\.learner-motivation[^}]+\{[^}]*\}/gs) ?? []
  assert.ok(promptRules.length > 0)
  for (const rule of promptRules) {
    assert.doesNotMatch(rule, /#[0-9a-fA-F]{3,8}\b|rgba?\(|hsla?\(/)
  }
})

test('H5 learner stylesheet disables Clay movement for reduced motion', () => {
  const reducedMotion = stylesheet.slice(stylesheet.indexOf('@media (prefers-reduced-motion: reduce)'))
  for (const selector of [
    '.adm-button-primary:hover',
    '.header-action:hover',
    '.header-action:active',
    '.course-filters button:hover',
    '.course-filters button:active',
  ]) {
    assert.ok(reducedMotion.includes(selector), `${selector} must be included in reduced motion`)
  }
  assert.match(reducedMotion, /transform:\s*none/s)
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

test('H5 learner header, continue, and filter controls share the 6-4-0 Clay press contract', () => {
  for (const selector of [
    '\\.header-action',
    '\\.continue-action\\.adm-button-primary',
    '\\.course-filters button',
  ]) {
    assert.match(
      stylesheet,
      new RegExp(`${selector}\\s*\\{[^}]*0\\s+6px\\s+0\\s+0\\s+var\\(--learner-clay-(?:white-)?shadow\\)`, 's'),
      `${selector} resting depth`,
    )
    assert.match(
      stylesheet,
      new RegExp(`${selector}:hover\\s*\\{[^}]*transform:\\s*translateY\\(2px\\)[^}]*0\\s+4px\\s+0\\s+0\\s+var\\(--learner-clay-(?:white-)?shadow\\)`, 's'),
      `${selector} hover depth`,
    )
    assert.match(
      stylesheet,
      new RegExp(`${selector}:active\\s*\\{[^}]*transform:\\s*translateY\\(6px\\)[^}]*0\\s+0\\s+0\\s+0\\s+var\\(--learner-clay-(?:white-)?shadow\\)`, 's'),
      `${selector} pressed depth`,
    )
  }
})
