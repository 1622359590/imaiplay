import { Empty } from 'antd'
import * as echarts from 'echarts'
import { useEffect, useRef, useState } from 'react'
import type { TenantDashboard } from '../api/dashboard'
import { resourceSeries } from '../utils/dashboardViewModel'

interface ResourceDonutProps {
  data: TenantDashboard
}

export default function ResourceDonut({ data }: ResourceDonutProps) {
  const container = useRef<HTMLDivElement>(null)
  const [chartFailed, setChartFailed] = useState(false)
  const series = resourceSeries(data)
  const total = series.reduce((sum, item) => sum + item.value, 0)

  useEffect(() => {
    if (!container.current || total <= 0) return
    let chart: echarts.ECharts | undefined
    let observer: ResizeObserver | undefined
    try {
      chart = echarts.init(container.current)
      chart.setOption({
        animationDuration: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 500,
        tooltip: { trigger: 'item', formatter: '{b}：{c}（{d}%）' },
        title: {
          text: String(total),
          subtext: '总资源数',
          left: 'center',
          top: '39%',
          textStyle: { color: '#262626', fontSize: 26, fontWeight: 600 },
          subtextStyle: { color: '#737373', fontSize: 12 },
        },
        series: [{
          name: '资源类型',
          type: 'pie',
          radius: ['50%', '72%'],
          center: ['50%', '48%'],
          avoidLabelOverlap: true,
          itemStyle: { borderColor: '#fff', borderWidth: 2 },
          label: { formatter: '{b}\n{c}', color: '#595959' },
          labelLine: { length: 10, length2: 8 },
          data: series.filter((item) => item.value > 0).map((item) => ({
            name: item.name,
            value: item.value,
            itemStyle: { color: item.color },
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
  }, [data, total])

  if (total <= 0) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无资源" />

  return (
    <div className="resource-donut-wrap">
      {!chartFailed && <div ref={container} className="resource-donut-chart" role="img" aria-label={`资源统计，总计 ${total} 个`} />}
      {chartFailed && <div className="chart-fallback" role="status">图表暂时无法显示，以下为资源明细。</div>}
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
