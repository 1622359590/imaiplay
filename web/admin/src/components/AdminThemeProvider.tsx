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
          colorPrimaryHover: tokens.primaryHover,
          colorPrimaryActive: tokens.primaryActive,
          colorInfo: tokens.info,
          colorSuccess: tokens.success,
          colorWarning: tokens.warning,
          colorError: tokens.danger,
          colorLink: tokens.link,
          colorLinkHover: tokens.link,
          colorLinkActive: tokens.link,
          colorTextLightSolid: tokens.primaryText,
          colorText: palette.text,
          colorTextHeading: palette.heading,
          colorBgLayout: palette.page,
          colorBgContainer: palette.card,
          colorBgElevated: palette.card,
          colorBorderSecondary: palette.line,
          colorFillSecondary: palette.page,
          borderRadius: 12,
          fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif',
        },
        components: {
          Button: {
            borderRadius: 10,
            primaryShadow: `0 4px 0 ${palette.clayShadow}, 0 9px 18px ${palette.clayAtmosphere}`,
            defaultShadow: `0 3px 0 ${palette.clayWhiteShadow}`,
          },
          Card: { headerBg: palette.card },
          Table: {
            headerBg: palette.page,
            headerColor: palette.heading,
            borderColor: palette.line,
            rowHoverBg: palette.accentLight,
          },
          Layout: { headerBg: palette.card, siderBg: palette.card, bodyBg: palette.page },
          Menu: {
            itemBg: palette.card,
            itemColor: palette.text,
            itemHoverColor: tokens.menuHoverColor,
            itemHoverBg: tokens.menuHoverBackground,
            itemSelectedColor: tokens.menuSelectedColor,
            itemSelectedBg: tokens.menuSelectedBackground,
            groupTitleColor: palette.muted,
            itemBorderRadius: 10,
          },
          Input: {
            activeBorderColor: palette.accent,
            hoverBorderColor: palette.accent,
            activeShadow: `0 0 0 3px ${palette.clayAtmosphere}`,
          },
          Modal: { contentBg: palette.card, headerBg: palette.card },
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
