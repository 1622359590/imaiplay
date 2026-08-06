import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react'
import { useLocation } from 'react-router-dom'
import { normalizePrimaryColor } from '@imaiplay/shared/theme/tenantTheme'
import {
  getSessionTenantPortal,
  getTenantPortal,
  type TenantPortal,
} from '../api/theme'
import {
  bindPortalSession,
  getActivePortalCode,
  portalRoutePath,
  PORTAL_SESSION_EXPIRED_EVENT,
  readPortalAccessToken,
  setActivePortalCode,
  setActivePortalIdentity,
} from '../api/authSession'
import {
  resolvePortalLocation,
  shouldRestoreSessionPortal,
  type PortalMode,
} from '../api/portalResolution'
import { applyLearnerPalette, LEARNER_PALETTE } from '../theme/learnerPalette'

const fallback = {
  primary_color: '#4F46E5',
  logo_url: '',
  welcome_text: '',
  name: 'iMaiPlay',
}
export interface TenantThemeContextValue {
  portal?: TenantPortal
  tenantCode?: string
  mode: PortalMode
  loading: boolean
  error?: unknown
  primary_color: string
  logo_url: string
  welcome_text: string
  name: string
  routePath: (childPath: string) => string
  loginPath: string
  forgotPasswordPath: string
}

interface PortalState {
  key: string
  portal?: TenantPortal
  loading: boolean
  error?: unknown
}

const ThemeContext = createContext<TenantThemeContextValue | null>(null)

export function TenantThemeProvider({ children }: PropsWithChildren) {
  const location = useLocation()
  const hostname = window.location.hostname
  const resolution = resolvePortalLocation(location.pathname, hostname)
  const shouldRestoreSession = shouldRestoreSessionPortal(
    resolution,
    readPortalAccessToken(),
  )
  const [state, setState] = useState<PortalState>({
    key: resolution.key,
    loading: resolution.shouldResolve || shouldRestoreSession,
  })

  useEffect(() => {
    let cancelled = false
    const load = () => {
      const restoreSession = shouldRestoreSessionPortal(
        resolution,
        readPortalAccessToken(),
      )
      if (!resolution.shouldResolve && !restoreSession) {
        setActivePortalCode()
        setState({ key: resolution.key, loading: false })
        return
      }

      setActivePortalCode(resolution.tenantCode)
      setState({ key: resolution.key, loading: true })
      const request = restoreSession
        ? getSessionTenantPortal()
        : getTenantPortal(resolution.tenantCode)
      void request
        .then((portal) => {
          if (cancelled) return
          const hadSession = Boolean(readPortalAccessToken())
          if (hadSession && !bindPortalSession(portal)) {
            window.dispatchEvent(new Event(PORTAL_SESSION_EXPIRED_EVENT))
          }
          setActivePortalIdentity(portal)
          setState({ key: resolution.key, portal, loading: false })
        })
        .catch((error: unknown) => {
          if (cancelled) return
          setActivePortalCode()
          setState({ key: resolution.key, loading: false, error })
        })
    }

    load()
    window.addEventListener('tenant-theme-changed', load)
    return () => {
      cancelled = true
      window.removeEventListener('tenant-theme-changed', load)
    }
  }, [
    resolution.key,
    resolution.shouldResolve,
    resolution.tenantCode,
  ])

  const currentState = state.key === resolution.key
    ? state
    : {
      key: resolution.key,
      loading: resolution.shouldResolve || shouldRestoreSession,
    }
  const portal = currentState.portal
  const tenantCode = portal?.code ?? resolution.tenantCode ?? getActivePortalCode()
  const routePath = useCallback(
    (childPath: string) => portalRoutePath(tenantCode, hostname, childPath),
    [hostname, tenantCode],
  )
  const theme = useMemo(() => ({
    primary_color: normalizePrimaryColor(portal?.primary_color, fallback.primary_color),
    logo_url: portal?.logo_url || fallback.logo_url,
    welcome_text: portal?.welcome_text || fallback.welcome_text,
    name: portal?.name || fallback.name,
    browser_title: portal?.browser_title || '',
  }), [portal])

  useEffect(() => {
    applyLearnerPalette()
    document.documentElement.style.setProperty('--brand-600', LEARNER_PALETTE.accent)
    document.documentElement.style.setProperty('--adm-color-primary', LEARNER_PALETTE.accent)
    document.title = portal?.browser_title?.trim() || (portal
      ? `${portal.name} | 企业学习中心`
      : 'iMaiPlay 企业学习中心')
    let favicon = document.querySelector<HTMLLinkElement>('link[data-imaiplay-favicon]')
    if (!favicon) {
      favicon = document.createElement('link')
      favicon.rel = 'icon'
      favicon.dataset.imaiplayFavicon = 'true'
      document.head.appendChild(favicon)
    }
    favicon.href = portal?.logo_url || "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='16' fill='%232563EB'/%3E%3Cpath d='M20 18h24v28H20z' fill='none' stroke='white' stroke-width='4'/%3E%3Cpath d='M26 25h12M26 32h12M26 39h8' stroke='white' stroke-width='3'/%3E%3C/svg%3E"
  }, [portal])

  const value = useMemo<TenantThemeContextValue>(() => ({
    portal,
    tenantCode,
    mode: resolution.mode,
    loading: currentState.loading,
    error: currentState.error,
    ...theme,
    routePath,
    loginPath: routePath('/login'),
    forgotPasswordPath: routePath('/forgot-password'),
  }), [
    currentState.error,
    currentState.loading,
    portal,
    resolution.mode,
    routePath,
    tenantCode,
    theme,
  ])

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTenantTheme() {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTenantTheme must be used within TenantThemeProvider')
  }
  return context
}
