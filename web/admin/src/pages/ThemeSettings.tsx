import { AppstoreOutlined } from '@ant-design/icons'
import {
  normalizePrimaryColor,
  normalizeSelectionColors,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'
import { Alert, Button, Card, ColorPicker, Form, Input, Space, message } from 'antd'
import { useEffect, useState } from 'react'
import { resourceApi } from '../api/resource'
import { themeApi, type TenantTheme } from '../api/theme'
import MediaUploader, {
  type UploadedMedia,
} from '../components/MediaUploader'
import PageHeader from '../components/PageHeader'
import {
  hasLowSelectionContrast,
  syncSelectionColorsForPrimaryChange,
} from '../theme/selectionSettings'

const DEFAULT_PRIMARY = '#4F46E5'

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
  const [primaryColor, setPrimaryColor] = useState(DEFAULT_PRIMARY)
  const [selectionColors, setSelectionColors] = useState<TenantSelectionColors>(
    recommendedSelectionColors(DEFAULT_PRIMARY),
  )
  const [logo, setLogo] = useState<UploadedMedia>()

  useEffect(() => {
    void themeApi.get().then(({ data }) => {
      form.setFieldsValue(data)
      const resolvedPrimary = normalizePrimaryColor(data.primary_color, DEFAULT_PRIMARY)
      setPrimaryColor(resolvedPrimary)
      setSelectionColors(normalizeSelectionColors(data, resolvedPrimary))
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
        ...selectionColors,
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
        description="调整企业品牌色、选中效果、Logo 和欢迎语。"
      />
      <Card style={{ maxWidth: 720 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ primary_color: DEFAULT_PRIMARY }}
        >
          <Form.Item label="品牌主色">
            <ColorPicker
              format="hex"
              value={primaryColor}
              onChange={(color) => {
                const nextPrimary = normalizePrimaryColor(color.toHexString(), DEFAULT_PRIMARY)
                setSelectionColors((current) => syncSelectionColorsForPrimaryChange(
                  primaryColor,
                  nextPrimary,
                  current,
                ))
                setPrimaryColor(nextPrimary)
              }}
              showText
            />
          </Form.Item>
          <div className="theme-selection-settings">
            <div className="theme-selection-pickers">
              <Form.Item label="选中背景色">
                <ColorPicker
                  format="hex"
                  value={selectionColors.selected_background_color}
                  onChange={(color) => setSelectionColors((current) => ({
                    ...current,
                    selected_background_color: color.toHexString().toUpperCase(),
                  }))}
                  showText
                />
              </Form.Item>
              <Form.Item label="选中文字色">
                <ColorPicker
                  format="hex"
                  value={selectionColors.selected_text_color}
                  onChange={(color) => setSelectionColors((current) => ({
                    ...current,
                    selected_text_color: color.toHexString().toUpperCase(),
                  }))}
                  showText
                />
              </Form.Item>
              <Form.Item label="选中图标色">
                <ColorPicker
                  format="hex"
                  value={selectionColors.selected_icon_color}
                  onChange={(color) => setSelectionColors((current) => ({
                    ...current,
                    selected_icon_color: color.toHexString().toUpperCase(),
                  }))}
                  showText
                />
              </Form.Item>
            </div>
            <div className="theme-selection-preview" aria-label="选中效果预览">
              <span className="theme-selection-preview-title">选中效果预览</span>
              <div className="theme-selection-preview-row">
                <AppstoreOutlined />
                <span>普通菜单</span>
              </div>
              <div
                className="theme-selection-preview-row is-selected"
                style={{
                  color: selectionColors.selected_text_color,
                  background: selectionColors.selected_background_color,
                }}
              >
                <AppstoreOutlined style={{ color: selectionColors.selected_icon_color }} />
                <span>选中菜单</span>
              </div>
            </div>
          </div>
          {hasLowSelectionContrast(selectionColors) && (
            <Alert
              className="theme-selection-warning"
              type="warning"
              showIcon
              message="当前配色对比度较低，可能看不清"
            />
          )}
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
                setPrimaryColor(DEFAULT_PRIMARY)
                setSelectionColors(recommendedSelectionColors(DEFAULT_PRIMARY))
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
