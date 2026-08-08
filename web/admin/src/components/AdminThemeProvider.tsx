import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { useEffect, useMemo, type PropsWithChildren } from 'react'
import { AdminThemeContextProvider, useAdminTheme } from '../context/AdminThemeContext'
import { applyAdminPalette, createAdminPalette, createAdminThemeTokens } from '../theme/adminPalette'

function DynamicTheme({ children }: PropsWithChildren) {
  const theme = useAdminTheme()
  const palette = useMemo(() => createAdminPalette(theme.primaryColor), [theme.primaryColor])
  const selectionColors = useMemo(() => ({
    selected_background_color: theme.selectedBackgroundColor,
    selected_text_color: theme.selectedTextColor,
    selected_icon_color: theme.selectedIconColor,
  }), [theme.selectedBackgroundColor, theme.selectedIconColor, theme.selectedTextColor])
  const tokens = useMemo(
    () => createAdminThemeTokens(palette, selectionColors),
    [palette, selectionColors],
  )

  useEffect(() => {
    applyAdminPalette(document.documentElement, palette, selectionColors)
  }, [palette, selectionColors])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: tokens.primary,
          colorInfo: tokens.info,
          colorText: palette.text,
          colorTextHeading: palette.heading,
          colorBgLayout: palette.page,
          colorBgContainer: palette.card,
          colorBorderSecondary: palette.line,
          borderRadius: 10,
          fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif',
        },
        components: {
          Table: { headerBg: palette.page },
          Layout: { headerBg: palette.card, siderBg: palette.card, bodyBg: palette.page },
          Menu: {
            itemBg: palette.card,
            itemColor: palette.text,
            itemHoverColor: tokens.menuHoverColor,
            itemHoverBg: tokens.menuHoverBackground,
            itemSelectedColor: tokens.menuSelectedColor,
            itemSelectedBg: tokens.menuSelectedBackground,
            groupTitleColor: palette.muted,
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
