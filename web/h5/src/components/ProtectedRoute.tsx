import { useEffect } from 'react'
import { DotLoading, ErrorBlock } from 'antd-mobile'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { isAuthenticated } from '../api/auth'
import { portalErrorContent } from '../api/portalResolution'
import { useTenantTheme } from '../context/TenantThemeContext'

export function ProtectedRoute() {
  const location = useLocation()
  const theme = useTenantTheme()

  if (theme.loading) {
    return <div className="loading-state"><DotLoading color="primary" /> 正在加载企业门户</div>
  }
  if (theme.mode === 'platform') {
    if (theme.error) {
      const content = portalErrorContent(theme.error)
      return (
        <ErrorBlock
          status="disconnected"
          title={content.title}
          description={content.description}
        />
      )
    }
    return theme.portal && isAuthenticated(theme.portal)
      ? <Navigate to={theme.routePath('/')} replace />
      : <PlatformLoginRedirect />
  }
  if (theme.error || !theme.portal) {
    const content = portalErrorContent(theme.error)
    return (
      <ErrorBlock
        status="disconnected"
        title={content.title}
        description={content.description}
      />
    )
  }

  return isAuthenticated(theme.portal) ? (
    <Outlet />
  ) : (
    <Navigate to={theme.loginPath} replace state={{ from: location.pathname }} />
  )
}

function PlatformLoginRedirect() {
  useEffect(() => {
    window.location.replace('/login')
  }, [])
  return <div className="loading-state"><DotLoading color="primary" /> 正在前往统一登录</div>
}
