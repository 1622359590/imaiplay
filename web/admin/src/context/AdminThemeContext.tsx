import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react'
import { useSelector } from 'react-redux'
import { ADMIN_TENANT_NAME_KEY } from '../api/authSession'
import { themeApi } from '../api/theme'
import type { RootState } from '../store'
import {
  FALLBACK_ADMIN_THEME,
  resolveAdminThemeValue,
  type AdminThemeValue,
} from '../theme/adminThemeValue'

const DEFAULT_BROWSER_TITLE = 'ImaiPlay 管理后台'
const DEFAULT_FAVICON = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='16' fill='%234F46E5'/%3E%3Cpath d='M26 19l22 13-22 13V19z' fill='white'/%3E%3C/svg%3E"

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

const AdminThemeContext = createContext<AdminThemeValue>(FALLBACK_ADMIN_THEME)

export function AdminThemeContextProvider({ children }: PropsWithChildren) {
  const profile = useSelector((state: RootState) => state.user.profile)
  const [value, setValue] = useState(FALLBACK_ADMIN_THEME)

  useEffect(() => {
    applyBrowserBranding(FALLBACK_ADMIN_THEME.browserTitle)
    if (!profile || profile.role === 'superadmin') {
      setValue(FALLBACK_ADMIN_THEME)
      return
    }
    let active = true
    const load = () => {
      void themeApi.get().then(({ data }) => {
        if (!active) return
        const resolved = resolveAdminThemeValue(data, localStorage.getItem(ADMIN_TENANT_NAME_KEY))
        setValue(resolved)
        applyBrowserBranding(
          resolved.browserTitle,
          resolved.logoURL,
        )
      }).catch(() => active && setValue(FALLBACK_ADMIN_THEME))
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
