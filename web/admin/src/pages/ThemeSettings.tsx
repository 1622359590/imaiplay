import { AppstoreOutlined, BookOutlined, CheckOutlined } from '@ant-design/icons'
import {
  normalizePrimaryColor,
  normalizeSelectionColors,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'
import { Alert, Button, Card, ColorPicker, Form, Input, Space, Typography, message } from 'antd'
import { useEffect, useMemo, useState, type CSSProperties } from 'react'
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
import { ADMIN_PALETTE } from '../theme/adminPalette'
import { createThemePreviewStyle } from '../theme/themePreview'

const DEFAULT_PRIMARY = ADMIN_PALETTE.accent

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
  const [primaryColor, setPrimaryColor] = useState<string>(DEFAULT_PRIMARY)
  const [selectionColors, setSelectionColors] = useState<TenantSelectionColors>(
    recommendedSelectionColors(DEFAULT_PRIMARY),
  )
  const [logo, setLogo] = useState<UploadedMedia>()
  const brandName = Form.useWatch('brand_name', form)
  const welcomeText = Form.useWatch('welcome_text', form)
  const previewStyle = useMemo(
    () => createThemePreviewStyle(primaryColor, selectionColors) as CSSProperties,
    [primaryColor, selectionColors],
  )

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

  const reset = () => {
    form.resetFields()
    setPrimaryColor(DEFAULT_PRIMARY)
    setSelectionColors(recommendedSelectionColors(DEFAULT_PRIMARY))
    setLogo(undefined)
  }

  return (
    <div className="admin-page theme-settings-page">
      <PageHeader
        title="主题设置"
        description="统一配置管理后台、PC 与 H5 学员端使用的品牌主题。"
      />
      <Form
        form={form}
        className="theme-settings-form"
        layout="vertical"
        initialValues={{ primary_color: DEFAULT_PRIMARY }}
      >
        <div className="settings-section-grid">
          <Card className="admin-section-card theme-section-brand-basics" title="品牌基础" extra={<Typography.Text type="secondary">全端共享文案</Typography.Text>}>
            <div className="form-grid form-grid-two">
              <Form.Item
                label="品牌名称"
                name="brand_name"
                rules={[{ max: 50, message: '品牌名称不能超过 50 个字符' }]}
              >
                <Input maxLength={50} showCount placeholder="例如：小明科技学习中心" />
              </Form.Item>
              <Form.Item label="浏览器标签页标题" name="browser_title">
                <Input maxLength={255} placeholder="例如：小明科技学习中心" />
              </Form.Item>
            </div>
            <Form.Item label="欢迎语" name="welcome_text">
              <Input.TextArea rows={3} maxLength={255} showCount placeholder="例如：欢迎来到小明科技学习中心" />
            </Form.Item>
          </Card>

          <Card className="admin-section-card theme-section-color-system" title="颜色系统" extra={<Typography.Text type="secondary">保存后即时同步</Typography.Text>}>
            <Form.Item label="品牌主色">
              <ColorPicker
                format="hex"
                value={primaryColor}
                onChange={(color) => {
                  const nextPrimary = normalizePrimaryColor(color.toHexString(), DEFAULT_PRIMARY)
                  setSelectionColors((current) => syncSelectionColorsForPrimaryChange(primaryColor, nextPrimary, current))
                  setPrimaryColor(nextPrimary)
                }}
                showText
              />
            </Form.Item>
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
            {hasLowSelectionContrast(selectionColors) && (
              <Alert className="theme-selection-warning" type="warning" showIcon message="当前配色对比度较低，可能看不清" />
            )}
          </Card>

          <Card className="admin-section-card theme-section-brand-assets" title="品牌资产" extra={<Typography.Text type="secondary">建议使用透明背景图片</Typography.Text>}>
            <Form.Item label="企业 Logo" className="theme-logo-field">
              <MediaUploader
                value={logo}
                accept="image"
                upload={(file, onProgress) => resourceApi.upload(file, onProgress).then((response) => response.data)}
                onChange={setLogo}
              />
            </Form.Item>
          </Card>

          <Card className="admin-section-card theme-section-live-preview" title="实时预览" extra={<Typography.Text type="secondary">当前编辑值</Typography.Text>}>
            <div className="theme-live-preview" style={previewStyle}>
              <section className="theme-preview-admin" aria-label="管理后台预览">
                <div className="theme-preview-admin-brand"><span className="theme-preview-logo">{logo ? <img src={logo.url} alt="" /> : <BookOutlined />}</span><strong>{brandName || 'ImaiPlay'}</strong></div>
                <div className="theme-preview-admin-nav theme-selection-preview-row is-selected">
                  <AppstoreOutlined className="theme-preview-admin-nav-icon" />
                  <span>课程管理</span>
                  <CheckOutlined />
                </div>
                <button type="button" className="theme-preview-primary-button">主要操作</button>
              </section>
              <section className="theme-preview-learner-hero" aria-label="学员端首页预览">
                <span className="theme-preview-eyebrow">继续学习</span>
                <strong>{welcomeText || `欢迎来到${brandName || '学习中心'}`}</strong>
                <p>PC 与 H5 会使用相同的品牌主色与 Clay 深度。</p>
                <div className="theme-preview-progress" aria-label="学习进度 68%"><span /></div>
                <small>课程进度 68%</small>
              </section>
              <div className="theme-preview-clay-contact" aria-label="Clay 接触阴影预览">实体接触阴影</div>
            </div>
          </Card>
        </div>
        <div className="admin-form-actions theme-settings-actions">
          <Space>
            <Button type="primary" loading={loading} onClick={() => void save()}>保存主题</Button>
            <Button onClick={reset}>重置</Button>
          </Space>
          <Typography.Text type="secondary">保存成功后当前后台将立即刷新主题。</Typography.Text>
        </div>
      </Form>
    </div>
  )
}
