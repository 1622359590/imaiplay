import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { useEffect, useState, type PropsWithChildren } from 'react'
import { themeApi, type TenantTheme } from '../api/theme'
import { tokenRole } from '../api/auth'

const fallback: TenantTheme = { primary_color: '#4F46E5', logo_url: '', welcome_text: '' }

export default function AdminThemeProvider({ children }: PropsWithChildren) {
  const [theme, setTheme] = useState(fallback)
  useEffect(() => {
    const load = () => {
      if (tokenRole() !== 'tenant_admin') return
      void themeApi.get().then(({ data }) => setTheme({ ...fallback, ...data })).catch(() => undefined)
    }
    load()
    window.addEventListener('tenant-theme-changed', load)
    return () => window.removeEventListener('tenant-theme-changed', load)
  }, [])
  const primary = /^#[0-9a-f]{6}$/i.test(theme.primary_color) ? theme.primary_color : fallback.primary_color
  return <ConfigProvider locale={zhCN} theme={{ token: { colorPrimary: primary, colorInfo: primary, borderRadius: 10, fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif' }, components: { Table: { headerBg: '#F8FAFC' }, Layout: { headerBg: '#fff', siderBg: '#312E81' }, Menu: { darkItemBg: '#312E81', darkItemSelectedBg: primary, darkItemHoverBg: '#4338CA' } } }}>{children}</ConfigProvider>
}
