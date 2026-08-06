import { createContext, useContext, useEffect, useMemo, type PropsWithChildren } from 'react';
import { ConfigProvider } from 'antd';
import { normalizePrimaryColor } from '@imaiplay/shared/theme/tenantTheme';
import { usePortal } from './PortalContext';
import {
  applyLearnerPalette,
  createLearnerPalette,
  createLearnerThemeTokens,
  LEARNER_PALETTE,
} from '../theme/learnerPalette';

const defaultTheme = { primary_color: LEARNER_PALETTE.accent, logo_url: '', welcome_text: '' };
const ThemeContext = createContext(defaultTheme);

export function TenantThemeProvider({ children }: PropsWithChildren) {
  const { portal } = usePortal();
  const theme = useMemo(() => {
    if (!portal) return defaultTheme;
    return {
      primary_color: normalizePrimaryColor(portal.primary_color, defaultTheme.primary_color),
      logo_url: portal.logo_url || '',
      welcome_text: portal.welcome_text || '',
    };
  }, [portal]);
  const palette = useMemo(() => createLearnerPalette(theme.primary_color), [theme.primary_color]);
  const tokens = useMemo(() => createLearnerThemeTokens(theme.primary_color), [theme.primary_color]);

  useEffect(() => {
    applyLearnerPalette(document.documentElement, palette);
    document.documentElement.style.setProperty('--brand-600', palette.accent);
  }, [palette]);

  const value = useMemo(() => theme, [theme]);
  return (
    <ThemeContext.Provider value={value}>
      <ConfigProvider
        theme={{
          token: tokens,
        }}
      >
        {children}
      </ConfigProvider>
    </ThemeContext.Provider>
  );
}

export function useTenantTheme() { return useContext(ThemeContext); }
