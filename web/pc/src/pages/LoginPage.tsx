import { CheckOutlined, LockOutlined, MailOutlined, MobileOutlined, ReadOutlined } from '@ant-design/icons';
import { Button, Checkbox, Form, Input, message } from 'antd';
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
      <section className="login-hero">
        <div className="login-hero-orb login-hero-orb-top" />
        <div className="login-hero-orb login-hero-orb-bottom" />
        <div className="login-brand">
          {portal?.logo_url
            ? <img className="login-brand-logo" src={portal.logo_url} alt={`${portal.name} logo`} />
            : <span className="login-brand-mark"><ReadOutlined /></span>}
          <span>{portal?.name || 'iMaiPlay'}</span>
        </div>
        <div className="login-hero-copy">
          <span>ENTERPRISE LEARNING</span>
          <h1>{portal?.welcome_text || '让学习成为组织持续成长的力量'}</h1>
          <p>一个可信赖的企业学习空间，帮助你清晰规划、专注学习并持续获得进步。</p>
          <ul>
            {['集中管理你的学习任务', '随时延续上次学习进度', '清晰掌握课程完成情况', '沉淀可持续的成长记录'].map((feature) => <li key={feature}><CheckOutlined />{feature}</li>)}
          </ul>
        </div>
      </section>
      <section className="login-form-panel">
        <div className="login-container reveal">
          <div className="login-card">
            <header><span>欢迎回来</span><h2>登录学习账户</h2><p>{portal ? `进入 ${portal.name}，继续你的成长旅程` : '登录企业学习中心，继续你的成长旅程'}</p></header>
            <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark={false}>
              <Form.Item
                label="邮箱"
                name="identifier"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效邮箱' },
                ]}
              >
                <Input size="large" prefix={<MailOutlined />} placeholder="name@company.com" autoFocus />
              </Form.Item>
              <Form.Item
                label="密码"
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
              >
                <Input.Password size="large" prefix={<LockOutlined />} placeholder="请输入密码" />
              </Form.Item>
              <div className="login-form-options"><Checkbox>记住我</Checkbox><a href="/admin/forgot-password">忘记密码？</a></div>
              <Form.Item>
                <Button type="primary" htmlType="submit" size="large" block>
                  登录学习中心
                </Button>
              </Form.Item>
            </Form>
            <div className="login-divider"><span>或</span></div>
            <Button className="sms-login-button" icon={<MobileOutlined />} size="large" block>短信登录</Button>
            <p className="login-footer">还没有账户？请联系企业管理员</p>
          </div>
        </div>
      </section>
    </div>
  );
}
