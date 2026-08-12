import type { TenantDashboard } from '../api/dashboard'

export interface MetricCard {
  title: string
  value: number
  detail?: string
  comparison?: { direction: 'up' | 'down' | 'same'; value: number }
}

export function stationDashboardCards(data: TenantDashboard): MetricCard[] {
  const delta = data.today_learning_user_delta
  return [
    {
      title: '今日学习学员',
      value: data.today_learning_user_count,
      comparison: {
        direction: delta > 0 ? 'up' : delta < 0 ? 'down' : 'same',
        value: Math.abs(delta),
      },
    },
    { title: '学员总数', value: data.learner_count, detail: `今日新增 ${data.today_new_learner_count}` },
    { title: '已发布课程', value: data.published_course_count, detail: `共 ${data.course_count} 门` },
    { title: '资源分类', value: data.resource_category_count },
    { title: '资源总数', value: data.resource_count },
    { title: '管理成员', value: data.manager_count },
  ]
}

const RESOURCE_META = {
  video: { name: '视频', color: 'var(--admin-warning)' },
  image: { name: '图片', color: 'var(--admin-info)' },
  document: { name: '文档', color: 'var(--admin-success)' },
  attachment: { name: '课程附件', color: 'var(--admin-accent)' },
} as const

export function resourceSeries(data: TenantDashboard) {
  return (Object.keys(RESOURCE_META) as Array<keyof typeof RESOURCE_META>).map((key) => ({
    key,
    ...RESOURCE_META[key],
    value: data.resource_type_counts[key] || 0,
  }))
}

export function planQuotaView(usedBytes: number, quotaBytes: number) {
  return {
    percent: quotaBytes > 0 ? Math.min(100, Math.round(usedBytes / quotaBytes * 100)) : 0,
    unlimited: quotaBytes <= 0,
  }
}

export function rankingPosition(index: number) {
  const rank = Math.max(1, Math.floor(index) + 1)
  return { rank, label: `第 ${rank} 名`, medal: rank <= 3 }
}

export function formatStudyDuration(seconds: number): string {
  const total = Number.isFinite(seconds) ? Math.max(0, Math.floor(seconds)) : 0
  if (total < 60) return `${total} 秒`
  const minutes = Math.floor(total / 60)
  const remainder = total % 60
  return remainder ? `${minutes} 分 ${remainder} 秒` : `${minutes} 分钟`
}
