import { BookOutlined, CheckCircleOutlined, RiseOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons'
import { Card, Col, Progress, Row, Space, Statistic, Tag, Timeline, Typography } from 'antd'
import PageHeader from '../components/PageHeader'

const stats = [
  { title: '活跃租户', value: 28, icon: <TeamOutlined />, trend: '+8.2%' },
  { title: '平台用户', value: 12460, icon: <UserOutlined />, trend: '+12.5%' },
  { title: '在线课程', value: 386, icon: <BookOutlined />, trend: '+6.8%' },
  { title: '本月完课', value: 3254, icon: <CheckCircleOutlined />, trend: '+18.3%' },
]

export default function Dashboard() {
  return (
    <>
      <PageHeader title="工作台" description="欢迎回来，这里是平台今日运营概况。" />
      <Row gutter={[20, 20]}>
        {stats.map((item) => (
          <Col xs={24} sm={12} xl={6} key={item.title}>
            <Card className="stat-card">
              <div className="stat-icon">{item.icon}</div>
              <Statistic title={item.title} value={item.value} />
              <Tag color="success" bordered={false} icon={<RiseOutlined />}>{item.trend} 较上月</Tag>
            </Card>
          </Col>
        ))}
        <Col xs={24} lg={16}>
          <Card title="学习完成情况" className="dashboard-card">
            <div className="progress-row">
              <span>新员工入职培训</span><Progress percent={86} strokeColor="#1769e0" />
            </div>
            <div className="progress-row">
              <span>信息安全合规</span><Progress percent={72} strokeColor="#438cf2" />
            </div>
            <div className="progress-row">
              <span>管理者领导力</span><Progress percent={64} strokeColor="#69a3f3" />
            </div>
            <div className="progress-row">
              <span>产品知识认证</span><Progress percent={91} strokeColor="#1769e0" />
            </div>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="最近动态" className="dashboard-card">
            <Timeline
              items={[
                { color: 'blue', children: <Space direction="vertical" size={0}><Typography.Text>课程《产品知识认证》已发布</Typography.Text><Typography.Text type="secondary">10 分钟前</Typography.Text></Space> },
                { color: 'green', children: <Space direction="vertical" size={0}><Typography.Text>新增租户「华东培训中心」</Typography.Text><Typography.Text type="secondary">1 小时前</Typography.Text></Space> },
                { color: 'gray', children: <Space direction="vertical" size={0}><Typography.Text>批量导入 128 名学员</Typography.Text><Typography.Text type="secondary">昨天 16:20</Typography.Text></Space> },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </>
  )
}
