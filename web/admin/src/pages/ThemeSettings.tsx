import { Button, Card, ColorPicker, Form, Input, Space, message } from 'antd'
import { useEffect, useState } from 'react'
import { resourceApi } from '../api/resource'
import { themeApi, type TenantTheme } from '../api/theme'
import MediaUploader, {
  type UploadedMedia,
} from '../components/MediaUploader'
import PageHeader from '../components/PageHeader'

function currentLogo(theme: TenantTheme): UploadedMedia | undefined {
  if (!theme.logo_url) return undefined
  return {
    id: 'current-tenant-logo',
    name: '当前企业 Logo',
    resource_type: 'image',
    url: theme.logo_url,
    size_bytes: 0,
  }
}

export default function ThemeSettings() {
  const [form] = Form.useForm<TenantTheme>()
  const [loading, setLoading] = useState(false)
  const [primaryColor, setPrimaryColor] = useState('#4F46E5')
  const [logo, setLogo] = useState<UploadedMedia>()

  useEffect(() => {
    void themeApi.get().then(({ data }) => {
      form.setFieldsValue(data)
      setPrimaryColor(
        /^#[0-9a-fA-F]{6}$/.test(data.primary_color)
          ? data.primary_color
          : '#4F46E5',
      )
      setLogo(currentLogo(data))
    })
  }, [form])

  const save = async () => {
    const values = await form.validateFields()
    setLoading(true)
    try {
      await themeApi.update({
        ...values,
        primary_color: primaryColor,
        logo_url: logo?.url || '',
      })
      window.dispatchEvent(new Event('tenant-theme-changed'))
      message.success('主题设置已保存')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <PageHeader
        title="主题设置"
        description="调整企业品牌色、Logo 和欢迎语。"
      />
      <Card style={{ maxWidth: 720 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ primary_color: '#4F46E5' }}
        >
          <Form.Item label="品牌主色">
            <ColorPicker
              format="hex"
              value={primaryColor}
              onChange={(color) => setPrimaryColor(color.toHexString())}
              showText
            />
          </Form.Item>
          <Form.Item label="企业 Logo">
            <MediaUploader
              value={logo}
              accept="image"
              upload={(file, onProgress) =>
                resourceApi.upload(file, onProgress)
                  .then((response) => response.data)}
              onChange={setLogo}
            />
          </Form.Item>
          <Form.Item label="欢迎语" name="welcome_text">
            <Input.TextArea
              rows={3}
              maxLength={255}
              showCount
              placeholder="例如：欢迎来到小明科技学习中心"
            />
          </Form.Item>
          <Form.Item label="浏览器标签页标题" name="browser_title">
            <Input maxLength={255} placeholder="例如：小明科技学习中心" />
          </Form.Item>
          <Space>
            <Button
              type="primary"
              loading={loading}
              onClick={() => void save()}
            >
              保存主题
            </Button>
            <Button
              onClick={() => {
                form.resetFields()
                setPrimaryColor('#4F46E5')
                setLogo(undefined)
              }}
            >
              重置
            </Button>
          </Space>
        </Form>
      </Card>
    </>
  )
}
