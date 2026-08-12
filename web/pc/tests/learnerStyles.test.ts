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
