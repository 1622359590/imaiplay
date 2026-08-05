import {
  BookOutlined,
  CloudUploadOutlined,
  PlusOutlined,
  RiseOutlined,
  TeamOutlined,
  TrophyOutlined,
  UserAddOutlined,
} from '@ant-design/icons'
import { Button, Card, Empty, List, message, Modal, Progress, Skeleton, Space, Statistic, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { dashboardApi, type DashboardResponse, type InstructorDashboard, type PlatformDashboard, type TenantDashboard } from '../api/dashboard'
import { domainApi, type DomainBindStatus } from '../api/domain'
import { planApi, type Plan } from '../api/plan'
import { tenantApi } from '../api/tenant'
import ResourceDonut from '../components/ResourceDonut'
import PageHeader from '../components/PageHeader'
import { formatStudyDuration, planQuotaView, rankingPosition, stationDashboardCards } from '../utils/dashboardViewModel'

interface PlanUsage {
  plan: Plan
  used_bytes: number
  quota_bytes: number
}

const formatBytes = (bytes: number) => {
  if (bytes < 1024 ** 2) return `${Math.round(bytes / 1024)} KB`
  if (bytes < 1024 ** 3) return `${(bytes / 1024 ** 2).toFixed(1)} MB`
  return `${(bytes / 1024 ** 3).toFixed(1)} GB`
}

function PlatformWorkbench({ data }: { data: PlatformDashboard }) {
  const cards = [
    ['租户总数', data.tenant_count, <TeamOutlined />],
    ['活跃租户', data.active_tenant_count, <RiseOutlined />],
    ['全平台学员', data.learner_count, <UserAddOutlined />],
    ['全平台课程', data.course_count, <BookOutlined />],
  ] as const
  return (
    <>
      <PageHeader title="平台概览" description="查看全平台租户、学员和课程运营情况。" />
      <div className="platform-metric-grid">
        {cards.map(([title, value, icon]) => <Card className="stat-card" key={title}><div className="stat-icon">{icon}</div><Statistic title={title} value={value} /></Card>)}
      </div>
      <Card title="最近注册租户" className="dashboard-list-card">
        {data.recent_tenants.length ? <List dataSource={data.recent_tenants} renderItem={(tenant) => <List.Item><List.Item.Meta title={tenant.name} description={`${tenant.code} · ${new Date(tenant.created_at).toLocaleString()}`} /><Tag color={tenant.status === 1 ? 'success' : 'default'}>{tenant.status === 1 ? '正常' : '停用'}</Tag></List.Item>} /> : <Empty description="暂无注册租户" />}
      </Card>
    </>
  )
}

function InstructorWorkbench({ data }: { data: InstructorDashboard }) {
  const navigate = useNavigate()
  return (
    <>
      <PageHeader title="教学工作台" description="查看你负责课程的今日教学概况。" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/courses?create=1')}>新建课程</Button>} />
      <div className="instructor-metric-grid">
        <Card><Statistic title="我的课程" value={data.course_count} /></Card>
        <Card><Statistic title="已发布课程" value={data.published_course_count} /></Card>
        <Card><Statistic title="今日学习学员" value={data.today_learning_user_count} /></Card>
      </div>
      <Card title="最近编辑课程" className="dashboard-list-card">
        {data.recent_courses.length ? <List dataSource={data.recent_courses} renderItem={(course) => <List.Item actions={[<Button type="link" key="open" onClick={() => navigate(`/courses/${course.id}`)}>打开</Button>]}><List.Item.Meta title={course.title} description={new Date(course.updated_at).toLocaleString()} /><Tag color={course.status === 1 ? 'success' : 'default'}>{course.status === 1 ? '已发布' : '草稿'}</Tag></List.Item>} /> : <Empty description="暂无课程" />}
      </Card>
    </>
  )
}

function SitePlanCard({ plan, domain, planFailed, domainFailed, retryPlan, retryDomain, data, onClear }: {
  plan?: PlanUsage
  domain?: DomainBindStatus
  planFailed: boolean
  domainFailed: boolean
  retryPlan: () => void
  retryDomain: () => void
  data: TenantDashboard
  onClear: () => void
}) {
  const navigate = useNavigate()
  const quota = plan ? planQuotaView(plan.used_bytes, plan.quota_bytes) : undefined
  return (
    <Card title="站点与套餐" className="station-card station-site-card">
      <div className="site-plan-row">
        <span>当前套餐</span>
        {planFailed ? <Button type="link" onClick={retryPlan}>套餐加载失败，重试</Button> : plan ? <strong>{plan.plan.name}</strong> : <Skeleton.Input active size="small" />}
      </div>
      {plan && <div className="site-plan-progress"><Progress percent={quota?.percent} size="small" /><Typography.Text type="secondary">{formatBytes(plan.used_bytes)} / {quota?.unlimited ? '不限额' : formatBytes(plan.quota_bytes)}</Typography.Text></div>}
      <div className="site-plan-row">
        <span>站点地址</span>
        {domainFailed ? <Button type="link" onClick={retryDomain}>站点加载失败，重试</Button> : domain ? <Typography.Link href={domain.default_portal_url || (domain.domain ? `https://${domain.domain}` : undefined)} target="_blank">{domain.domain || domain.default_portal_url || '尚未绑定域名'}</Typography.Link> : <Skeleton.Input active size="small" />}
      </div>
      <Space wrap className="site-plan-links"><Button type="link" onClick={() => navigate('/theme-settings')}>主题设置</Button><Button type="link" onClick={() => navigate('/domain-settings')}>域名设置</Button><Button type="link" onClick={() => navigate('/resources')}>资源管理</Button></Space>
      {data.has_demo_data && <div className="demo-data-action"><Button danger size="small" onClick={onClear}>清除演示数据</Button></div>}
    </Card>
  )
}

function StationWorkbench({ data, onDataChange }: { data: TenantDashboard; onDataChange: (data: TenantDashboard) => void }) {
  const navigate = useNavigate()
  const cards = stationDashboardCards(data)
  const [plan, setPlan] = useState<PlanUsage>()
  const [domain, setDomain] = useState<DomainBindStatus>()
  const [planFailed, setPlanFailed] = useState(false)
  const [domainFailed, setDomainFailed] = useState(false)

  const loadPlan = () => {
    setPlanFailed(false)
    void planApi.current().then(({ data: result }) => setPlan(result)).catch(() => setPlanFailed(true))
  }
  const loadDomain = () => {
    setDomainFailed(false)
    void domainApi.status().then(setDomain).catch(() => setDomainFailed(true))
  }
  useEffect(() => { loadPlan(); loadDomain() }, [])

  const clearDemoData = () => Modal.confirm({
    title: '清除演示数据？',
    content: '已登记的演示课程、成员和资源将被删除，此操作不可撤销。',
    okText: '确认清除',
    okButtonProps: { danger: true },
    cancelText: '取消',
    onOk: async () => {
      await tenantApi.clearDemoData()
      onDataChange({ ...data, has_demo_data: false })
      message.success('演示数据已清除')
    },
  })

  const comparison = cards[0].comparison
  const metric = (item: typeof cards[number], index: number) => (
    <div className="station-metric" key={item.title}>
      <Typography.Text type="secondary">{item.title}</Typography.Text>
      <strong>{item.value}</strong>
      {index === 0 && comparison ? <small className={`metric-${comparison.direction}`}>较昨日 {comparison.direction === 'up' ? '增加' : comparison.direction === 'down' ? '减少' : '持平'} {comparison.value}</small> : item.detail && <small>{item.detail}</small>}
    </div>
  )

  const quickActions = [
    { label: '添加学员', icon: <UserAddOutlined />, path: '/users?create=1' },
    { label: '新建课程', icon: <PlusOutlined />, path: '/courses?create=1' },
    { label: '上传资源', icon: <CloudUploadOutlined />, path: '/resources?upload=1' },
    { label: '管理官方课程', icon: <BookOutlined />, path: '/official-courses' },
  ]

  return (
    <>
      <PageHeader title="首页概览" description="查看本站今日运营数据与常用操作。" />
      <div className="station-dashboard-grid">
        <Card className="station-card station-metrics-card"><div className="station-metrics">{cards.slice(0, 3).map(metric)}</div></Card>
        <Card className="station-card station-metrics-card"><div className="station-metrics">{cards.slice(3).map((item, index) => metric(item, index + 3))}</div></Card>
        <Card title="快捷操作" className="station-card station-quick-card"><div className="quick-action-grid">{quickActions.map((action) => <Button key={action.path} className="quick-action" icon={action.icon} onClick={() => navigate(action.path)}>{action.label}</Button>)}</div></Card>
        <SitePlanCard plan={plan} domain={domain} planFailed={planFailed} domainFailed={domainFailed} retryPlan={loadPlan} retryDomain={loadDomain} data={data} onClear={clearDemoData} />
        <Card title="今日学习排行" className="station-card station-ranking-card">
          {data.today_learning_ranking.length ? <List dataSource={data.today_learning_ranking} renderItem={(item, index) => {
            const position = rankingPosition(index)
            return <List.Item><Space><span className={`ranking-position rank-${position.rank}`} aria-label={position.label}>{position.medal && <TrophyOutlined aria-hidden="true" />}<span>{position.label}</span></span><strong>{item.display_name}</strong></Space><Typography.Text type="secondary">{formatStudyDuration(item.duration_seconds)}</Typography.Text></List.Item>
          }} /> : <Empty description="今日暂无学习记录" />}
        </Card>
        <Card title="资源统计" className="station-card station-resource-card"><ResourceDonut data={data} /></Card>
      </div>
    </>
  )
}

export default function Dashboard() {
  const [data, setData] = useState<DashboardResponse>()
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const load = () => {
    setLoading(true)
    setFailed(false)
    void dashboardApi.get().then(({ data: response }) => setData(response)).catch(() => setFailed(true)).finally(() => setLoading(false))
  }
  useEffect(load, [])
  if (loading) return <div className="dashboard-skeleton"><Skeleton active /><Skeleton active /></div>
  if (failed || !data) return <Empty description="统计数据暂时不可用"><Button type="primary" onClick={load}>重新加载</Button></Empty>
  if (data.scope === 'platform') return <PlatformWorkbench data={data} />
  if (data.scope === 'instructor') return <InstructorWorkbench data={data} />
  return <StationWorkbench data={data} onDataChange={setData} />
}
