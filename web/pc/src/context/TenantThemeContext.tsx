import { createContext, useContext, useEffect, useMemo, type PropsWithChildren } from 'react';
import { ConfigProvider } from 'antd';
import { usePortal } from './PortalContext';
import { applyLearnerPalette } from '../theme/learnerPalette';

const defaultTheme = { primary_color: '#4F46E5', logo_url: '', welcome_text: '' };
const ThemeContext = createContext(defaultTheme);

function validColor(value: string): boolean { return /^#[0-9a-f]{6}$/i.test(value); }

export function TenantThemeProvider({ children }: PropsWithChildren) {
  const { portal } = usePortal();
  const theme = useMemo(() => {
    if (!portal) return defaultTheme;
    return {
      primary_color: validColor(portal.primary_color) ? portal.primary_color : defaultTheme.primary_color,
      logo_url: portal.logo_url || '',
      welcome_text: portal.welcome_text || '',
    };
  }, [portal]);

  useEffect(() => {
    applyLearnerPalette();
    document.documentElement.style.setProperty('--brand-600', theme.primary_color);
    document.documentElement.style.setProperty('--gradient-brand', `linear-gradient(135deg, ${theme.primary_color}, #8B5CF6)`);
  }, [theme]);

  const value = useMemo(() => theme, [theme]);
  return <ThemeContext.Provider value={value}><ConfigProvider theme={{ token: { colorPrimary: theme.primary_color, colorInfo: theme.primary_color, borderRadius: 10 } }}>{children}</ConfigProvider></ThemeContext.Provider>;
}

export function useTenantTheme() { return useContext(ThemeContext); }
