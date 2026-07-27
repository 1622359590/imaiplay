import { LockOutlined, MailOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography } from 'antd'
import { useState } from 'react'
import { useDispatch } from 'react-redux'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { login, type LoginPayload } from '../api/auth'
import { setSession } from '../store/userSlice'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const location = useLocation()

  const submit = async (values: LoginPayload) => {
    setLoading(true)
    try {
      const session = await login(values)
      dispatch(setSession(session))
      const target = (location.state as { from?: string } | null)?.from || '/'
      navigate(target, { replace: true })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <section className="login-intro">
        <div className="login-logo">I</div>
        <Typography.Title>让每一次学习，都推动组织成长</Typography.Title>
        <Typography.Paragraph>
          ImaiPlay 为企业提供课程、成员与学习运营的一体化管理体验。
        </Typography.Paragraph>
        <div className="login-grid" />
      </section>
      <main className="login-form-wrap">
        <Card className="login-card" bordered={false}>
          <Typography.Title level={2}>欢迎回来</Typography.Title>
          <Typography.Paragraph type="secondary">登录企业培训管理后台</Typography.Paragraph>
          <Form<LoginPayload> layout="vertical" size="large" onFinish={submit} requiredMark={false}>
            <Form.Item label="租户编码" name="tenant_code" rules={[{ required: true, message: '请输入租户编码' }]}>
              <Input prefix={<SafetyCertificateOutlined />} placeholder="例如：acme" autoComplete="organization" />
            </Form.Item>
            <Form.Item label="邮箱" name="email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
              <Input prefix={<MailOutlined />} placeholder="name@company.com" autoComplete="email" />
            </Form.Item>
            <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading} className="login-button">
              登录
            </Button>
          </Form>
          <Typography.Paragraph style={{ marginTop: 20, textAlign: 'center' }}>
            还没有企业账号？ <Link to="/register">开通租户</Link>
          </Typography.Paragraph>
        </Card>
      </main>
    </div>
  )
}
