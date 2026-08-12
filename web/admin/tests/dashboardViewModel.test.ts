import assert from 'node:assert/strict'
import test from 'node:test'
import type {
  DashboardResponse,
  InstructorDashboard,
  PlatformDashboard,
  TenantDashboard,
} from '../src/api/dashboard.ts'
import {
  formatStudyDuration,
  planQuotaView,
  rankingPosition,
  resourceSeries,
  stationDashboardCards,
} from '../src/utils/dashboardViewModel.ts'

function tenant(delta: number, resourceCount = 0): TenantDashboard {
  return {
    scope: 'tenant',
    today_learning_user_count: 4,
    yesterday_learning_user_count: 6,
    today_learning_user_delta: delta,
    learner_count: 12,
    today_new_learner_count: 1,
    published_course_count: 3,
    course_count: 4,
    resource_category_count: 2,
    resource_count: resourceCount,
    manager_count: 2,
    has_demo_data: false,
    resource_type_counts: { video: resourceCount, image: 0, document: 0, attachment: 0 },
    today_learning_ranking: [],
  }
}

test('station cards distinguish down, unchanged, and up comparisons', () => {
  assert.deepEqual(stationDashboardCards(tenant(-2))[0].comparison, {
    direction: 'down',
    value: 2,
  })
  assert.deepEqual(stationDashboardCards(tenant(0))[0].comparison, {
    direction: 'same',
    value: 0,
  })
  assert.deepEqual(stationDashboardCards(tenant(3))[0].comparison, {
    direction: 'up',
    value: 3,
  })
})

test('resource series preserves all four types and reconciles with total', () => {
  const data = tenant(0, 5)
  const series = resourceSeries(data)
  assert.deepEqual(series.map((item) => item.key), ['video', 'image', 'document', 'attachment'])
  assert.equal(series.reduce((sum, item) => sum + item.value, 0), data.resource_count)
  assert.deepEqual(series.map((item) => item.color), [
    'var(--admin-warning)',
    'var(--admin-info)',
    'var(--admin-success)',
    'var(--admin-accent)',
  ])
})

test('empty ranking and unlimited quota have explicit presentation values', () => {
  assert.deepEqual(tenant(0).today_learning_ranking, [])
  assert.deepEqual(planQuotaView(1024, 0), {
    percent: 0,
    unlimited: true,
  })
})

test('dashboard response narrows cleanly for every scope', () => {
  const responses: DashboardResponse[] = [
    tenant(0),
    {
      scope: 'platform',
      tenant_count: 2,
      active_tenant_count: 1,
      learner_count: 8,
      course_count: 3,
      recent_tenants: [],
    } satisfies PlatformDashboard,
    {
      scope: 'instructor',
      course_count: 2,
      published_course_count: 1,
      today_learning_user_count: 4,
      recent_courses: [],
    } satisfies InstructorDashboard,
  ]
  assert.deepEqual(responses.map((response) => response.scope), ['tenant', 'platform', 'instructor'])
})

test('ranking positions have visible labels and short sessions stay honest', () => {
  assert.deepEqual(rankingPosition(0), { rank: 1, label: '第 1 名', medal: true })
  assert.deepEqual(rankingPosition(3), { rank: 4, label: '第 4 名', medal: false })
  assert.equal(formatStudyDuration(0), '0 秒')
  assert.equal(formatStudyDuration(42), '42 秒')
  assert.equal(formatStudyDuration(125), '2 分 5 秒')
})
