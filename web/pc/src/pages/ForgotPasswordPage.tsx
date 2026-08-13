import {
  CheckOutlined,
  LockOutlined,
  MobileOutlined,
  ReadOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { Button, Form, Input, message } from 'antd';
import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { userFacingErrorMessage } from '@imaiplay/shared/api/errors';
import { forgotPassword, resetPassword } from '../api/auth';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';

interface RecoveryValues {
  phone: string;
  code?: string;
  new_password?: string;
}

export function ForgotPasswordPage() {
  const [loading, setLoading] = useState(false);
  const [codeSent, setCodeSent] = useState(false);
  const [form] = Form.useForm<RecoveryValues>();
  const navigate = useNavigate();
  const { portal, mode, tenantCode } = usePortal();
  const loginPath = portalRoutePath(mode, tenantCode, '/login');

  useEffect(() => {
    document.title = portal
      ? `找回密码 | ${portal.name}`
      : '找回密码 | iMaiPlay';
  }, [portal]);

  const handleSubmit = async (values: RecoveryValues) => {
    setLoading(true);
    try {
      if (!codeSent) {
        await forgotPassword(values.phone);
        setCodeSent(true);
        message.success('如手机号已注册，验证码将发送到该手机');
        return;
      }

      await resetPassword(values.phone, values.code || '', values.new_password || '');
      message.success('密码已重置，请重新登录');
      navigate(loginPath, { replace: true });
    } catch (error) {
      message.error(userFacingErrorMessage(error, '操作失败，请稍后重试'));
    } finally {
      setLoading(false);
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
          <span>ACCOUNT RECOVERY</span>
          <h1>安全找回你的学习账户</h1>
          <p>通过企业账号绑定的手机号完成验证，重新进入学习中心。</p>
          <ul>
            {['验证绑定手机号', '接收一次性验证码', '设置全新的登录密码'].map((feature) => (
              <li key={feature}><CheckOutlined />{feature}</li>
            ))}
          </ul>
        </div>
      </section>
      <section className="login-form-panel">
        <div className="login-container reveal">
          <div className="login-card">
            <header>
              <span>找回账户</span>
              <h2>{codeSent ? '设置新密码' : '验证手机号'}</h2>
              <p>{codeSent ? '输入短信验证码并设置新密码' : '输入企业账号绑定的手机号获取验证码'}</p>
            </header>
            <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark={false}>
              <Form.Item
                label="手机号"
                name="phone"
                rules={[{ required: true, message: '请输入手机号' }]}
              >
                <Input size="large" prefix={<MobileOutlined />} placeholder="请输入绑定手机号" autoFocus />
              </Form.Item>
              {codeSent && <>
                <Form.Item
                  label="短信验证码"
                  name="code"
                  rules={[{ required: true, message: '请输入验证码' }]}
                >
                  <Input size="large" prefix={<SafetyCertificateOutlined />} placeholder="请输入短信验证码" />
                </Form.Item>
                <Form.Item
                  label="新密码"
                  name="new_password"
                  rules={[
                    { required: true, message: '请输入新密码' },
                    { min: 8, message: '密码至少 8 位' },
                  ]}
                >
                  <Input.Password size="large" prefix={<LockOutlined />} placeholder="至少 8 位新密码" />
                </Form.Item>
              </>}
              <Form.Item>
                <Button type="primary" htmlType="submit" size="large" loading={loading} block>
                  {codeSent ? '重置密码' : '获取验证码'}
                </Button>
              </Form.Item>
            </Form>
            <p className="login-footer"><Link to={loginPath}>返回登录</Link></p>
          </div>
        </div>
      </section>
    </div>
  );
}
