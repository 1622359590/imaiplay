import {
  contrastRatio,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'

export function syncSelectionColorsForPrimaryChange(
  previousPrimary: string,
  nextPrimary: string,
  colors: TenantSelectionColors,
): TenantSelectionColors {
  const previousRecommended = recommendedSelectionColors(previousPrimary)
  const nextRecommended = recommendedSelectionColors(nextPrimary)
  const matches = (left: string, right: string) => left.toUpperCase() === right.toUpperCase()
  return {
    selected_background_color: matches(
      colors.selected_background_color,
      previousRecommended.selected_background_color,
    ) ? nextRecommended.selected_background_color : colors.selected_background_color,
    selected_text_color: matches(
      colors.selected_text_color,
      previousRecommended.selected_text_color,
    ) ? nextRecommended.selected_text_color : colors.selected_text_color,
    selected_icon_color: matches(
      colors.selected_icon_color,
      previousRecommended.selected_icon_color,
    ) ? nextRecommended.selected_icon_color : colors.selected_icon_color,
  }
}

export function hasLowSelectionContrast(colors: TenantSelectionColors): boolean {
  return contrastRatio(colors.selected_text_color, colors.selected_background_color) < 3 ||
    contrastRatio(colors.selected_icon_color, colors.selected_background_color) < 3
}
