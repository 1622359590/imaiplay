import { LockOutlined, MailOutlined, ReadOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, Typography } from 'antd';
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
    await login(values);
    const from = (location.state as { from?: string } | null)?.from ?? '/';
    navigate(from, { replace: true });
  };

  return (
    <div className="login-page">
      <section className="login-intro">
        <div className="login-brand"><ReadOutlined /> iMaiPlay</div>
        <div>
          <Typography.Title>让每一次学习<br />都成为成长的力量</Typography.Title>
          <Typography.Paragraph>
            汇聚企业精品课程，记录学习轨迹，与团队一起持续进步。
          </Typography.Paragraph>
        </div>
        <div className="login-points">
          <span>体系化课程</span>
          <span>随时随地学习</span>
          <span>成长清晰可见</span>
        </div>
      </section>
      <main className="login-panel">
        <Card className="login-card" bordered={false}>
          <Typography.Title level={2}>欢迎回来</Typography.Title>
          <Typography.Paragraph type="secondary">
            登录企业学习中心，继续你的成长旅程
          </Typography.Paragraph>
          <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark={false}>
            <Form.Item
              label="手机号或邮箱"
              name="identifier"
              rules={[{ required: true, message: '请输入手机号或邮箱' }]}
            >
              <Input size="large" prefix={<MailOutlined />} placeholder="name@company.com 或 13800138000" autoFocus />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password size="large" prefix={<LockOutlined />} placeholder="请输入密码" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" size="large" block>
                登录学习中心
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </main>
    </div>
  );
}
