import { Card, Form, Input, Select, Button, Space, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { smsApi, type SMSConfig as SMSConfigValue } from '../api/sms'
import PageHeader from '../components/PageHeader'

export default function SMSConfig() {
  const [form] = Form.useForm(); const [loading, setLoading] = useState(false); const [config, setConfig] = useState<SMSConfigValue>()
  useEffect(() => { smsApi.get().then(({ data }) => { setConfig(data); form.setFieldsValue(data) }) }, [form])
  const save = async () => { setLoading(true); try { const { data } = await smsApi.save(await form.validateFields()); setConfig(data); form.setFieldsValue(data); message.success('短信配置已保存') } finally { setLoading(false) } }
  const test = async () => { const phone = await form.validateFields(['test_phone']); await smsApi.test(phone.test_phone); message.success('测试短信已发送') }
  return <><PageHeader title="短信配置" description="未配置阿里云时自动使用日志通道，验证码会写入服务日志。" /><Card><Typography.Paragraph type="secondary">当前通道：{config?.provider || 'log'}</Typography.Paragraph><Form form={form} layout="vertical"><Form.Item label="Provider" name="provider"><Select options={[{ value: 'log', label: '日志' }, { value: 'aliyun', label: '阿里云' }]} /></Form.Item><Form.Item label="AccessKey ID" name="access_key_id"><Input /></Form.Item><Form.Item label="AccessKey Secret（仅写入，不回显）" name="access_key_secret"><Input.Password /></Form.Item><Form.Item label="短信签名" name="sign_name"><Input /></Form.Item><Form.Item label="模板 ID" name="template_code"><Input /></Form.Item><Space><Button type="primary" onClick={save} loading={loading}>保存</Button><Form.Item name="test_phone" noStyle><Input placeholder="测试手机号" /></Form.Item><Button onClick={test}>测试发送</Button></Space></Form></Card></>
}
