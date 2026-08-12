import {
  normalizePrimaryColor,
  normalizeSelectionColors,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantThemeContract } from '@imaiplay/shared/types/theme'
import { resolveAdminBrandName } from '../utils/adminBrandName'
import { ADMIN_PALETTE } from './adminPalette'

const DEFAULT_BROWSER_TITLE = 'ImaiPlay 管理后台'

export interface AdminThemeValue {
  logoURL?: string
  brandName: string
  browserTitle: string
  primaryColor: string
  selectedBackgroundColor: string
  selectedTextColor: string
  selectedIconColor: string
}

const fallbackSelection = recommendedSelectionColors(ADMIN_PALETTE.accent)

export const FALLBACK_ADMIN_THEME: AdminThemeValue = {
  logoURL: undefined,
  brandName: 'ImaiPlay',
  browserTitle: DEFAULT_BROWSER_TITLE,
  primaryColor: ADMIN_PALETTE.accent,
  selectedBackgroundColor: fallbackSelection.selected_background_color,
  selectedTextColor: fallbackSelection.selected_text_color,
  selectedIconColor: fallbackSelection.selected_icon_color,
}

export function resolveAdminThemeValue(
  theme: Partial<TenantThemeContract>,
  tenantName?: string | null,
): AdminThemeValue {
  const primaryColor = normalizePrimaryColor(theme.primary_color, ADMIN_PALETTE.accent)
  const selectionColors = normalizeSelectionColors(theme, primaryColor)
  return {
    logoURL: theme.logo_url || undefined,
    brandName: resolveAdminBrandName(theme.welcome_text, theme.brand_name, tenantName),
    browserTitle: theme.browser_title?.trim() || tenantName?.trim() || DEFAULT_BROWSER_TITLE,
    primaryColor,
    selectedBackgroundColor: selectionColors.selected_background_color,
    selectedTextColor: selectionColors.selected_text_color,
    selectedIconColor: selectionColors.selected_icon_color,
  }
}
