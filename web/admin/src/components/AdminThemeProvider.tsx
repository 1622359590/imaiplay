import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { useEffect, type PropsWithChildren } from 'react'
import { AdminThemeContextProvider } from '../context/AdminThemeContext'
import { ADMIN_PALETTE, ADMIN_THEME_TOKENS, applyAdminPalette } from '../theme/adminPalette'

function DynamicTheme({ children }: PropsWithChildren) {
  useEffect(() => {
    applyAdminPalette()
  }, [])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: ADMIN_THEME_TOKENS.primary,
          colorInfo: ADMIN_THEME_TOKENS.info,
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
            itemHoverColor: ADMIN_THEME_TOKENS.menuSelectedColor,
            itemHoverBg: '#fff7f6',
            itemSelectedColor: ADMIN_THEME_TOKENS.menuSelectedColor,
            itemSelectedBg: ADMIN_THEME_TOKENS.menuSelectedBackground,
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
