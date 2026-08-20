export type LearnerMotivationKind = 'none' | 'welcome' | 'daily_summary' | 'reengagement'

export interface LearnerMotivationCourse {
  id: string
  title: string
  assignmentType: string
  lessonCount: number
  progressPercent: number
  lessonId: string
  lessonTitle: string
  lastPositionSeconds: number
}

export interface LearnerMotivationMetrics {
  yesterdaySeconds: number
  lessonCount: number
  completedLessonCount: number
  completedCourseCount: number
  requiredCompleted: number
  requiredTotal: number
}

export interface LearnerMotivationComparison {
  durationChangeSeconds?: number
  exceededPercent?: number
  activeLearnerCount?: number
}

export type LearnerMotivation =
  | { kind: 'none' }
  | {
      kind: 'welcome' | 'reengagement'
      promptKey: string
      title: string
      message: string
      course: LearnerMotivationCourse
    }
  | {
      kind: 'daily_summary'
      promptKey: string
      studyDate: string
      title: string
      message: string
      metrics: LearnerMotivationMetrics
      comparison?: LearnerMotivationComparison
      course: LearnerMotivationCourse
    }

type RawRecord = Record<string, unknown>

function record(value: unknown, field: string): RawRecord {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new TypeError(`${field} must be an object`)
  }
  return value as RawRecord
}

function text(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.trim() === '') {
    throw new TypeError(`${field} must be a non-empty string`)
  }
  return value.trim()
}

function integer(value: unknown, field: string, minimum = 0, maximum?: number): number {
  if (!Number.isSafeInteger(value) || (value as number) < minimum || (maximum !== undefined && (value as number) > maximum)) {
    throw new TypeError(`${field} is outside its valid range`)
  }
  return value as number
}

function optionalInteger(
  source: RawRecord,
  key: string,
  minimum: number,
  maximum?: number,
): number | undefined {
  const value = source[key]
  if (value === undefined || value === null) return undefined
  return integer(value, key, minimum, maximum)
}

function normalizeCourse(value: unknown): LearnerMotivationCourse {
  const source = record(value, 'course')
  return {
    id: text(source.id, 'course.id'),
    title: text(source.title, 'course.title'),
    assignmentType: text(source.assignment_type, 'course.assignment_type'),
    lessonCount: integer(source.lesson_count, 'course.lesson_count'),
    progressPercent: integer(source.progress_percent, 'course.progress_percent', 0, 100),
    lessonId: text(source.lesson_id, 'course.lesson_id'),
    lessonTitle: text(source.lesson_title, 'course.lesson_title'),
    lastPositionSeconds: integer(source.last_position_seconds, 'course.last_position_seconds'),
  }
}

function normalizeMetrics(value: unknown): LearnerMotivationMetrics {
  const source = record(value, 'metrics')
  return {
    yesterdaySeconds: integer(source.yesterday_seconds, 'metrics.yesterday_seconds'),
    lessonCount: integer(source.lesson_count, 'metrics.lesson_count'),
    completedLessonCount: integer(source.completed_lesson_count, 'metrics.completed_lesson_count'),
    completedCourseCount: integer(source.completed_course_count, 'metrics.completed_course_count'),
    requiredCompleted: integer(source.required_completed, 'metrics.required_completed'),
    requiredTotal: integer(source.required_total, 'metrics.required_total'),
  }
}

function normalizeComparison(value: unknown): LearnerMotivationComparison | undefined {
  if (value === undefined || value === null) return undefined
  const source = record(value, 'comparison')
  const durationChangeSeconds = optionalInteger(source, 'duration_change_seconds', Number.MIN_SAFE_INTEGER)
  const exceededPercent = optionalInteger(source, 'exceeded_percent', 0, 99)
  const activeLearnerCount = optionalInteger(source, 'active_learner_count', 0)
  if (durationChangeSeconds === undefined && exceededPercent === undefined && activeLearnerCount === undefined) {
    return undefined
  }
  return { durationChangeSeconds, exceededPercent, activeLearnerCount }
}

export function normalizeLearnerMotivation(raw: unknown): LearnerMotivation {
  const source = record(raw, 'motivation')
  if (source.kind === 'none') return { kind: 'none' }
  if (source.kind !== 'welcome' && source.kind !== 'daily_summary' && source.kind !== 'reengagement') {
    throw new TypeError('motivation.kind is not supported')
  }

  const common = {
    promptKey: text(source.prompt_key, 'motivation.prompt_key'),
    title: text(source.title, 'motivation.title'),
    message: text(source.message, 'motivation.message'),
    course: normalizeCourse(source.course),
  }
  if (source.kind === 'daily_summary') {
    return {
      kind: source.kind,
      ...common,
      studyDate: text(source.study_date, 'motivation.study_date'),
      metrics: normalizeMetrics(source.metrics),
      comparison: normalizeComparison(source.comparison),
    }
  }
  return { kind: source.kind, ...common }
}

export function formatLearningDuration(seconds: number): string {
  const normalized = Number.isFinite(seconds) ? Math.max(0, Math.floor(seconds)) : 0
  if (normalized < 60) return '不足 1 分钟'
  const minutes = Math.floor(normalized / 60)
  if (minutes < 60) return `${minutes} 分钟`
  const hours = Math.floor(minutes / 60)
  const remainder = minutes % 60
  return remainder === 0 ? `${hours} 小时` : `${hours} 小时 ${remainder} 分钟`
}

export function motivationTargetPath(prompt: LearnerMotivation): string | undefined {
  if (prompt.kind === 'none') return undefined
  return `/courses/${encodeURIComponent(prompt.course.id)}/lessons/${encodeURIComponent(prompt.course.lessonId)}`
}

export async function acknowledgeAndContinue(
  acknowledge: () => Promise<unknown>,
  continuation: () => void,
): Promise<void> {
  try {
    await acknowledge()
  } finally {
    continuation()
  }
}
