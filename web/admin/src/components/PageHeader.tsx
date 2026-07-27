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
    <div className="page-header">
      <div>
        <h1>{title}</h1>
        {description && <p>{description}</p>}
      </div>
      {extra}
    </div>
  )
}
