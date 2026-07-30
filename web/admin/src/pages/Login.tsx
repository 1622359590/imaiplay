import { ApartmentOutlined, LockOutlined, MailOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography, message } from 'antd'
import { isAxiosError } from 'axios'
import { useState } from 'react'
import { useDispatch } from 'react-redux'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { login, type LoginPayload } from '../api/auth'
import { setSession } from '../store/userSlice'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const [needTenantCode, setNeedTenantCode] = useState(false)
  const [form] = Form.useForm<LoginPayload>()
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
    } catch (error) {
      if (
        isAxiosError<{ error?: string }>(error) &&
        error.response?.data?.error === 'account_exists_multiple_tenants'
      ) {
        setNeedTenantCode(true)
        message.warning('检测到该账号存在于多个企业，请输入租户编码')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page admin-login-page">
      <main className="admin-login-container">
        <div className="admin-login-brand">
          <div className="login-logo">I</div>
          <strong>iMaiPlay</strong>
        </div>
        <Card className="login-card admin-login-card" bordered={false}>
          <Typography.Title level={2} className="admin-login-title">欢迎回来</Typography.Title>
          <Typography.Paragraph type="secondary">登录企业培训管理后台</Typography.Paragraph>
          <Form form={form} layout="vertical" size="large" onFinish={submit} requiredMark={false}>
            <Form.Item label="邮箱或手机号" name="identifier" rules={[{ required: true, message: '请输入邮箱或手机号' }]}>
              <Input prefix={<MailOutlined />} placeholder="name@company.com 或 13800138000" autoComplete="username" />
            </Form.Item>
            <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}><Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" /></Form.Item>
            {needTenantCode && (
              <Form.Item
                label="租户编码"
                name="tenant_code"
                rules={[{ required: true, message: '请输入租户编码' }]}
              >
                <Input prefix={<ApartmentOutlined />} placeholder="请输入租户编码" autoFocus />
              </Form.Item>
            )}
            <Button type="primary" htmlType="submit" block loading={loading} className="login-button">
              登录
            </Button>
          </Form>
          <Typography.Paragraph style={{ textAlign: 'center' }}><Link to="/forgot-password">忘记密码？</Link></Typography.Paragraph>
          <Typography.Paragraph style={{ marginTop: 20, textAlign: 'center' }}>
            还没有企业账号？ <Link to="/register">开通租户</Link>
          </Typography.Paragraph>
        </Card>
      </main>
    </div>
  )
}
