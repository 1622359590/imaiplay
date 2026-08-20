import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const player = readFileSync(new URL('../src/pages/LessonPlayerPage.tsx', import.meta.url), 'utf8')
const home = readFileSync(new URL('../src/pages/HomePage.tsx', import.meta.url), 'utf8')
const detail = readFileSync(new URL('../src/pages/CourseDetailPage.tsx', import.meta.url), 'utf8')

test('resource-backed lessons bind lifecycle after the real video node mounts', () => {
  assert.match(player, /getResourceFile\(lesson\.resourceId\)[\s\S]*setResourceURL\(url\)/)
  assert.match(player, /const bindVideoRef = useCallback\(\(node:[^)]+\) => \{\s*setMediaElement\(node\)/)
  assert.match(player, /<video[\s\S]*?ref=\{bindVideoRef\}/)
  assert.match(player, /useLayoutEffect\(\(\) => \{[\s\S]*?!mediaElement\) return[\s\S]*?lifecycleGate\.bind\(activeLessonId, mediaElement, controller\)/)
  assert.match(player, /\}, \[lesson\?\.duration, lessonId, loadedLessonId, mediaElement\]\)/)
  assert.match(player, /document\.addEventListener\('visibilitychange'/)
  assert.match(player, /window\.addEventListener\('pagehide'/)
  assert.match(player, /window\.setInterval\([\s\S]*?controller\.periodicFlush\(\)/)
})

test('learner pages keep required data errors separate from optional overview fallback', () => {
  assert.match(home, /loadCoursesWithOptionalOverview/)
  assert.match(home, /role="alert"/)
  assert.match(detail, /loadCourseWithOptionalOverview/)
  assert.match(detail, /课程不可访问/)
})

test('H5 motivation popup mounts only after required courses succeed', () => {
  const prompt = readFileSync(new URL('../src/components/LearnerMotivationPrompt.tsx', import.meta.url), 'utf8')
  assert.match(home, /<LearnerMotivationPrompt\s+enabled=\{!loading && !loadError\}/)
  assert.match(prompt, /position="bottom"/)
  assert.match(prompt, /afterShow=/)
  assert.match(prompt, /acknowledgeAndContinue/)
  assert.match(prompt, /theme\.routePath/)
  assert.match(prompt, /role="dialog"/)
  assert.match(prompt, /aria-labelledby="h5-motivation-title"/)
  assert.match(prompt, /\.catch\(\(\) => undefined\)/)
})
