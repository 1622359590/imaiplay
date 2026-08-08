import { useState } from 'react'
import { Button, Form, Input, Toast } from 'antd-mobile'
import { FileOutline, LockOutline, MailOutline } from 'antd-mobile-icons'
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { login, type LoginPayload } from '../api/auth'
import { isValidPortalSession, readPortalAccessToken, readPortalTenantCode } from '../api/authSession'
import { useTenantTheme } from '../context/TenantThemeContext'
import { userFacingErrorMessage } from '@imaiplay/shared/api/errors'

export function LoginPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const theme = useTenantTheme()
  const destination =
    (location.state as { from?: string } | null)?.from ??
    theme.routePath('/')

  if (
    theme.portal &&
    isValidPortalSession(readPortalAccessToken(), theme.portal.tenant_id) &&
    readPortalTenantCode() === theme.portal.code.toLowerCase()
  ) {
    return <Navigate to={theme.routePath('/')} replace />
  }

  const handleSubmit = async (values: LoginPayload) => {
    if (!theme.portal) {
      Toast.show({ icon: 'fail', content: '企业门户尚未就绪' })
      return
    }
    setLoading(true)
    try {
      await login(values, theme.portal)
      Toast.show({ icon: 'success', content: '登录成功' })
      navigate(destination, { replace: true })
    } catch (error) {
      Toast.show({
        icon: 'fail',
        content: userFacingErrorMessage(error, '登录失败，请稍后重试'),
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
            : <div className="brand-logo"><FileOutline /></div>}
          <strong>{theme.name}</strong>
        </div>
        <section className="login-card glass-card">
          <div className="login-panel-title">
            <h2 className="gradient-text">{theme.welcome_text || '欢迎回来'}</h2>
            <p>登录 {theme.name}，继续你的成长旅程</p>
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
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效邮箱' },
              ]}
            >
              <Input className="dark-input" placeholder="邮箱" clearable />
            </Form.Item>
            <Form.Item
              name="password"
              label={<LockOutline className="input-icon" />}
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input className="dark-input" placeholder="登录密码" type="password" clearable />
            </Form.Item>
          </Form>
          <p className="login-help">
            <Link to={theme.forgotPasswordPath}>忘记密码？</Link>
          </p>
        </section>
      </div>
    </div>
  )
}
