import { Empty } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { TenantDashboard } from '../api/dashboard'
import { useAdminTheme } from '../context/AdminThemeContext'
import { createAdminPalette } from '../theme/adminPalette'
import { resourceSeries } from '../utils/dashboardViewModel'
import {
  createResourceChartOption,
  loadResourceChart,
  resourceChartThemeKey,
} from '../utils/resourceChart'

type EChartsModule = typeof import('echarts')

interface ResourceDonutProps {
  data: TenantDashboard
}

export default function ResourceDonut({ data }: ResourceDonutProps) {
  const adminTheme = useAdminTheme()
  const container = useRef<HTMLDivElement>(null)
  const [chartLibrary, setChartLibrary] = useState<EChartsModule | null>()
  const [chartFailed, setChartFailed] = useState(false)
  const series = resourceSeries(data)
  const total = series.reduce((sum, item) => sum + item.value, 0)
  const chartTheme = useMemo(() => {
    const palette = createAdminPalette(adminTheme.primaryColor)
    return {
      primaryColor: adminTheme.primaryColor,
      heading: palette.heading,
      muted: palette.muted,
      text: palette.text,
      card: palette.card,
      warning: palette.warning,
      info: palette.info,
      success: palette.success,
      accent: palette.accent,
    }
  }, [adminTheme.primaryColor])
  const chartThemeVersion = resourceChartThemeKey(chartTheme)

  useEffect(() => {
    let active = true
    void loadResourceChart(() => import('echarts'))
      .then((library) => { if (active) setChartLibrary(library) })
    return () => { active = false }
  }, [])

  useEffect(() => {
    if (!container.current || total <= 0 || !chartLibrary) return
    let chart: import('echarts').ECharts | undefined
    let observer: ResizeObserver | undefined
    try {
      chart = chartLibrary.init(container.current)
      chart.setOption(createResourceChartOption(
        series,
        chartTheme,
        window.matchMedia('(prefers-reduced-motion: reduce)').matches,
      ))
      observer = new ResizeObserver(() => chart?.resize())
      observer.observe(container.current)
      setChartFailed(false)
    } catch {
      setChartFailed(true)
    }
    return () => {
      observer?.disconnect()
      chart?.dispose()
    }
  }, [chartLibrary, chartThemeVersion, data, total])

  if (total <= 0) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无资源" />

  return (
    <div className="resource-donut-wrap">
      {chartLibrary !== null && !chartFailed && <div ref={container} className="resource-donut-chart" role="img" aria-label={`资源统计，总计 ${total} 个`} />}
      {(chartLibrary === null || chartFailed) && <div className="chart-fallback" role="status">图表暂时无法显示，以下为资源明细。</div>}
      <ul className="resource-donut-legend" aria-label="资源类型明细">
        {series.map((item) => (
          <li key={item.key}>
            <span className="resource-legend-swatch" style={{ backgroundColor: item.color }} aria-hidden="true" />
            <span>{item.name}</span>
            <strong>{item.value}</strong>
          </li>
        ))}
      </ul>
    </div>
  )
}
