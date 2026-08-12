import type { ReactNode } from 'react'

export default function PageHeader({
  title,
  description,
  extra,
}: {
  title: string
  description?: string
  extra?: ReactNode
}) {
  return (
    <header className="page-header">
      <div className="page-header-copy">
        <h1>{title}</h1>
        {description && <p>{description}</p>}
      </div>
      {extra && <div className="page-header-actions">{extra}</div>}
    </header>
  )
}
