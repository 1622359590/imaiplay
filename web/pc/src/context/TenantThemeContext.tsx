import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react';
import { ConfigProvider } from 'antd';
import { getTenantTheme } from '../api/theme';

const defaultTheme = { primary_color: '#4F46E5', logo_url: '', welcome_text: '' };
const ThemeContext = createContext(defaultTheme);

function validColor(value: string): boolean { return /^#[0-9a-f]{6}$/i.test(value); }

export function TenantThemeProvider({ children }: PropsWithChildren) {
  const [theme, setTheme] = useState(defaultTheme);
  useEffect(() => {
    const load = () => {
      void getTenantTheme().then((next) => {
        const resolved = { ...defaultTheme, ...next, primary_color: validColor(next.primary_color) ? next.primary_color : defaultTheme.primary_color };
        setTheme(resolved);
        document.documentElement.style.setProperty('--brand-600', resolved.primary_color);
      }).catch(() => { setTheme(defaultTheme); document.documentElement.style.setProperty('--brand-600', defaultTheme.primary_color); });
    };
    load();
    window.addEventListener('tenant-theme-changed', load);
    return () => window.removeEventListener('tenant-theme-changed', load);
  }, []);
  const value = useMemo(() => theme, [theme]);
  return <ThemeContext.Provider value={value}><ConfigProvider theme={{ token: { colorPrimary: theme.primary_color, colorInfo: theme.primary_color, borderRadius: 10 } }}>{children}</ConfigProvider></ThemeContext.Provider>;
}

export function useTenantTheme() { return useContext(ThemeContext); }
