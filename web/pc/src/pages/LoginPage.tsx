import { LockOutlined, MailOutlined, ReadOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, message, Typography } from 'antd';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import type { LoginValues } from '../api/auth';
import { useAuth } from '../context/AuthContext';

export function LoginPage() {
  const { authenticated, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm<LoginValues>();

  if (authenticated) {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (values: LoginValues) => {
    try {
      await login(values);
      const from = (location.state as { from?: string } | null)?.from ?? '/';
      navigate(from, { replace: true });
    } catch (error) {
      if (error instanceof Error && error.message === '请使用学员账号登录') {
        message.error(error.message);
      }
    }
  };

  return (
    <div className="login-page">
      <div className="login-container reveal">
        <div className="login-brand">
          <span className="login-brand-mark"><ReadOutlined /></span>
          <span>iMaiPlay</span>
        </div>
        <Card className="login-card glass-card" bordered={false}>
          <Typography.Title level={2} className="gradient-text">欢迎回来</Typography.Title>
          <Typography.Paragraph type="secondary">
            登录企业学习中心，继续你的成长旅程
          </Typography.Paragraph>
          <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark={false}>
            <Form.Item
              label="手机号或邮箱"
              name="identifier"
              rules={[{ required: true, message: '请输入手机号或邮箱' }]}
            >
              <Input className="dark-input" size="large" prefix={<MailOutlined />} placeholder="name@company.com 或 13800138000" autoFocus />
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
