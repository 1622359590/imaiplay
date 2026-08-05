import { Button, Result } from 'antd'
import { useRouteError } from 'react-router-dom'

function errorDetail(error: unknown): string | null {
  if (error instanceof Error) return error.message
  return null
}

export default function RouteErrorPage() {
  const error = useRouteError()
  const detail = errorDetail(error)

  return (
    <main className="route-error-page">
      <Result
        status="warning"
        title="页面资源加载失败"
        subTitle="系统可能刚刚完成更新，或当前网络暂时不可用。请刷新页面后重试。"
        extra={<Button type="primary" onClick={() => window.location.reload()}>刷新页面</Button>}
      />
      {detail && <details className="route-error-detail"><summary>错误详情</summary><code>{detail}</code></details>}
    </main>
  )
}
