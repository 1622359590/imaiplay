import { ApartmentOutlined, LockOutlined, MailOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography, message } from 'antd'
import { useRef, useState } from 'react'
import { useDispatch } from 'react-redux'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { userFacingErrorMessage } from '@imaiplay/shared/api/errors'
import {
  login,
  persistAdminLogin,
  selectTenant,
  type LoginPayload,
  type OrganizationOption,
  type TenantSelectionRequired,
} from '../api/auth'
import { useAdminTheme } from '../context/AdminThemeContext'
import { setSession } from '../store/userSlice'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const [pendingSelection, setPendingSelection] = useState<TenantSelectionRequired>()
  const submittingRef = useRef(false)
  const [form] = Form.useForm<LoginPayload>()
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const theme = useAdminTheme()

  const completeLogin = (result: Parameters<typeof persistAdminLogin>[0]) => {
    const session = persistAdminLogin(result)
    dispatch(setSession(session))
    const target = (location.state as { from?: string } | null)?.from || '/'
    navigate(target, { replace: true })
  }

  const submit = async (values: LoginPayload) => {
    if (submittingRef.current) return
    submittingRef.current = true
    setLoading(true)
    try {
      const result = await login(values)
      if (result.requires_tenant_selection) {
        setPendingSelection(result)
        return
      }
      if (result.user.role === 'learner') {
        message.info('请前往学习门户登录')
        return
      }
      completeLogin(result)
    } catch (error) {
      message.error(userFacingErrorMessage(error, '登录失败，请稍后重试'))
    } finally {
      setLoading(false)
      submittingRef.current = false
    }
  }

  const selectOrganization = async (organization: OrganizationOption) => {
    if (organization.role === 'learner') {
      message.info('请前往学习门户登录')
      return
    }
    if (!pendingSelection) return
    if (submittingRef.current) return
    submittingRef.current = true
    setLoading(true)
    try {
      const result = await selectTenant(pendingSelection.selection_token, organization.code)
      if (result.user.role === 'learner') {
        message.info('请前往学习门户登录')
        return
      }
      completeLogin(result)
    } catch (error) {
      message.error(userFacingErrorMessage(error, '选择企业失败，请重新登录'))
    } finally {
      setLoading(false)
      submittingRef.current = false
    }
  }

  return (
    <div className="login-page admin-login-page admin-auth-shell">
      <aside className="auth-brand-panel" aria-label={`${theme.brandName} 管理后台`}>
        <div className="admin-login-brand">
          {theme.logoURL
            ? <img className="auth-brand-logo" src={theme.logoURL} alt={`${theme.brandName} Logo`} />
            : <div className="login-logo" aria-hidden="true">I</div>}
          <strong>{theme.brandName}</strong>
        </div>
        <div className="auth-brand-copy">
          <span className="auth-brand-eyebrow">企业学习管理平台</span>
          <h1>让培训管理更清晰、更高效</h1>
          <p>统一管理课程、成员与学习数据，随时掌握组织培训进展。</p>
        </div>
      </aside>
      <main className="admin-login-container auth-form-panel">
        <Card className="login-card admin-login-card" variant="borderless">
          <Typography.Title level={2} className="admin-login-title">
            {pendingSelection ? '选择要管理的企业' : '欢迎回来'}
          </Typography.Title>
          <Typography.Paragraph type="secondary">
            {pendingSelection ? '请选择一个拥有管理权限的企业继续' : '登录企业培训管理后台'}
          </Typography.Paragraph>
          {pendingSelection ? (
            <div className="organization-picker" aria-label="企业选择">
              {pendingSelection.organizations.map((organization) => {
                const learner = organization.role === 'learner'
                return (
                  <Button
                    key={`${organization.code}:${organization.role}`}
                    className="organization-card"
                    disabled={loading}
                    onClick={() => void selectOrganization(organization)}
                  >
                    <span className="organization-card-icon"><ApartmentOutlined /></span>
                    <span className="organization-card-copy">
                      <strong>{organization.name}</strong>
                      <small>{learner ? '学员账号，请前往学习门户登录' : `角色：${organization.role}`}</small>
                    </span>
                  </Button>
                )
              })}
              <Button type="link" disabled={loading} onClick={() => setPendingSelection(undefined)}>返回重新登录</Button>
            </div>
          ) : (
            <Form form={form} layout="vertical" size="large" onFinish={submit} requiredMark={false}>
              <Form.Item
                label="邮箱"
                name="identifier"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效邮箱' },
                ]}
              >
                <Input prefix={<MailOutlined />} placeholder="name@company.com" autoComplete="username" />
              </Form.Item>
              <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
                <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
              </Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading} className="login-button">登录</Button>
            </Form>
          )}
          {!pendingSelection && <Typography.Paragraph className="auth-link-row"><Link to="/forgot-password">忘记密码？</Link></Typography.Paragraph>}
          {!pendingSelection && <Typography.Paragraph className="auth-link-row auth-register-row">
            还没有企业账号？ <Link to="/register">开通租户</Link>
          </Typography.Paragraph>}
        </Card>
      </main>
    </div>
  )
}
