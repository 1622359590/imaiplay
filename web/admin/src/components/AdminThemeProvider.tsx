import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { useEffect, type PropsWithChildren } from 'react'
import { AdminThemeContextProvider, useAdminTheme } from '../context/AdminThemeContext'
import { ADMIN_PALETTE, applyAdminPalette } from '../theme/adminPalette'

function DynamicTheme({ children }: PropsWithChildren) {
  const theme = useAdminTheme()
  useEffect(() => {
    applyAdminPalette()
    document.documentElement.style.setProperty('--admin-accent', theme.primaryColor)
    document.documentElement.style.setProperty('--brand-600', theme.primaryColor)
    document.documentElement.style.setProperty('--tenant-primary', theme.primaryColor)
    document.documentElement.style.setProperty('--tenant-selected', theme.selectedMenuColor)
    document.documentElement.style.setProperty('--tenant-focus', theme.focusColor)
  }, [theme])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: theme.primaryColor,
          colorInfo: theme.primaryColor,
          colorText: ADMIN_PALETTE.text,
          colorTextHeading: ADMIN_PALETTE.heading,
          colorBgLayout: ADMIN_PALETTE.page,
          colorBgContainer: ADMIN_PALETTE.card,
          colorBorderSecondary: ADMIN_PALETTE.line,
          borderRadius: 10,
          fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif',
        },
        components: {
          Table: { headerBg: ADMIN_PALETTE.page },
          Layout: { headerBg: ADMIN_PALETTE.card, siderBg: ADMIN_PALETTE.card, bodyBg: ADMIN_PALETTE.page },
          Menu: {
            itemBg: ADMIN_PALETTE.card,
            itemColor: ADMIN_PALETTE.text,
            itemHoverColor: theme.selectedMenuColor,
            itemHoverBg: '#fff7f6',
            itemSelectedColor: '#ffffff',
            itemSelectedBg: theme.selectedMenuColor,
            groupTitleColor: ADMIN_PALETTE.muted,
          },
        },
      }}
    >
      {children}
    </ConfigProvider>
  )
}

export default function AdminThemeProvider({ children }: PropsWithChildren) {
  return <AdminThemeContextProvider><DynamicTheme>{children}</DynamicTheme></AdminThemeContextProvider>
}
