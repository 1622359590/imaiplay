import { Outlet } from 'react-router-dom'
import { AppTabBar } from './AppTabBar'

export function PageShell() {
  return (
    <div className="page-shell">
      <main className="page-content">
        <Outlet />
      </main>
      <AppTabBar />
    </div>
  )
}
