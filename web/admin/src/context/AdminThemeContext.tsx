import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react'
import { useSelector } from 'react-redux'
import { ADMIN_TENANT_NAME_KEY } from '../api/authSession'
import { themeApi } from '../api/theme'
import type { RootState } from '../store'

const FALLBACK_PRIMARY = '#ff4e4f'
const DEFAULT_BROWSER_TITLE = 'ImaiPlay 管理后台'
const DEFAULT_FAVICON = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='16' fill='%23FF5A5F'/%3E%3Cpath d='M26 19l22 13-22 13V19z' fill='white'/%3E%3C/svg%3E"

function rgb(hex: string): [number, number, number] | undefined {
  if (!/^#[0-9a-f]{6}$/i.test(hex)) return undefined
  return [1, 3, 5].map((index) => Number.parseInt(hex.slice(index, index + 2), 16)) as [number, number, number]
}

function luminance(hex: string) {
  const channels = rgb(hex)
  if (!channels) return 0
  const [red, green, blue] = channels.map((channel) => {
    const value = channel / 255
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
  })
  return red * 0.2126 + green * 0.7152 + blue * 0.0722
}

function whiteContrast(hex: string) {
  return 1.05 / (luminance(hex) + 0.05)
}

export function contrastSafeColor(value: string) {
  let channels = rgb(value) || rgb(FALLBACK_PRIMARY)!
  for (let attempt = 0; attempt < 24; attempt += 1) {
    const color = `#${channels.map((channel) => Math.round(channel).toString(16).padStart(2, '0')).join('')}`
    if (whiteContrast(color) >= 4.5) return color
    channels = channels.map((channel) => channel * 0.9) as [number, number, number]
  }
  return '#8f2528'
}

interface AdminThemeValue {
  logoURL?: string
  brandName: string
  browserTitle: string
  primaryColor: string
  selectedMenuColor: string
  focusColor: string
}

const fallbackTheme: AdminThemeValue = {
  brandName: 'ImaiPlay',
  browserTitle: DEFAULT_BROWSER_TITLE,
  primaryColor: FALLBACK_PRIMARY,
  selectedMenuColor: contrastSafeColor(FALLBACK_PRIMARY),
  focusColor: contrastSafeColor(FALLBACK_PRIMARY),
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
        const primaryColor = /^#[0-9a-f]{6}$/i.test(data.primary_color) ? data.primary_color : FALLBACK_PRIMARY
        const selectedMenuColor = contrastSafeColor(primaryColor)
        setValue({
          logoURL: data.logo_url || undefined,
          brandName: localStorage.getItem(ADMIN_TENANT_NAME_KEY) || 'ImaiPlay',
          browserTitle: data.browser_title?.trim() || localStorage.getItem(ADMIN_TENANT_NAME_KEY) || 'ImaiPlay 管理后台',
          primaryColor,
          selectedMenuColor,
          focusColor: selectedMenuColor,
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
