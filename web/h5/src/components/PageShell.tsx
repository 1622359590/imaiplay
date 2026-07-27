import { Outlet } from 'react-router-dom'
import { AppTabBar } from './AppTabBar'
import { useTenantTheme } from '../context/TenantThemeContext'

export function PageShell() {
  const theme = useTenantTheme()
  return (
    <div className="page-shell" data-tenant-logo={theme.logo_url || undefined}>
      <main className="page-content">
        <Outlet />
      </main>
      <AppTabBar />
    </div>
  )
}
