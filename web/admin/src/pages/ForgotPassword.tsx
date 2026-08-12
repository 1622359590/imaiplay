import { LockOutlined, MobileOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography, message } from 'antd'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { forgotPassword, resetPassword } from '../api/auth'
import { useAdminTheme } from '../context/AdminThemeContext'

export default function ForgotPassword() {
  const [sent, setSent] = useState(false)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const theme = useAdminTheme()

  const requestCode = async (values: { phone: string }) => {
    setLoading(true)
    try {
      await forgotPassword(values.phone)
      setSent(true)
      message.success('验证码已发送，请查看短信或开发日志')
    } finally {
      setLoading(false)
    }
  }

  const reset = async (values: { phone: string; code: string; new_password: string }) => {
    setLoading(true)
    try {
      await resetPassword(values.phone, values.code, values.new_password)
      message.success('密码已重置')
      navigate('/login')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page admin-login-page admin-auth-shell">
      <aside className="auth-brand-panel auth-brand-panel-recovery" aria-label={`${theme.brandName} 账户恢复`}>
        <div className="admin-login-brand">
          {theme.logoURL
            ? <img className="auth-brand-logo" src={theme.logoURL} alt={`${theme.brandName} Logo`} />
            : <div className="login-logo" aria-hidden="true">I</div>}
          <strong>{theme.brandName}</strong>
        </div>
        <div className="auth-brand-copy">
          <span className="auth-brand-eyebrow">安全中心</span>
          <h1>安全找回你的管理账号</h1>
          <p>通过已绑定手机号完成身份验证并设置新密码。</p>
        </div>
      </aside>
      <main className="admin-login-container auth-form-panel">
        <Card className="login-card admin-login-card" variant="borderless">
          <Typography.Title level={2} className="admin-login-title">找回密码</Typography.Title>
          <Typography.Paragraph type="secondary">通过手机号验证码重置密码</Typography.Paragraph>
          <Form layout="vertical" size="large" onFinish={sent ? reset : requestCode} requiredMark={false}>
            <Form.Item label="手机号" name="phone" rules={[{ required: true, message: '请输入手机号' }]}>
              <Input prefix={<MobileOutlined />} autoComplete="tel" />
            </Form.Item>
            {sent && (
              <>
                <Form.Item label="验证码" name="code" rules={[{ required: true, message: '请输入验证码' }]}>
                  <Input prefix={<SafetyCertificateOutlined />} inputMode="numeric" />
                </Form.Item>
                <Form.Item label="新密码" name="new_password" rules={[{ required: true, min: 8, message: '密码至少 8 位' }]}>
                  <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
                </Form.Item>
              </>
            )}
            <Button type="primary" htmlType="submit" block loading={loading} className="login-button">
              {sent ? '重置密码' : '发送验证码'}
            </Button>
          </Form>
          <Typography.Paragraph className="auth-link-row auth-register-row"><Link to="/login">返回登录</Link></Typography.Paragraph>
        </Card>
      </main>
    </div>
  )
}
