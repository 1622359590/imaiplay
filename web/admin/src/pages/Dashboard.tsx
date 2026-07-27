import { BookOutlined, CheckCircleOutlined, ClockCircleOutlined, RiseOutlined, TeamOutlined, UserAddOutlined } from '@ant-design/icons'
import { Button, Card, Col, Empty, message, Modal, Row, Spin, Statistic, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { dashboardApi, type DashboardStats } from '../api/dashboard'
import PageHeader from '../components/PageHeader'
import { tenantApi } from '../api/tenant'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats>()
  const [loading, setLoading] = useState(true)
  const [clearing, setClearing] = useState(false)

  useEffect(() => {
    dashboardApi.get()
      .then(({ data }) => setStats(data))
      .catch(() => setStats(undefined))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="center-spin"><Spin size="large" /></div>
  if (!stats) return <Empty description="统计数据暂时不可用" />

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
