import { Empty } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { TenantDashboard } from '../api/dashboard'
import { resourceSeries } from '../utils/dashboardViewModel'
import { loadResourceChart } from '../utils/resourceChart'

type EChartsModule = typeof import('echarts')

interface ResourceDonutProps {
  data: TenantDashboard
}

export default function ResourceDonut({ data }: ResourceDonutProps) {
  const container = useRef<HTMLDivElement>(null)
  const [chartLibrary, setChartLibrary] = useState<EChartsModule | null>()
  const [chartFailed, setChartFailed] = useState(false)
  const series = resourceSeries(data)
  const total = series.reduce((sum, item) => sum + item.value, 0)

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
      const adminStyles = getComputedStyle(document.documentElement)
      const adminColor = (property: string) => adminStyles.getPropertyValue(property).trim()
      const chartColor = (value: string) => {
        const property = value.match(/^var\((--[^)]+)\)$/)?.[1]
        return property ? adminColor(property) : value
      }
      chart = chartLibrary.init(container.current)
      chart.setOption({
        animationDuration: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 500,
        tooltip: { trigger: 'item', formatter: '{b}：{c}（{d}%）' },
        title: {
          text: String(total),
          subtext: '总资源数',
          left: 'center',
          top: '39%',
          textStyle: { color: adminColor('--admin-heading'), fontSize: 26, fontWeight: 600 },
          subtextStyle: { color: adminColor('--admin-muted'), fontSize: 12 },
        },
        series: [{
          name: '资源类型',
          type: 'pie',
          radius: ['50%', '72%'],
          center: ['50%', '48%'],
          avoidLabelOverlap: true,
          itemStyle: { borderColor: adminColor('--admin-card'), borderWidth: 2 },
          label: { formatter: '{b}\n{c}', color: adminColor('--admin-text') },
          labelLine: { length: 10, length2: 8 },
          data: series.filter((item) => item.value > 0).map((item) => ({
            name: item.name,
            value: item.value,
            itemStyle: { color: chartColor(item.color) },
          })),
        }],
      })
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
  }, [chartLibrary, data, total])

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
