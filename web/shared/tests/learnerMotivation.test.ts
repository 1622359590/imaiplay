import assert from 'node:assert/strict'
import test from 'node:test'
import {
  acknowledgeAndContinue,
  formatLearningDuration,
  motivationTargetPath,
  normalizeLearnerMotivation,
} from '../src/learning/learnerMotivation.ts'

const course = {
  id: 'course-1',
  title: '新员工入职培训',
  assignment_type: 'required',
  lesson_count: 4,
  progress_percent: 25,
  lesson_id: 'lesson-2',
  lesson_title: '安全规范',
  last_position_seconds: 90,
}

test('normalizes all learner motivation prompt kinds', () => {
  assert.deepEqual(normalizeLearnerMotivation({ kind: 'none', prompt_key: 'ignored' }), { kind: 'none' })

  const welcome = normalizeLearnerMotivation({
    kind: 'welcome',
    prompt_key: 'welcome-key',
    title: '欢迎开启你的学习旅程',
    message: '从第一门课程开始。',
    course,
  })
  assert.equal(welcome.kind, 'welcome')
  assert.ok(welcome.course)
  assert.equal(welcome.course.progressPercent, 25)
  assert.equal(welcome.course.lastPositionSeconds, 90)

  const welcomeWithoutTask = normalizeLearnerMotivation({
    kind: 'welcome',
    prompt_key: 'welcome-empty-key',
    title: '欢迎开启你的学习旅程',
    message: '当前暂无学习任务。',
  })
  assert.equal(welcomeWithoutTask.kind, 'welcome')
  assert.equal(welcomeWithoutTask.course, undefined)
  assert.equal(motivationTargetPath(welcomeWithoutTask), undefined)

  const summary = normalizeLearnerMotivation({
    kind: 'daily_summary',
    prompt_key: 'summary-key',
    study_date: '2026-08-19',
    title: '昨天的学习有了新成果',
    message: '继续保持。',
    metrics: {
      yesterday_seconds: 3900,
      lesson_count: 3,
      completed_lesson_count: 2,
      completed_course_count: 1,
      required_completed: 2,
      required_total: 4,
    },
    comparison: {
      duration_change_seconds: 600,
      exceeded_percent: 72,
      active_learner_count: 18,
    },
    course,
  })
  assert.equal(summary.kind, 'daily_summary')
  assert.equal(summary.metrics.yesterdaySeconds, 3900)
  assert.equal(summary.comparison?.exceededPercent, 72)

  const reengagement = normalizeLearnerMotivation({
    kind: 'reengagement',
    prompt_key: 'return-key',
    title: '欢迎回来',
    message: '今天也向前一点。',
    course,
  })
  assert.equal(reengagement.kind, 'reengagement')
})

test('rejects malformed or unsafe visible prompts', () => {
  const valid = {
    kind: 'welcome',
    prompt_key: 'key',
    title: '标题',
    message: '正文',
    course,
  }
  for (const payload of [
    { ...valid, kind: 'unknown' },
    { ...valid, prompt_key: '   ' },
    { ...valid, course: { ...course, progress_percent: -1 } },
    { ...valid, course: { ...course, progress_percent: 101 } },
    { ...valid, course: { ...course, lesson_id: '' } },
    {
      ...valid,
      kind: 'daily_summary',
      metrics: { yesterday_seconds: -1 },
    },
    {
      ...valid,
      kind: 'daily_summary',
      metrics: {
        yesterday_seconds: 1,
        lesson_count: 0,
        completed_lesson_count: 0,
        completed_course_count: 0,
        required_completed: 0,
        required_total: 0,
      },
      comparison: { exceeded_percent: 100 },
    },
  ]) {
    assert.throws(() => normalizeLearnerMotivation(payload))
  }
})

test('formats persisted learning duration without inventing precision', () => {
  assert.equal(formatLearningDuration(0), '不足 1 分钟')
  assert.equal(formatLearningDuration(59), '不足 1 分钟')
  assert.equal(formatLearningDuration(60), '1 分钟')
  assert.equal(formatLearningDuration(3600), '1 小时')
  assert.equal(formatLearningDuration(3900), '1 小时 5 分钟')
})

test('derives the lesson target only when the action is complete', () => {
  const prompt = normalizeLearnerMotivation({
    kind: 'welcome', prompt_key: 'key', title: '标题', message: '正文', course,
  })
  assert.equal(motivationTargetPath(prompt), '/courses/course-1/lessons/lesson-2')
  assert.equal(motivationTargetPath({ kind: 'none' }), undefined)
})

test('always continues after acknowledgement settles', async () => {
  let continued = 0
  await acknowledgeAndContinue(async () => undefined, () => { continued += 1 })
  assert.equal(continued, 1)

  await assert.rejects(() => acknowledgeAndContinue(
    async () => { throw new Error('offline') },
    () => { continued += 1 },
  ), /offline/)
  assert.equal(continued, 2)
})
