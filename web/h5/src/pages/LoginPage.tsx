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
      <div className="login-orb login-orb-one" />
      <div className="login-orb login-orb-two" />
      <section className="login-brand">
        <div className="brand-logo">IP</div>
        <p>IMAI PLAY</p>
        <h1>开启成长新旅程</h1>
        <span>企业人才学习与发展平台</span>
      </section>
      <section className="login-panel">
        <div className="login-panel-title">
          <h2>欢迎登录</h2>
          <p>请使用企业学习账号继续</p>
        </div>
        <Form
          layout="horizontal"
          mode="card"
          onFinish={handleSubmit}
          footer={
            <Button block color="primary" size="large" loading={loading} type="submit">
              登录学习中心
            </Button>
          }
        >
          <Form.Item
            name="identifier"
            label={<MailOutline className="input-icon" />}
            rules={[{ required: true, message: '请输入手机号或邮箱' }]}
          >
            <Input placeholder="手机号或邮箱" clearable />
          </Form.Item>
          <Form.Item
            name="password"
            label={<LockOutline className="input-icon" />}
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input placeholder="登录密码" type="password" clearable />
          </Form.Item>
        </Form>
        <p className="login-help">遇到问题？请联系企业培训管理员</p>
      </section>
    </div>
  )
}
