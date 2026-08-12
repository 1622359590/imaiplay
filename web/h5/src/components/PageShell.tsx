import { Outlet } from 'react-router-dom'

export function PageShell() {
  return (
    <main className="page-content">
      <span className="page-blob page-blob-one" aria-hidden="true" />
      <span className="page-blob page-blob-two" aria-hidden="true" />
      <div className="page-view"><Outlet /></div>
    </main>
  )
}
