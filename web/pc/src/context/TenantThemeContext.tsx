import { createContext, useContext, useEffect, useMemo, type PropsWithChildren } from 'react';
import { ConfigProvider } from 'antd';
import { usePortal } from './PortalContext';
import { applyLearnerPalette, LEARNER_PALETTE } from '../theme/learnerPalette';

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
    document.documentElement.style.setProperty('--brand-600', LEARNER_PALETTE.accent);
  }, []);

  const value = useMemo(() => theme, [theme]);
  return (
    <ThemeContext.Provider value={value}>
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: LEARNER_PALETTE.accent,
            colorInfo: LEARNER_PALETTE.accent,
            colorText: LEARNER_PALETTE.text,
            colorTextHeading: LEARNER_PALETTE.heading,
            colorBgLayout: LEARNER_PALETTE.page,
            colorBgContainer: LEARNER_PALETTE.card,
            colorBorderSecondary: LEARNER_PALETTE.line,
            borderRadius: 10,
          },
        }}
      >
        {children}
      </ConfigProvider>
    </ThemeContext.Provider>
  );
}

export function useTenantTheme() { return useContext(ThemeContext); }
