import { LockOutlined, MailOutlined, ReadOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, message, Typography } from 'antd';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { userFacingErrorMessage } from '@imaiplay/shared/api/errors';
import type { LoginValues } from '../api/auth';
import { useAuth } from '../context/AuthContext';
import { usePortal } from '../context/PortalContext';
import {
  boundPortalLoginPath,
  performLoginNavigation,
  portalRoutePath,
} from '../utils/portalRouting';
import {
  isPortalSessionToken,
  portalSessionMatchesPortal,
  readPortalAccessToken,
  readPortalTenantCode,
} from '../api/authSession';

export function LoginPage() {
  const { authenticated, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const { portal, mode, tenantCode } = usePortal();
  const [form] = Form.useForm<LoginValues>();

  useEffect(() => {
    document.title = portal?.browser_title?.trim() || (portal ? `${portal.name} | 企业学习中心` : 'iMaiPlay 企业学习中心');
  }, [portal]);

  const platformPortalPath = mode === 'platform' &&
    isPortalSessionToken(readPortalAccessToken())
    ? boundPortalLoginPath(readPortalTenantCode())
    : undefined;
  const sessionMatchesPortal = portal
    ? portalSessionMatchesPortal(portal)
    : false;

  if (authenticated && (mode === 'platform' ? platformPortalPath : sessionMatchesPortal)) {
    const destination = mode === 'platform'
      ? platformPortalPath!
      : portalRoutePath(mode, tenantCode, '/');
    return <Navigate to={destination} replace />;
  }

  const handleSubmit = async (values: LoginValues) => {
    try {
      const outcome = await login(values, mode, tenantCode);
      if (outcome.requiresSelection) {
        navigate('/select-organization', { replace: true });
        return;
      }
      const from = (location.state as { from?: string } | null)?.from ?? '/';
      performLoginNavigation(outcome.redirect ?? from, navigate);
    } catch (error) {
      message.error(userFacingErrorMessage(error, '登录失败，请稍后重试'));
    }
  };

  return (
    <div className="login-page">
      <div className="login-container reveal">
        <div className="login-brand">
          {portal?.logo_url
            ? <img className="login-brand-logo" src={portal.logo_url} alt={`${portal.name} logo`} />
            : <span className="login-brand-mark"><ReadOutlined /></span>}
          <span>{portal?.name || 'iMaiPlay'}</span>
        </div>
        <Card className="login-card glass-card" variant="borderless">
          <Typography.Title level={2} className="gradient-text">{portal?.welcome_text || '欢迎回来'}</Typography.Title>
          <Typography.Paragraph type="secondary">
            {portal ? `登录 ${portal.name}，继续你的成长旅程` : '登录企业学习中心，继续你的成长旅程'}
          </Typography.Paragraph>
          <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark={false}>
            <Form.Item
              label="邮箱"
              name="identifier"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效邮箱' },
              ]}
            >
              <Input className="dark-input" size="large" prefix={<MailOutlined />} placeholder="name@company.com" autoFocus />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password className="dark-input" size="large" prefix={<LockOutlined />} placeholder="请输入密码" />
            </Form.Item>
            <Form.Item>
              <Button className="btn-primary" type="primary" htmlType="submit" size="large" block>
                登录学习中心
              </Button>
            </Form.Item>
          </Form>
          <div className="login-footer">
            <a href="/admin/forgot-password">忘记密码？</a>
          </div>
        </Card>
      </div>
    </div>
  );
}
