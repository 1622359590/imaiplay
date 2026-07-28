import { LockOutlined, MailOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography, Radio, message } from 'antd'
import { useState } from 'react'
import { useDispatch } from 'react-redux'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { login, loginWithCode, sendLoginCode, type LoginPayload } from '../api/auth'
import { setSession } from '../store/userSlice'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const [mode, setMode] = useState<'password' | 'code'>('password')
  const [form] = Form.useForm<LoginPayload & { code?: string }>()
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const location = useLocation()

  const submit = async (values: LoginPayload & { code?: string }) => {
    setLoading(true)
    try {
      const session = mode === 'password' ? await login(values) : await loginWithCode({ tenant_code: values.tenant_code, phone: values.identifier, code: values.code || '' })
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
          <Radio.Group value={mode} onChange={(event) => setMode(event.target.value)} options={[{ label: '密码登录', value: 'password' }, { label: '验证码登录', value: 'code' }]} style={{ marginBottom: 16 }} />
          <Form form={form} layout="vertical" size="large" onFinish={submit} requiredMark={false}>
            <Form.Item label="租户编码" name="tenant_code" rules={[{ required: true, message: '请输入租户编码' }]}>
              <Input prefix={<SafetyCertificateOutlined />} placeholder="例如：acme" autoComplete="organization" />
            </Form.Item>
            <Form.Item label="邮箱或手机号" name="identifier" rules={[{ required: true, message: '请输入邮箱或手机号' }]}>
              <Input prefix={<MailOutlined />} placeholder="name@company.com 或 13800138000" autoComplete="username" />
            </Form.Item>
            {mode === 'password' ? <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}><Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" /></Form.Item> : <Form.Item label="验证码" name="code" rules={[{ required: true, message: '请输入验证码' }]}><Input prefix={<LockOutlined />} placeholder="6 位验证码" addonAfter={<Button type="link" onClick={async () => { const values = await form.validateFields(['tenant_code', 'identifier']); await sendLoginCode(values.tenant_code, values.identifier); message.success('验证码已发送') }}>发送验证码</Button>} /></Form.Item>}
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
