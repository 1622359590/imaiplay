import { Button, Card, ColorPicker, Form, Input, Space, Upload, message } from 'antd'
import { useEffect, useState } from 'react'
import PageHeader from '../components/PageHeader'
import { themeApi, type TenantTheme } from '../api/theme'
import { resourceApi } from '../api/resource'

export default function ThemeSettings() {
  const [form] = Form.useForm<TenantTheme>()
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [primaryColor, setPrimaryColor] = useState('#4F46E5')
  useEffect(() => { void themeApi.get().then(({ data }) => { form.setFieldsValue(data); setPrimaryColor(/^#[0-9a-fA-F]{6}$/.test(data.primary_color) ? data.primary_color : '#4F46E5') }) }, [form])
  const save = async () => { const values = await form.validateFields(); setLoading(true); try { await themeApi.update({ ...values, primary_color: primaryColor }); window.dispatchEvent(new Event('tenant-theme-changed')); message.success('主题设置已保存') } finally { setLoading(false) } }
  return <>
    <PageHeader title="主题设置" description="只调整品牌色、Logo 和欢迎语，页面结构保持不变。" />
    <Card style={{ maxWidth: 680 }}>
      <Form form={form} layout="vertical" initialValues={{ primary_color: '#4F46E5' }}>
        <Form.Item label="品牌主色"><ColorPicker format="hex" value={primaryColor} onChange={(_, hex) => setPrimaryColor(hex)} showText /></Form.Item>
        <Form.Item label="Logo 地址" name="logo_url" extra="可填写外部图片地址，或上传后使用现有资源地址。"><Input placeholder="https://..." /></Form.Item>
        <Form.Item label="上传 Logo"><Upload showUploadList={false} maxCount={1} beforeUpload={async (file) => { setUploading(true); try { const { data } = await resourceApi.upload(file); form.setFieldValue('logo_url', data.url); message.success('Logo 已上传') } finally { setUploading(false) } return false }}><Button loading={uploading}>上传图片</Button></Upload></Form.Item>
        <Form.Item label="欢迎语" name="welcome_text"><Input.TextArea rows={3} maxLength={255} showCount placeholder="例如：欢迎来到小明科技学习中心" /></Form.Item>
        <Space><Button type="primary" loading={loading} onClick={() => void save()}>保存主题</Button><Button onClick={() => { form.resetFields(); setPrimaryColor('#4F46E5') }}>重置</Button></Space>
      </Form>
    </Card>
  </>
}
