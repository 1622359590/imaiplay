import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react'
import { getTenantTheme } from '../api/theme'

const fallback = { primary_color: '#4F46E5', logo_url: '', welcome_text: '' }
const ThemeContext = createContext(fallback)
const validColor = (value: string) => /^#[0-9a-f]{6}$/i.test(value)

export function TenantThemeProvider({ children }: PropsWithChildren) {
  const [theme, setTheme] = useState(fallback)
  useEffect(() => {
    const load = () => {
      void getTenantTheme().then((next) => { const resolved = { ...fallback, ...next, primary_color: validColor(next.primary_color) ? next.primary_color : fallback.primary_color }; setTheme(resolved); applyTheme(resolved.primary_color) }).catch(() => { setTheme(fallback); applyTheme(fallback.primary_color) })
    }
    load(); window.addEventListener('tenant-theme-changed', load)
    return () => window.removeEventListener('tenant-theme-changed', load)
  }, [])
  return <ThemeContext.Provider value={useMemo(() => theme, [theme])}>{children}</ThemeContext.Provider>
}

function applyTheme(color: string) { document.documentElement.style.setProperty('--brand-600', color); document.documentElement.style.setProperty('--adm-color-primary', color); document.documentElement.style.setProperty('--gradient-brand', `linear-gradient(135deg, ${color}, #8B5CF6)`) }
export function useTenantTheme() { return useContext(ThemeContext) }
