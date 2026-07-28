import { Button, Card, Form, Input, message } from 'antd'
import { useState } from 'react'
import { domainApi } from '../api/domain'
import PageHeader from '../components/PageHeader'
import { tokenRole } from '../api/auth'

export default function DomainSettings() {
  const [loading, setLoading] = useState(false); const [form] = Form.useForm<{ custom_domain: string }>(); const superadmin = tokenRole() === 'superadmin'
  const save = async (values: { custom_domain: string }) => { setLoading(true); try { await domainApi.updateMine(values.custom_domain); message.success('域名绑定已保存') } finally { setLoading(false) } }
  return <><PageHeader title="域名绑定" description={superadmin ? 'superadmin 可通过租户管理接口维护域名；租户管理员可绑定当前租户域名。' : '请输入已 CNAME 到平台的域名。HTTPS 证书由部署层负责。'} /><Card style={{ maxWidth: 640 }}><Form form={form} layout="vertical" onFinish={(values) => void save(values)}><Form.Item name="custom_domain" label="自定义域名" rules={[{ pattern: /^[a-z0-9-]+(\.[a-z0-9-]+)+$/, message: '请输入合法域名' }]}><Input placeholder="academy.example.com" /></Form.Item><Button type="primary" htmlType="submit" loading={loading}>保存域名</Button></Form>{superadmin && <p style={{ marginTop: 16 }}>superadmin 可在租户详情中配置指定租户域名。</p>}</Card></>
}
