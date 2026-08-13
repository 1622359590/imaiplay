import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { readdirSync } from 'node:fs'
import test from 'node:test'
import { readStyleBundle } from './styleSource.ts'

const entry = new URL('../src/styles.css', import.meta.url)
const stylesheet = readStyleBundle(entry)

test('style entrypoint contains only ordered imports', () => {
  assert.deepEqual(readFileSync(entry, 'utf8').trim().split('\n'), [
    "@import './styles/base.css';",
    "@import './styles/layout.css';",
    "@import './styles/login.css';",
    "@import './styles/dashboard.css';",
    "@import './styles/course.css';",
    "@import './styles/player.css';",
    "@import './styles/responsive.css';",
  ])
})

function closingBrace(source: string, openingBrace: number): number {
  let depth = 0
  for (let index = openingBrace; index < source.length; index += 1) {
    if (source[index] === '{') depth += 1
    if (source[index] === '}') depth -= 1
    if (depth === 0) return index
  }
  throw new Error('unbalanced stylesheet')
}

function mediaMatches(query: string, width: number): boolean {
  const minimum = query.match(/min-width:\s*(\d+)px/i)
  const maximum = query.match(/max-width:\s*(\d+)px/i)
  return (!minimum || width >= Number(minimum[1]))
    && (!maximum || width <= Number(maximum[1]))
}

function courseColumnsAt(width: number): string | undefined {
  let columns: string | undefined

  function visit(source: string) {
    let cursor = 0
    while (cursor < source.length) {
      const openingBrace = source.indexOf('{', cursor)
      if (openingBrace === -1) return
      const header = source.slice(cursor, openingBrace).replace(/\/\*[\s\S]*?\*\//g, '').trim()
      const end = closingBrace(source, openingBrace)
      const body = source.slice(openingBrace + 1, end)

      if (header.startsWith('@media') && mediaMatches(header, width)) {
        visit(body)
      } else if (header.split(',').some((selector) => selector.trim() === '.course-grid')) {
        const declaration = body.match(/grid-template-columns:\s*([^;]+);/i)
        if (declaration) columns = declaration[1].trim()
      }
      cursor = end + 1
    }
  }

  visit(stylesheet)
  return columns
}

function declarationAt(selector: string, property: string, width: number): string | undefined {
  let value: string | undefined
  const escapedProperty = property.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

  function visit(source: string) {
    let cursor = 0
    while (cursor < source.length) {
      const openingBrace = source.indexOf('{', cursor)
      if (openingBrace === -1) return
      const header = source.slice(cursor, openingBrace).replace(/\/\*[\s\S]*?\*\//g, '').trim()
      const end = closingBrace(source, openingBrace)
      const body = source.slice(openingBrace + 1, end)

      if (header.startsWith('@media') && mediaMatches(header, width)) {
        visit(body)
      } else if (header.split(',').some((candidate) => candidate.trim() === selector)) {
        const declaration = body.match(new RegExp(`${escapedProperty}:\\s*([^;]+);`, 'i'))
        if (declaration) value = declaration[1].trim()
      }
      cursor = end + 1
    }
  }

  visit(stylesheet)
  return value
}

test('course grid keeps one, two, and three-column breakpoint contract', () => {
  const expected = new Map<number, string>([
    [759, '1fr'],
    [760, 'repeat(2, minmax(0, 1fr))'],
    [900, 'repeat(2, minmax(0, 1fr))'],
    [1199, 'repeat(2, minmax(0, 1fr))'],
    [1200, 'repeat(3, minmax(0, 1fr))'],
  ])

  for (const [width, columns] of expected) {
    assert.equal(courseColumnsAt(width), columns, `${width}px course grid`)
  }
})

test('category Select exposes focus on the actual Ant Design focused root', () => {
  assert.match(
    stylesheet,
    /\.learner-category-select\.ant-select-focused\s+\.ant-select-selector\s*\{[^}]*outline:\s*3px\s+solid\s+var\(--learner-accent-foreground\)/s,
  )
})

test('notification control owns one square positioning context for its icon and dot', () => {
  const layoutSource = readFileSync(new URL('../src/components/AppLayout.tsx', import.meta.url), 'utf8')
  assert.match(
    layoutSource,
    /<button className="learner-notification-button" type="button" aria-label="通知">\s*<BellOutlined \/>\s*<span className="learner-notification-dot"/s,
  )
  assert.doesNotMatch(layoutSource, /<Button className="learner-notification-button"/)
  assert.match(
    stylesheet,
    /\.learner-notification-button\s*\{[^}]*display:\s*inline-grid[^}]*width:\s*36px[^}]*height:\s*36px[^}]*padding:\s*0[^}]*border:\s*0[^}]*place-items:\s*center[^}]*line-height:\s*1/s,
  )
  assert.match(
    stylesheet,
    /\.learner-notification-button\s+\.anticon\s*\{[^}]*display:\s*grid[^}]*place-items:\s*center[^}]*line-height:\s*1/s,
  )
})

test('primary buttons use contrast text derived for the tenant accent surface', () => {
  assert.match(
    stylesheet,
    /\.ant-btn-primary[^}]*\{[^}]*color:\s*var\(--learner-accent-contrast-text\)/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-primary:hover\s*\{[^}]*color:\s*var\(--learner-accent-hover-contrast-text\)/s,
  )
})

test('primary buttons expose the Clay depth and press interaction contract', () => {
  assert.match(
    stylesheet,
    /\.ant-btn-primary\s*\{[^}]*box-shadow:\s*0\s+6px\s+0\s+var\(--learner-clay-shadow\)[^}]*transition:[^;]*100ms\s+ease-out/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-primary:hover\s*\{[^}]*transform:\s*translateY\(2px\)[^}]*box-shadow:\s*0\s+4px\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-primary:active\s*\{[^}]*transform:\s*translateY\(6px\)[^}]*box-shadow:\s*0\s+0\s+0\s+var\(--learner-clay-shadow\)/s,
  )
})

test('colored Clay cards use the derived tenant contact shadow', () => {
  assert.match(
    stylesheet,
    /\.continue-learning-banner\s*\{[^}]*box-shadow:[^;]*var\(--learner-clay-shadow\)/s,
  )
})

test('continue-learning CTA keeps the same 6-4-0 interaction depth at every viewport', () => {
  assert.match(
    stylesheet,
    /\.continue-learning-action\s*\{[^}]*box-shadow:\s*0\s+6px\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.continue-learning-action:hover\s*\{[^}]*transform:\s*translateY\(2px\)[^}]*box-shadow:\s*0\s+4px\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.continue-learning-action:active\s*\{[^}]*transform:\s*translateY\(6px\)[^}]*box-shadow:\s*0\s+0\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.doesNotMatch(
    stylesheet,
    /@media\s*\(max-width:\s*759px\)[\s\S]*?\.continue-learning-action(?::hover|:active)?\s*\{[^}]*box-shadow:/s,
  )
})

test('secondary controls keep white Clay depth while text actions stay flat', () => {
  assert.match(
    stylesheet,
    /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\)\s*\{[^}]*box-shadow:\s*0\s+6px\s+0\s+var\(--learner-clay-white-shadow\)[^}]*transition:[^;]*100ms\s+ease-out/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-text,\s*\.ant-btn-link\s*\{[^}]*box-shadow:\s*none/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\):not\(:disabled\):hover\s*\{/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\):not\(:disabled\):active\s*\{/s,
  )
  assert.match(
    stylesheet,
    /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\):disabled\s*\{[^}]*background:\s*var\(--learner-card\)[^}]*box-shadow:\s*0\s+2px\s+0\s+var\(--learner-clay-white-shadow\)[^}]*transform:\s*none\s*!important/s,
  )
})

test('white learner surfaces use neutral Clay contact depth', () => {
  for (const selector of [
    '.learning-summary-card',
    '.learner-course-card',
    '.course-chapter',
    '.login-card',
  ]) {
    assert.match(
      stylesheet,
      new RegExp(`\\${selector}\\s*\\{[^}]*box-shadow:\\s*0\\s+(?:6|8)px\\s+0\\s+var\\(--learner-clay-white-shadow\\)`, 's'),
      selector,
    )
  }
})

test('course progress uses molded pill tracks', () => {
  assert.match(
    stylesheet,
    /\.course-progress-bar \.ant-progress-inner\s*\{[^}]*border-radius:\s*999px[^}]*box-shadow:\s*inset/s,
  )
  assert.match(
    stylesheet,
    /\.course-progress-bar \.ant-progress-bg\s*\{[^}]*border-radius:\s*999px[^}]*background:\s*var\(--learner-clay-surface\)[^}]*var\(--learner-clay-shadow\)/s,
  )
})

test('player Clay stays on controls and sidebar navigation, not the video', () => {
  assert.match(
    stylesheet,
    /\.player-controls \.player-play-button\s*\{[^}]*box-shadow:[^;]*var\(--learner-clay-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.player-lesson-item\.active\s*\{[^}]*box-shadow:[^;]*var\(--learner-clay-white-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.player-controls \.ant-btn-default\.player-play-button\s*\{[^}]*box-shadow:\s*0\s+6px\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.player-controls \.ant-btn-default\.player-play-button:hover\s*\{[^}]*transform:\s*translateY\(2px\)[^}]*box-shadow:\s*0\s+4px\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.match(
    stylesheet,
    /\.player-controls \.ant-btn-default\.player-play-button:active\s*\{[^}]*transform:\s*translateY\(6px\)[^}]*box-shadow:\s*0\s+0\s+0\s+var\(--learner-clay-shadow\)/s,
  )
  assert.doesNotMatch(stylesheet, /\.lesson-video\s*\{[^}]*box-shadow:/s)
})

test('neutral contact shadows stay on white surfaces and colored surfaces use tenant depth', () => {
  for (const [selector, depth] of [
    ['.learning-summary-icon', 4],
    ['.course-detail-page .lesson-name > .anticon', 3],
    ['.course-material-icon', 3],
    ['.organization-logo', 3],
  ] as const) {
    const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    assert.match(
      stylesheet,
      new RegExp(`${escaped}\\s*\\{[^}]*background:\\s*var\\(--learner-card\\)[^}]*box-shadow:\\s*0\\s+${depth}px\\s+0\\s+var\\(--learner-clay-white-shadow\\)`, 's'),
      selector,
    )
  }

  for (const selector of [
    '.learning-summary-progress .learning-summary-icon',
    '.learning-summary-complete .learning-summary-icon',
    '.learning-summary-today .learning-summary-icon',
    '.learning-summary-total .learning-summary-icon',
    '.course-detail-page .lesson-type-document > .anticon',
    '.course-detail-page .lesson-type-text > .anticon',
  ]) {
    assert.equal(declarationAt(selector, 'background', 1200), undefined, `${selector} background override`)
  }

  for (const selector of [
    '.learner-avatar',
    '.continue-learning-banner',
    '.player-controls .ant-btn-default.player-play-button',
  ]) {
    assert.match(declarationAt(selector, 'background', 1200) ?? '', /var\(--learner-(?:clay-surface|accent)/, selector)
    assert.match(declarationAt(selector, 'box-shadow', 1200) ?? '', /var\(--learner-clay-shadow\)/, selector)
  }
})

test('narrow and reduced-motion styles reduce Clay depth without moving controls', () => {
  for (const selector of [
    '.detail-loading',
    '.learner-request-state',
    '.page-empty',
    '.learning-summary-card',
    '.continue-learning-banner',
    '.learner-course-card',
    '.learner-course-skeleton',
    '.learner-filter-skeleton',
    '.detail-hero',
    '.course-chapter',
    '.course-materials',
    '.recent-course-card',
    '.recent-course-skeleton',
    '.lesson-player-layout',
    '.login-card',
    '.organization-select-card',
  ]) {
    const shadow = declarationAt(selector, 'box-shadow', 759) ?? ''
    assert.match(shadow, /^0\s+[2-5]px\s+0\s+var\(--learner-clay-(?:white-)?shadow\)/, `${selector} contact depth`)
    assert.doesNotMatch(shadow, /\b(?:22|24|26|28|30|34|36|44)px\b/, `${selector} atmosphere blur`)
  }
  assert.match(
    stylesheet,
    /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.ant-btn-primary:hover[^{]*\{[^}]*transform:\s*none\s*!important/s,
  )
})

test('responsive styles never downgrade primary or player play-button interaction depth', () => {
  const narrow = stylesheet.match(/@media\s*\(max-width:\s*759px\)\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
  assert.doesNotMatch(narrow, /\.ant-btn-primary(?::hover|:active)?\s*\{[^}]*box-shadow:/s)
  assert.doesNotMatch(narrow, /\.player-controls \.ant-btn-default\.player-play-button(?::hover|:active)?\s*\{[^}]*box-shadow:/s)
  assert.match(narrow, /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\):not\(\.player-play-button\)\s*\{/)
  assert.match(narrow, /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\):not\(\.player-play-button\):not\(:disabled\):hover\s*\{/)
  assert.match(narrow, /\.ant-btn-default:not\(\.ant-btn-text\):not\(\.ant-btn-link\):not\(\.player-play-button\):not\(:disabled\):active\s*\{/)
})

test('course lesson hover and status text use readable tenant palette tokens', () => {
  assert.match(
    stylesheet,
    /\.course-detail-page\s+\.lesson-row:hover\s*\{[^}]*color:\s*var\(--learner-accent-foreground\)/s,
  )
  assert.match(
    stylesheet,
    /\.course-detail-page\s+\.lesson-state\s*\{[^}]*color:\s*var\(--learner-muted\)/s,
  )
  assert.match(
    stylesheet,
    /\.course-detail-page\s+\.lesson-row:hover\s+\.lesson-state\s*\{[^}]*color:\s*var\(--learner-accent-foreground\)/s,
  )
})

test('completed recent-course status is a readable semantic badge', () => {
  assert.match(
    stylesheet,
    /\.recent-complete-status\s*\{[^}]*display:\s*inline-flex[^}]*background:\s*var\(--learner-success-light\)[^}]*color:\s*var\(--learner-success-strong\)[^}]*border-radius:\s*var\(--learner-radius-badge\)/s,
  )
  assert.match(
    stylesheet,
    /\.recent-complete-status \.anticon\s*\{[^}]*color:\s*var\(--learner-success\)/s,
  )
})

test('learner styles use palette variables instead of color literals', () => {
  const stylesDirectory = new URL('../src/styles/', import.meta.url)
  for (const file of readdirSync(stylesDirectory).filter((name) => name.endsWith('.css'))) {
    const source = readFileSync(new URL(file, stylesDirectory), 'utf8')
    assert.doesNotMatch(source, /#[0-9a-f]{3,8}\b|\brgba?\(|\bhsla?\(/i, file)
  }
})

test('modern SaaS layouts expose the required visual regions', () => {
  for (const selector of [
    '.top-header',
    '.continue-learning-banner',
    '.learner-course-cover',
    '.detail-hero',
    '.lesson-player-layout',
    '.player-sidebar',
    '.login-hero',
    '.login-form-panel',
  ]) {
    assert.match(stylesheet, new RegExp(`\\${selector}\\s*\\{`), selector)
  }
  assert.match(stylesheet, /\.top-header\s*\{[^}]*backdrop-filter:\s*blur\(20px\)/s)
  assert.match(stylesheet, /\.lesson-player-layout\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*7fr\)\s+minmax\(300px,\s*3fr\)/s)
  assert.match(stylesheet, /\.login-page\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*55fr\)\s+minmax\(420px,\s*45fr\)/s)
})
