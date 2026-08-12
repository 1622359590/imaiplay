import type { resourceSeries } from './dashboardViewModel'

type ResourceSeries = ReturnType<typeof resourceSeries>

export interface ResourceChartTheme {
  primaryColor: string
  heading: string
  muted: string
  text: string
  card: string
  warning: string
  info: string
  success: string
  accent: string
}

export function resourceChartThemeKey(theme: ResourceChartTheme): string {
  return [
    theme.primaryColor,
    theme.heading,
    theme.muted,
    theme.text,
    theme.card,
    theme.warning,
    theme.info,
    theme.success,
    theme.accent,
  ].join('|')
}

export function createResourceChartOption(
  series: ResourceSeries,
  theme: ResourceChartTheme,
  reducedMotion: boolean,
) {
  const total = series.reduce((sum, item) => sum + item.value, 0)
  const semanticColors: Record<string, string> = {
    'var(--admin-warning)': theme.warning,
    'var(--admin-info)': theme.info,
    'var(--admin-success)': theme.success,
    'var(--admin-accent)': theme.accent,
  }

  return {
    animationDuration: reducedMotion ? 0 : 500,
    tooltip: { trigger: 'item', formatter: '{b}：{c}（{d}%）' },
    title: {
      text: String(total),
      subtext: '总资源数',
      left: 'center',
      top: '39%',
      textStyle: { color: theme.heading, fontSize: 26, fontWeight: 600 },
      subtextStyle: { color: theme.muted, fontSize: 12 },
    },
    series: [{
      name: '资源类型',
      type: 'pie',
      radius: ['50%', '72%'],
      center: ['50%', '48%'],
      avoidLabelOverlap: true,
      itemStyle: { borderColor: theme.card, borderWidth: 2 },
      label: { formatter: '{b}\n{c}', color: theme.text },
      labelLine: { length: 10, length2: 8 },
      data: series.filter((item) => item.value > 0).map((item) => ({
        name: item.name,
        value: item.value,
        itemStyle: { color: semanticColors[item.color] ?? theme.accent },
      })),
    }],
  }
}

export async function loadResourceChart<T>(
  loader: () => Promise<T>,
): Promise<T | null> {
  try {
    return await loader()
  } catch {
    return null
  }
}
