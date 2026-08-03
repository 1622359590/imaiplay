import { LockOutlined, MailOutlined, MobileOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Space, Typography, message } from 'antd'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { tenantApi, type RegisterTenantPayload } from '../api/tenant'
import { useDispatch } from 'react-redux'
import { setSession } from '../store/userSlice'
import { persistAdminLogin } from '../api/auth'
import { domainApi, type DomainBindStatus } from '../api/domain'
import { portalURLAfterRegistration } from '../utils/domainStatus'

interface RegisterForm extends RegisterTenantPayload { confirm_password: string }

export default function Register() {
  const [loading, setLoading] = useState(false)
  const [portalURL, setPortalURL] = useState<string>()
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const submit = async (values: RegisterForm) => {
    setLoading(true)
    try {
      const { data } = await tenantApi.register(values)
      const session = persistAdminLogin(data)
      dispatch(setSession(session))
      let domainStatus: DomainBindStatus | undefined
      try {
        domainStatus = await domainApi.status()
      } catch {
        // Registration has already succeeded. Keep the new session and show
        // the permanent production portal instead of asking users to retry.
      }
      setPortalURL(portalURLAfterRegistration(domainStatus, data.tenant.code))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <section className="login-intro">
        <div className="login-logo">I</div>
        <Typography.Title>从今天开始，建立学习型组织</Typography.Title>
        <Typography.Paragraph>开通企业空间，立即体验 ImaiPlay 的培训管理能力。</Typography.Paragraph>
        <div className="login-grid" />
      </section>
      <main className="login-form-wrap">
        <Card className="login-card" bordered={false}>
          {portalURL ? <>
            <Typography.Title level={2}>企业空间已开通</Typography.Title>
            <Typography.Paragraph type="secondary">学习门户已立即可用。你可将这个地址发给组织内的学员。</Typography.Paragraph>
            <Form layout="vertical" size="large">
              <Form.Item label="默认学习门户">
                <Input readOnly value={portalURL} addonAfter={<Button type="link" onClick={() => void navigator.clipboard.writeText(portalURL).then(() => message.success('学习门户地址已复制')).catch(() => message.error('复制失败，请手动复制'))}>复制</Button>} />
              </Form.Item>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Button type="primary" block onClick={() => navigate('/', { replace: true })}>进入管理后台</Button>
                <Typography.Text type="secondary">之后可在“域名设置”中选配自定义品牌域名。</Typography.Text>
              </Space>
            </Form>
          </> : <>
            <Typography.Title level={2}>开通企业空间</Typography.Title>
            <Typography.Paragraph type="secondary">填写信息后即可立即进入管理后台</Typography.Paragraph>
            <Form<RegisterForm> layout="vertical" size="large" onFinish={submit} requiredMark={false}>
            <Form.Item label="组织名称" name="organization_name" rules={[{ required: true, message: '请输入组织名称' }]}>
              <Input prefix={<TeamOutlined />} placeholder="例如：Acme 公司" />
            </Form.Item>
            <Form.Item label="管理员邮箱" name="admin_email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
              <Input prefix={<MailOutlined />} placeholder="name@company.com" autoComplete="email" />
            </Form.Item>
            <Form.Item label="管理员姓名" name="admin_name" rules={[{ required: true, message: '请输入姓名' }]}>
              <Input prefix={<UserOutlined />} placeholder="例如：张三" autoComplete="name" />
            </Form.Item>
            <Form.Item label="管理员手机号（可选）" name="phone"><Input prefix={<MobileOutlined />} placeholder="用于找回密码" /></Form.Item>
            <Form.Item label="密码" name="password" rules={[{ required: true, min: 8, message: '密码至少 8 位' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="至少 8 位" autoComplete="new-password" />
            </Form.Item>
            <Form.Item label="确认密码" name="confirm_password" dependencies={['password']} rules={[{ required: true, message: '请确认密码' }, ({ getFieldValue }) => ({ validator(_, value) { return !value || getFieldValue('password') === value ? Promise.resolve() : Promise.reject(new Error('两次密码不一致')) } })]}>
              <Input.Password prefix={<LockOutlined />} placeholder="再次输入密码" autoComplete="new-password" />
            </Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>立即开通</Button>
            </Form>
            <Typography.Paragraph style={{ marginTop: 20, textAlign: 'center' }}>
              已有企业账号？ <Link to="/login">返回登录</Link>
            </Typography.Paragraph>
          </>}
        </Card>
      </main>
    </div>
  )
}
