import { Button, Result } from 'antd'
import { useRouteError } from 'react-router-dom'
import { routeErrorPresentation } from './routeErrorModel'

export default function RouteErrorPage() {
  const error = useRouteError()
  const presentation = routeErrorPresentation(error)

  return (
    <main className="route-error-page">
      <Result
        status="warning"
        title={presentation.title}
        subTitle={presentation.description}
        extra={<Button type="primary" onClick={() => window.location.reload()}>刷新页面</Button>}
      />
    </main>
  )
}
