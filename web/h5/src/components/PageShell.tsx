import { Outlet } from 'react-router-dom'

export function PageShell() {
  return <main className="page-content"><Outlet /></main>
}
