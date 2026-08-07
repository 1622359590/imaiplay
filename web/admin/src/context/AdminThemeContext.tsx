import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react'
import { useSelector } from 'react-redux'
import { ADMIN_TENANT_NAME_KEY } from '../api/authSession'
import { themeApi } from '../api/theme'
import type { RootState } from '../store'
import { resolveAdminBrandName } from '../utils/adminBrandName'
import { normalizePrimaryColor } from '@imaiplay/shared/theme/tenantTheme'

const FALLBACK_PRIMARY = '#ff5156'
const DEFAULT_BROWSER_TITLE = 'ImaiPlay 管理后台'
const DEFAULT_FAVICON = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='16' fill='%23FF5A5F'/%3E%3Cpath d='M26 19l22 13-22 13V19z' fill='white'/%3E%3C/svg%3E"

interface AdminThemeValue {
  logoURL?: string
  brandName: string
  browserTitle: string
  primaryColor: string
}

const fallbackTheme: AdminThemeValue = {
  brandName: 'ImaiPlay',
  browserTitle: DEFAULT_BROWSER_TITLE,
  primaryColor: FALLBACK_PRIMARY,
}

function applyBrowserBranding(title: string, logoURL?: string) {
  document.title = title || DEFAULT_BROWSER_TITLE
  let favicon = document.querySelector<HTMLLinkElement>('link[data-imaiplay-favicon]')
  if (!favicon) {
    favicon = document.createElement('link')
    favicon.rel = 'icon'
    favicon.dataset.imaiplayFavicon = 'true'
    document.head.appendChild(favicon)
  }
  favicon.href = logoURL || DEFAULT_FAVICON
}

const AdminThemeContext = createContext<AdminThemeValue>(fallbackTheme)

export function AdminThemeContextProvider({ children }: PropsWithChildren) {
  const profile = useSelector((state: RootState) => state.user.profile)
  const [value, setValue] = useState(fallbackTheme)

  useEffect(() => {
    applyBrowserBranding(fallbackTheme.browserTitle)
    if (!profile || profile.role === 'superadmin') {
      setValue(fallbackTheme)
      return
    }
    let active = true
    const load = () => {
      void themeApi.get().then(({ data }) => {
        if (!active) return
        setValue({
          logoURL: data.logo_url || undefined,
          brandName: resolveAdminBrandName(
            data.brand_name,
            localStorage.getItem(ADMIN_TENANT_NAME_KEY),
          ),
          browserTitle: data.browser_title?.trim() || localStorage.getItem(ADMIN_TENANT_NAME_KEY) || 'ImaiPlay 管理后台',
          primaryColor: normalizePrimaryColor(data.primary_color, FALLBACK_PRIMARY),
        })
        applyBrowserBranding(
          data.browser_title?.trim() || localStorage.getItem(ADMIN_TENANT_NAME_KEY) || 'ImaiPlay 管理后台',
          data.logo_url || undefined,
        )
      }).catch(() => active && setValue(fallbackTheme))
    }
    load()
    window.addEventListener('tenant-theme-changed', load)
    return () => {
      active = false
      window.removeEventListener('tenant-theme-changed', load)
    }
  }, [profile?.tenant_id, profile?.role])

  const stableValue = useMemo(() => value, [value])
  return <AdminThemeContext.Provider value={stableValue}>{children}</AdminThemeContext.Provider>
}

export function useAdminTheme() {
  return useContext(AdminThemeContext)
}
