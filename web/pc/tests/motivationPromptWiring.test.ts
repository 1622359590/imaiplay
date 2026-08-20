import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const component = readFileSync(new URL('../src/components/LearnerMotivationPrompt.tsx', import.meta.url), 'utf8')
const home = readFileSync(new URL('../src/pages/HomePage.tsx', import.meta.url), 'utf8')

test('PC prompt is non-blocking, accessible, and acknowledgement-safe', () => {
  assert.match(component, /if \(!enabled\)/)
  assert.match(component, /prompt\.kind === 'none'/)
  assert.match(component, /\.catch\(\(\) => undefined\)/)
  assert.match(component, /afterOpenChange=/)
  assert.match(component, /acknowledgeAndContinue/)
  assert.match(component, /portalRoutePath/)
  assert.match(component, /aria-labelledby="learner-motivation-title"/)
  assert.match(component, /closable/)
})

test('PC home mounts motivation only alongside a successful overview', () => {
  assert.match(home, /<LearnerMotivationPrompt\s+enabled=\{!loading && !error && Boolean\(overview\)\}/)
})
