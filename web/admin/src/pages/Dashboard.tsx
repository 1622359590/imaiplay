import { BookOutlined, CheckCircleOutlined, ClockCircleOutlined, RiseOutlined, TeamOutlined, UserAddOutlined } from '@ant-design/icons'
import { Button, Card, Col, Empty, List, message, Modal, Progress, Row, Spin, Statistic, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { dashboardApi, type DashboardStats } from '../api/dashboard'
import PageHeader from '../components/PageHeader'
import { tenantApi } from '../api/tenant'
import { planApi } from '../api/plan'
import { tokenRole } from '../api/auth'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats>()
  const [loading, setLoading] = useState(true)
  const [clearing, setClearing] = useState(false)
  const [planUsage, setPlanUsage] = useState<{ plan: { name: string; storage_quota_bytes: number }; used_bytes: number; quota_bytes: number }>()
  const superadmin = tokenRole() === 'superadmin'

  useEffect(() => {
    dashboardApi.get()
      .then(({ data }) => setStats(data))
      .catch(() => setStats(undefined))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => { if (tokenRole() === 'tenant_admin') void planApi.current().then(({ data }) => setPlanUsage(data)).catch(() => undefined) }, [])

  if (loading) return <div className="center-spin"><Spin size="large" /></div>
  if (!stats) return <Empty description="统计数据暂时不可用" />

  if (superadmin && stats.platform) {
    const platformCards = [
      { title: '租户总数', value: stats.platform.tenant_count, icon: <TeamOutlined /> },
      { title: '活跃租户数', value: stats.platform.active_tenant_count, icon: <RiseOutlined /> },
      { title: '全平台学员', value: stats.platform.learner_count, icon: <UserAddOutlined /> },
      { title: '全平台课程', value: stats.platform.course_count, icon: <BookOutlined /> },
    ]
    return (
      <>
        <PageHeader title="平台概览" description="查看全平台租户、学员和课程运营情况。" />
        <Row gutter={[20, 20]}>
          {platformCards.map((item) => (
            <Col xs={24} sm={12} xl={6} key={item.title}>
              <Card className="stat-card"><div className="stat-icon">{item.icon}</div><Statistic title={item.title} value={item.value} /></Card>
            </Col>
          ))}
        </Row>
        <Card title="最近注册租户" style={{ marginTop: 20 }}>
          {stats.platform.recent_tenants.length > 0 ? (
            <List
              dataSource={stats.platform.recent_tenants}
              renderItem={(tenant) => (
                <List.Item>
                  <List.Item.Meta title={tenant.name} description={`${tenant.code} · ${new Date(tenant.created_at).toLocaleString()}`} />
                  <Tag color={tenant.status === 1 ? 'success' : 'default'}>{tenant.status === 1 ? '正常' : '停用'}</Tag>
                </List.Item>
              )}
            />
          ) : <Empty description="暂无注册租户" />}
        </Card>
      </>
    )
  }

  const cards = [
    { title: '学员总数', value: stats.user_count, icon: <TeamOutlined /> },
    {
      title: '课程总数',
      value: stats.course_count,
      suffix: `/ ${stats.published_course_count} 已发布`,
      icon: <BookOutlined />,
    },
    { title: '今日新增学员', value: stats.today_new_user_count, icon: <UserAddOutlined /> },
    { title: '今日学习人数', value: stats.today_learning_user_count, icon: <RiseOutlined /> },
    {
      title: '总学习时长',
      value: stats.total_learning_seconds / 3600,
      precision: 1,
      suffix: '小时',
      icon: <ClockCircleOutlined />,
    },
    {
      title: '课程完成率',
      value: stats.course_completion_rate * 100,
      precision: 1,
      suffix: '%',
      icon: <CheckCircleOutlined />,
    },
  ]

  return (
    <>
      <PageHeader title="工作台" description="欢迎回来，这里是平台今日运营概况。" />
      {planUsage && <Card style={{ marginBottom: 20 }} title={`当前套餐：${planUsage.plan.name}`}><Progress percent={planUsage.quota_bytes > 0 ? Math.min(100, Math.round(planUsage.used_bytes / planUsage.quota_bytes * 100)) : 0} format={() => planUsage.quota_bytes > 0 ? `${(planUsage.used_bytes / 1024 ** 2).toFixed(1)} MB / ${(planUsage.quota_bytes / 1024 ** 3).toFixed(1)} GB` : `${(planUsage.used_bytes / 1024 ** 2).toFixed(1)} MB / 不限额`} /></Card>}
      <Card style={{ marginBottom: 20 }}>
        <Typography.Text>当前空间包含一套示例课程和成员，可随时清除。</Typography.Text>
        <Button danger style={{ marginLeft: 16 }} loading={clearing} onClick={() => Modal.confirm({ title: '清除演示数据？', content: '课程、示例成员和示例资源将被删除，此操作不可撤销。', okText: '确认清除', cancelText: '取消', onOk: async () => { setClearing(true); try { await tenantApi.clearDemoData(); message.success('演示数据已清除'); window.location.reload() } finally { setClearing(false) } } })}>清除演示数据</Button>
      </Card>
      <Row gutter={[20, 20]}>
        {cards.map((item) => (
          <Col xs={24} sm={12} xl={8} key={item.title}>
            <Card className="stat-card">
              <div className="stat-icon">{item.icon}</div>
              <Statistic
                title={item.title}
                value={item.value}
                precision={item.precision}
                suffix={item.suffix}
              />
            </Card>
          </Col>
        ))}
      </Row>
      <Typography.Paragraph type="secondary" style={{ marginTop: 18 }}>
        学习时长按课时最后播放位置估算，完成率按有效报名学员计算。
      </Typography.Paragraph>
    </>
  )
}
