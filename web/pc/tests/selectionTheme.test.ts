import assert from 'node:assert/strict'
import test from 'node:test'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'
import { applyLearnerPalette, createLearnerPalette } from '../src/theme/learnerPalette.ts'
import { readStyleBundle } from './styleSource.ts'

function selectorSpecificity(selector: string): [number, number, number] {
  const ids = selector.match(/#[\w-]+/g)?.length ?? 0
  const classes = selector.match(/\.[\w-]+|\[[^\]]+\]|:(?!:)[\w-]+/g)?.length ?? 0
  const elements = selector
    .replace(/#[\w-]+|\.[\w-]+|\[[^\]]+\]|:{1,2}[\w-]+(?:\([^)]*\))?/g, '')
    .match(/(?:^|[>+~\s,])\s*[a-z][\w-]*/gi)?.length ?? 0
  return [ids, classes, elements]
}

function compareSpecificity(left: [number, number, number], right: [number, number, number]) {
  return left[0] - right[0] || left[1] - right[1] || left[2] - right[2]
}

function resolvedLearnerInkBarHeight(stylesheet: string) {
  const candidates = [{
    selector: '.ant-tabs-top > .ant-tabs-nav .ant-tabs-ink-bar',
    value: '2px',
    order: 0,
  }]
  const rulePattern = /([^{}]+)\{([^{}]*)\}/g
  let match: RegExpExecArray | null
  let order = 1
  while ((match = rulePattern.exec(stylesheet))) {
    const declaration = match[2].match(/(?:^|;)\s*height:\s*([^;!}]+)/)
    if (!declaration) continue
    for (const selector of match[1].split(',')) {
      if (selector.includes('.learner-filter-tabs') && selector.includes('.ant-tabs-ink-bar')) {
        candidates.push({ selector: selector.trim(), value: declaration[1].trim(), order })
      }
    }
    order += 1
  }
  candidates.sort((left, right) => {
    const specificity = compareSpecificity(selectorSpecificity(left.selector), selectorSpecificity(right.selector))
    return specificity || left.order - right.order
  })
  return candidates.at(-1)?.value
}

function resolvedCourseTabTextColor(stylesheet: string) {
  const candidates = [{
    selector: '.css-hash.ant-tabs .ant-tabs-tab.ant-tabs-tab-active .ant-tabs-tab-btn',
    value: 'var(--learner-accent)',
    important: false,
    order: 0,
  }]
  const rulePattern = /([^{}]+)\{([^{}]*)\}/g
  let match: RegExpExecArray | null
  let order = 1
  while ((match = rulePattern.exec(stylesheet))) {
    const declaration = match[2].match(/(?:^|;)\s*color:\s*([^;!}]+)(\s*!important)?/)
    if (!declaration) continue
    for (const selector of match[1].split(',')) {
      if (selector.includes('.course-experience-tabs') && selector.includes('.ant-tabs-tab-btn')) {
        candidates.push({
          selector: selector.trim(), value: declaration[1].trim(),
          important: Boolean(declaration[2]), order,
        })
      }
    }
    order += 1
  }
  candidates.sort((left, right) => {
    if (left.important !== right.important) return Number(left.important) - Number(right.important)
    const specificity = compareSpecificity(selectorSpecificity(left.selector), selectorSpecificity(right.selector))
    return specificity || left.order - right.order
  })
  return candidates.at(-1)?.value
}

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
  assert.equal(resolvedCourseTabTextColor(stylesheet), 'var(--tenant-selected-text)')
  for (const selector of [
    '.learner-top-nav-link.active',
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
    /\.learner-filter-tabs[^\{]* \.ant-tabs-ink-bar[^\{]*\{[^}]*background:\s*var\(--tenant-selected-background\)/s,
  )
  assert.match(stylesheet, /\.ant-btn-primary[^}]*\{[^}]*background:\s*var\(--learner-accent\)/s)
  assert.match(stylesheet, /\.ant-progress-bg[^}]*\{[^}]*background:\s*var\(--learner-accent\)/s)
})

test('learner selected controls share one readable geometry', () => {
  const stylesheet = readStyleBundle(new URL('../src/styles.css', import.meta.url))

  assert.match(
    stylesheet,
    /\.learner-top-nav-link\s*\{[^}]*height:\s*40px[^}]*padding:\s*0\s+18px[^}]*border-radius:\s*10px/s,
  )
  assert.match(
    stylesheet,
    /\.learner-filter-tabs \.ant-tabs-tab\s*\{[^}]*min-height:\s*40px[^}]*padding:\s*0\s+18px[^}]*border-radius:\s*10px/s,
  )
  assert.match(
    stylesheet,
    /\.course-experience-tabs \.ant-tabs-tab\s*\{[^}]*min-width:\s*96px[^}]*min-height:\s*40px[^}]*padding:\s*0\s+20px[^}]*border-radius:\s*10px/s,
  )
  assert.equal(resolvedLearnerInkBarHeight(stylesheet), '40px')
  assert.match(
    stylesheet,
    /\.learner-filter-tabs[^\{]* \.ant-tabs-ink-bar[^\{]*\{[^}]*height:\s*40px[^}]*border-radius:\s*10px/s,
  )
})

test('learner filter selection keeps its animated capsule behavior', () => {
  const stylesheet = readStyleBundle(new URL('../src/styles.css', import.meta.url))

  assert.match(
    stylesheet,
    /\.learner-filter-tabs \.ant-tabs-ink-bar-animated[^\{]*\{[^}]*transition:[^}]*240ms[^}]*cubic-bezier\(0\.22,\s*1,\s*0\.36,\s*1\)/s,
  )
  assert.match(
    stylesheet,
    /@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*\.learner-filter-tabs \.ant-tabs-ink-bar-animated[^\{]*\{[^}]*transition:\s*none/s,
  )
  assert.doesNotMatch(
    stylesheet,
    /\.learner-filter-tabs \.ant-tabs-tab-active\s*\{[^}]*background:\s*var\(--tenant-selected-background\)/s,
  )
})
