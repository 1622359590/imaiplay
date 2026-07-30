import { useState } from 'react'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { LockOutline, MailOutline } from 'antd-mobile-icons'
import { useLocation, useNavigate } from 'react-router-dom'
import { login, type LoginPayload } from '../api/auth'

export function LoginPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const destination = (location.state as { from?: string } | null)?.from ?? '/'

  const handleSubmit = async (values: LoginPayload) => {
    setLoading(true)
    try {
      await login(values)
      Toast.show({ icon: 'success', content: '登录成功' })
      navigate(destination, { replace: true })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-container reveal">
        <div className="login-brand">
          <div className="brand-logo">IP</div>
          <strong>iMaiPlay</strong>
        </div>
        <section className="login-card glass-card">
          <div className="login-panel-title">
            <h2 className="gradient-text">欢迎回来</h2>
            <p>登录企业学习中心，继续你的成长旅程</p>
          </div>
          <Form
            layout="horizontal"
            mode="card"
            onFinish={handleSubmit}
            footer={
              <Button className="btn-primary" block color="primary" size="large" loading={loading} type="submit">
                登录学习中心
              </Button>
            }
          >
            <Form.Item
              name="identifier"
              label={<MailOutline className="input-icon" />}
              rules={[{ required: true, message: '请输入手机号或邮箱' }]}
            >
              <Input className="dark-input" placeholder="手机号或邮箱" clearable />
            </Form.Item>
            <Form.Item
              name="password"
              label={<LockOutline className="input-icon" />}
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input className="dark-input" placeholder="登录密码" type="password" clearable />
            </Form.Item>
          </Form>
          <p className="login-help"><a href="/admin/forgot-password">忘记密码？</a></p>
        </section>
      </div>
    </div>
  )
}
