import { Button, Card, Form, Input, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import PageHeader from '../components/PageHeader'
import { tenantRegistrationApi, type TenantRegistrationInput } from '../api/tenantRegistration'

export default function CreateTenant() {
  const [form] = Form.useForm<TenantRegistrationInput>(); const navigate = useNavigate()
  const submit = async (values: TenantRegistrationInput) => { await tenantRegistrationApi.create(values); message.success('租户创建成功'); navigate('/tenants') }
  return <><PageHeader title="创建租户" description="为客户创建租户、管理员和演示数据，管理员不会自动登录。" /><Card style={{ maxWidth: 640 }}><Form form={form} layout="vertical" onFinish={(values) => void submit(values)}><Form.Item name="organization_name" label="组织名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="admin_email" label="管理员邮箱" rules={[{ required: true, type: 'email' }]}><Input /></Form.Item><Form.Item name="admin_name" label="管理员姓名" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="phone" label="手机号"><Input /></Form.Item><Form.Item name="password" label="初始密码" rules={[{ required: true, min: 8 }]}><Input.Password /></Form.Item><Form.Item name="plan_id" label="套餐 ID（可选）"><Input placeholder="留空使用默认套餐" /></Form.Item><Button type="primary" htmlType="submit">创建租户</Button></Form></Card></>
}
