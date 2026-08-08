import { createContext, useContext, useEffect, useMemo, type PropsWithChildren } from 'react';
import { ConfigProvider } from 'antd';
import {
  normalizePrimaryColor,
  normalizeSelectionColors,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme';
import { usePortal } from './PortalContext';
import {
  applyLearnerPalette,
  createLearnerPalette,
  createLearnerThemeTokens,
  LEARNER_PALETTE,
} from '../theme/learnerPalette';

const defaultSelection = recommendedSelectionColors(LEARNER_PALETTE.accent);
const defaultTheme = {
  primary_color: LEARNER_PALETTE.accent,
  ...defaultSelection,
  logo_url: '',
  welcome_text: '',
};
const ThemeContext = createContext(defaultTheme);

export function TenantThemeProvider({ children }: PropsWithChildren) {
  const { portal } = usePortal();
  const theme = useMemo(() => {
    if (!portal) return defaultTheme;
    const primaryColor = normalizePrimaryColor(portal.primary_color, defaultTheme.primary_color);
    return {
      primary_color: primaryColor,
      ...normalizeSelectionColors(portal, primaryColor),
      logo_url: portal.logo_url || '',
      welcome_text: portal.welcome_text || '',
    };
  }, [portal]);
  const palette = useMemo(() => createLearnerPalette(theme.primary_color), [theme.primary_color]);
  const tokens = useMemo(() => createLearnerThemeTokens(theme.primary_color), [theme.primary_color]);
  const selectionColors = useMemo(() => ({
    selected_background_color: theme.selected_background_color,
    selected_text_color: theme.selected_text_color,
    selected_icon_color: theme.selected_icon_color,
  }), [theme.selected_background_color, theme.selected_icon_color, theme.selected_text_color]);

  useEffect(() => {
    applyLearnerPalette(document.documentElement, palette, selectionColors);
    document.documentElement.style.setProperty('--brand-600', palette.accent);
  }, [palette, selectionColors]);

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
