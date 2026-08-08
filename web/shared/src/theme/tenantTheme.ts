import type { TenantSelectionColors, TenantThemeContract } from '../types/theme.ts'

export function normalizePrimaryColor(value: unknown, fallback: string): string {
  const color = typeof value === 'string' ? value.trim() : ''
  return isHexColor(color) ? color.toUpperCase() : fallback
}

export function contrastRatio(foreground: string, background: string): number {
  const foregroundLuminance = relativeLuminance(foreground)
  const backgroundLuminance = relativeLuminance(background)
  const lighter = Math.max(foregroundLuminance, backgroundLuminance)
  const darker = Math.min(foregroundLuminance, backgroundLuminance)
  return (lighter + 0.05) / (darker + 0.05)
}

export function recommendedSelectionColors(primaryColor: string): TenantSelectionColors {
  const background = normalizePrimaryColor(primaryColor, '#4F46E5')
  const text = contrastRatio('#000000', background) >= contrastRatio('#FFFFFF', background)
    ? '#000000'
    : '#FFFFFF'
  return {
    selected_background_color: background,
    selected_text_color: text,
    selected_icon_color: text,
  }
}

export function normalizeSelectionColors(
  theme: Partial<TenantThemeContract>,
  fallbackPrimary: string,
): TenantSelectionColors {
  const primary = normalizePrimaryColor(theme.primary_color, fallbackPrimary)
  const recommended = recommendedSelectionColors(primary)
  const background = normalizePrimaryColor(
    theme.selected_background_color,
    recommended.selected_background_color,
  )
  const recommendedForBackground = recommendedSelectionColors(background)
  const text = normalizePrimaryColor(
    theme.selected_text_color,
    recommendedForBackground.selected_text_color,
  )
  return {
    selected_background_color: background,
    selected_text_color: text,
    selected_icon_color: normalizePrimaryColor(theme.selected_icon_color, text),
  }
}

function isHexColor(value: string): boolean {
  return /^#[0-9a-f]{6}$/i.test(value)
}

function relativeLuminance(color: string): number {
  const normalized = normalizePrimaryColor(color, '#000000')
  const channels = [1, 3, 5].map((offset) => Number.parseInt(normalized.slice(offset, offset + 2), 16) / 255)
  const [red, green, blue] = channels.map((channel) => (
    channel <= 0.04045
      ? channel / 12.92
      : ((channel + 0.055) / 1.055) ** 2.4
  ))
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue
}
