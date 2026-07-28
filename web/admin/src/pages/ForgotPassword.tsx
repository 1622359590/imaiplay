import { LockOutlined, MobileOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography, message } from 'antd'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { forgotPassword, resetPassword } from '../api/auth'

export default function ForgotPassword() {
  const [sent, setSent] = useState(false); const [loading, setLoading] = useState(false); const navigate = useNavigate()
  const requestCode = async (values: { phone: string }) => { setLoading(true); try { await forgotPassword(values.phone); setSent(true); message.success('验证码已发送，请查看短信或开发日志') } finally { setLoading(false) } }
  const reset = async (values: { phone: string; code: string; new_password: string }) => { setLoading(true); try { await resetPassword(values.phone, values.code, values.new_password); message.success('密码已重置'); navigate('/login') } finally { setLoading(false) } }
  return <div className="login-page"><main className="login-form-wrap"><Card className="login-card" bordered={false}><Typography.Title level={2}>找回密码</Typography.Title><Typography.Paragraph type="secondary">通过手机号验证码重置密码</Typography.Paragraph><Form layout="vertical" size="large" onFinish={sent ? reset : requestCode}><Form.Item label="手机号" name="phone" rules={[{ required: true, message: '请输入手机号' }]}><Input prefix={<MobileOutlined />} /></Form.Item>{sent && <><Form.Item label="验证码" name="code" rules={[{ required: true }]}><Input prefix={<SafetyCertificateOutlined />} /></Form.Item><Form.Item label="新密码" name="new_password" rules={[{ required: true, min: 8, message: '密码至少 8 位' }]}><Input.Password prefix={<LockOutlined />} /></Form.Item></>}<Button type="primary" htmlType="submit" block loading={loading}>{sent ? '重置密码' : '发送验证码'}</Button></Form><Typography.Paragraph style={{ marginTop: 20, textAlign: 'center' }}><Link to="/login">返回登录</Link></Typography.Paragraph></Card></main></div>
}
