import {
  deriveClayColors,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'

export type AdminThemePreviewStyle = Record<`--admin-preview-${string}`, string>

export function createThemePreviewStyle(
  primaryColor: string,
  selectionColors: TenantSelectionColors,
): AdminThemePreviewStyle {
  const clay = deriveClayColors(primaryColor)
  const primaryText = recommendedSelectionColors(primaryColor).selected_text_color
  return {
    '--admin-preview-primary': primaryColor,
    '--admin-preview-primary-text': primaryText,
    '--admin-preview-selected-background': selectionColors.selected_background_color,
    '--admin-preview-selected-text': selectionColors.selected_text_color,
    '--admin-preview-selected-icon': selectionColors.selected_icon_color,
    '--admin-preview-clay-surface': clay.surface,
    '--admin-preview-clay-shadow': clay.shadow,
    '--admin-preview-clay-atmosphere': clay.atmosphere,
    '--admin-preview-clay-highlight': clay.highlight,
    '--admin-preview-button-surface': clay.surface,
    '--admin-preview-hero-surface': clay.surface,
    '--admin-preview-progress-surface': clay.surface,
    '--admin-preview-contact-shadow': clay.shadow,
  }
}
