import { useState } from 'react'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { KeyOutline, LockOutline, PhonebookOutline } from 'antd-mobile-icons'
import { Link, useNavigate } from 'react-router-dom'
import { forgotPassword, resetPassword } from '../api/auth'
import { useTenantTheme } from '../context/TenantThemeContext'

export function ForgotPasswordPage() {
  const [loading, setLoading] = useState(false)
  const [codeSent, setCodeSent] = useState(false)
  const theme = useTenantTheme()
  const navigate = useNavigate()

  const handleSubmit = async (values: {
    phone: string
    code?: string
    new_password?: string
  }) => {
    setLoading(true)
    try {
      if (!codeSent) {
        await forgotPassword(values.phone)
        setCodeSent(true)
        Toast.show({ icon: 'success', content: '如手机号已注册，验证码将发送到该手机' })
      } else {
        await resetPassword(values.phone, values.code || '', values.new_password || '')
        Toast.show({ icon: 'success', content: '密码已重置，请重新登录' })
        navigate(theme.loginPath, { replace: true })
      }
    } catch (error) {
      Toast.show({
        icon: 'fail',
        content: error instanceof Error ? error.message : '请求失败，请稍后重试',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-container reveal">
        <div className="login-brand">
          {theme.logo_url
            ? <img className="brand-logo-image" src={theme.logo_url} alt={`${theme.name} logo`} />
            : <div className="brand-logo">IP</div>}
          <strong>{theme.name}</strong>
        </div>
        <section className="login-card glass-card">
          <div className="login-panel-title">
            <h2 className="gradient-text">找回密码</h2>
            <p>{codeSent ? '输入短信验证码并设置新密码' : '输入企业账号绑定的手机号获取验证码'}</p>
          </div>
          <Form
            layout="horizontal"
            mode="card"
            onFinish={handleSubmit}
            footer={
              <Button
                className="btn-primary"
                block
                color="primary"
                size="large"
                loading={loading}
                type="submit"
              >
                {codeSent ? '重置密码' : '获取验证码'}
              </Button>
            }
          >
            <Form.Item
              name="phone"
              label={<PhonebookOutline className="input-icon" />}
              rules={[{ required: true, message: '请输入手机号' }]}
            >
              <Input className="dark-input" placeholder="手机号" clearable />
            </Form.Item>
            {codeSent && <>
              <Form.Item
                name="code"
                label={<KeyOutline className="input-icon" />}
                rules={[{ required: true, message: '请输入验证码' }]}
              >
                <Input className="dark-input" placeholder="短信验证码" clearable />
              </Form.Item>
              <Form.Item
                name="new_password"
                label={<LockOutline className="input-icon" />}
                rules={[
                  { required: true, message: '请输入新密码' },
                  { min: 8, message: '密码至少 8 位' },
                ]}
              >
                <Input className="dark-input" placeholder="至少 8 位新密码" type="password" clearable />
              </Form.Item>
            </>}
          </Form>
          <p className="login-help"><Link to={theme.loginPath}>返回登录</Link></p>
        </section>
      </div>
    </div>
  )
}
