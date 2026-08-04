import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { useEffect, type PropsWithChildren } from 'react'
import { ADMIN_PALETTE, applyAdminPalette } from '../theme/adminPalette'

export default function AdminThemeProvider({ children }: PropsWithChildren) {
  useEffect(() => applyAdminPalette(), [])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: ADMIN_PALETTE.accent,
          colorInfo: ADMIN_PALETTE.accent,
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
          Layout: {
            headerBg: ADMIN_PALETTE.card,
            siderBg: ADMIN_PALETTE.card,
            bodyBg: ADMIN_PALETTE.page,
          },
          Menu: {
            itemBg: ADMIN_PALETTE.card,
            itemColor: ADMIN_PALETTE.text,
            itemHoverColor: ADMIN_PALETTE.accent,
            itemHoverBg: '#fff7f6',
            itemSelectedColor: ADMIN_PALETTE.accent,
            itemSelectedBg: ADMIN_PALETTE.accentSoft,
            groupTitleColor: ADMIN_PALETTE.muted,
          },
        },
      }}
    >
      {children}
    </ConfigProvider>
  )
}
