import {
  deriveClayColors,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'

const DEFAULT_ACCENT = '#4F46E5'

const SURFACE_PALETTE = {
  heading: '#0F172A',
  text: '#334155',
  muted: '#64748B',
  page: '#F8FAFC',
  card: '#FFFFFF',
  line: '#E2E8F0',
  success: '#10B981',
  successLight: '#ECFDF5',
  warning: '#F59E0B',
  warningLight: '#FFFBEB',
  danger: '#EF4444',
  dangerLight: '#FEF2F2',
  info: '#3B82F6',
  infoLight: '#EFF6FF',
  white: '#FFFFFF',
  navGlass: 'rgba(255, 255, 255, 0.88)',
  overlayDark: 'rgba(15, 23, 42, 0.44)',
  shadowColor: 'rgba(15, 23, 42, 0.06)',
  shadowLightColor: 'rgba(15, 23, 42, 0.04)',
  shadowStrongColor: 'rgba(15, 23, 42, 0.10)',
  shadowSm: '0 1px 2px rgba(15, 23, 42, 0.04)',
  shadow: '0 1px 3px rgba(15, 23, 42, 0.06), 0 4px 12px rgba(15, 23, 42, 0.04)',
  shadowLg: '0 12px 30px rgba(15, 23, 42, 0.08)',
  clayWhiteShadow: '#D1D9E6',
  clayHighlightClear: 'rgba(255, 255, 255, 0)',
  transparent: 'rgba(255, 255, 255, 0)',
} as const

const DEFAULT_CLAY = deriveClayColors(DEFAULT_ACCENT)

export const ADMIN_PALETTE = {
  accent: DEFAULT_ACCENT,
  accentHover: '#4338CA',
  accentLight: '#EEF2FF',
  accentSoft: '#E0E7FF',
  accentStrong: '#3730A3',
  accentForeground: '#4F46E5',
  accentContrastText: '#FFFFFF',
  claySurface: DEFAULT_CLAY.surface,
  clayShadow: DEFAULT_CLAY.shadow,
  clayAtmosphere: DEFAULT_CLAY.atmosphere,
  clayHighlight: DEFAULT_CLAY.highlight,
  ...SURFACE_PALETTE,
} as const

export interface AdminPalette {
  accent: string
  accentHover: string
  accentLight: string
  accentSoft: string
  accentStrong: string
  accentForeground: string
  accentContrastText: string
  heading: string
  text: string
  muted: string
  page: string
  card: string
  line: string
  success: string
  successLight: string
  warning: string
  warningLight: string
  danger: string
  dangerLight: string
  info: string
  infoLight: string
  white: string
  navGlass: string
  overlayDark: string
  shadowColor: string
  shadowLightColor: string
  shadowStrongColor: string
  shadowSm: string
  shadow: string
  shadowLg: string
  claySurface: string
  clayShadow: string
  clayAtmosphere: string
  clayHighlight: string
  clayWhiteShadow: string
  clayHighlightClear: string
  transparent: string
}

function channels(hex: string): [number, number, number] {
  return [
    Number.parseInt(hex.slice(1, 3), 16),
    Number.parseInt(hex.slice(3, 5), 16),
    Number.parseInt(hex.slice(5, 7), 16),
  ]
}

function mix(color: string, target: '#000000' | '#FFFFFF', amount: number): string {
  const source = channels(color)
  const destination = channels(target)
  return `#${source.map((value, index) => (
    Math.round(value + (destination[index] - value) * amount)
      .toString(16)
      .padStart(2, '0')
  )).join('')}`.toUpperCase()
}

function isHexColor(value: string): boolean {
  return /^#[0-9a-f]{6}$/i.test(value)
}

function relativeLuminance(hex: string): number {
  const [red, green, blue] = channels(hex).map((channel) => {
    const value = channel / 255
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue
}

function contrastRatio(first: string, second: string): number {
  const lighter = Math.max(relativeLuminance(first), relativeLuminance(second))
  const darker = Math.min(relativeLuminance(first), relativeLuminance(second))
  return (lighter + 0.05) / (darker + 0.05)
}

function readableForeground(accent: string, surfaces: string[]): string {
  for (let percentage = 0; percentage <= 100; percentage += 1) {
    const candidate = mix(accent, '#000000', percentage / 100)
    if (surfaces.every((surface) => contrastRatio(candidate, surface) >= 4.5)) return candidate
  }
  return '#000000'
}

export function createAdminPalette(primaryColor: string = ADMIN_PALETTE.accent): AdminPalette {
  const accent = isHexColor(primaryColor) ? primaryColor.toUpperCase() : ADMIN_PALETTE.accent
  if (accent === ADMIN_PALETTE.accent) return ADMIN_PALETTE

  const clay = deriveClayColors(accent)
  const accentLight = mix(accent, '#FFFFFF', 0.92)
  const accentSoft = mix(accent, '#FFFFFF', 0.82)
  return {
    ...ADMIN_PALETTE,
    accent,
    accentHover: mix(accent, '#000000', 0.1),
    accentLight,
    accentSoft,
    accentStrong: mix(accent, '#000000', 0.22),
    accentForeground: readableForeground(accent, [ADMIN_PALETTE.card, ADMIN_PALETTE.page, accentLight, accentSoft]),
    accentContrastText: recommendedSelectionColors(accent).selected_text_color,
    claySurface: clay.surface,
    clayShadow: clay.shadow,
    clayAtmosphere: clay.atmosphere,
    clayHighlight: clay.highlight,
  }
}

export function createAdminThemeTokens(
  palette: AdminPalette = ADMIN_PALETTE,
  selectionColors: TenantSelectionColors = recommendedSelectionColors(palette.accent),
) {
  return {
    primary: palette.accent,
    primaryHover: palette.accentHover,
    primaryActive: palette.accentStrong,
    primaryText: palette.accentContrastText,
    link: palette.accentForeground,
    info: palette.info,
    success: palette.success,
    warning: palette.warning,
    danger: palette.danger,
    menuSelectedColor: selectionColors.selected_text_color,
    menuSelectedBackground: selectionColors.selected_background_color,
    menuHoverColor: palette.accentForeground,
    menuHoverBackground: palette.accentSoft,
  }
}

export function applyAdminPalette(
  element: HTMLElement = document.documentElement,
  palette: AdminPalette = ADMIN_PALETTE,
  selectionColors: TenantSelectionColors = recommendedSelectionColors(palette.accent),
) {
  for (const [name, value] of Object.entries(palette)) {
    const property = name.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)
    element.style.setProperty(`--admin-${property}`, value)
  }
  element.style.setProperty('--brand-600', palette.accent)
  element.style.setProperty('--tenant-primary', palette.accent)
  element.style.setProperty('--tenant-selected-background', selectionColors.selected_background_color)
  element.style.setProperty('--tenant-selected-text', selectionColors.selected_text_color)
  element.style.setProperty('--tenant-selected-icon', selectionColors.selected_icon_color)
  element.style.setProperty('--admin-selected-background', selectionColors.selected_background_color)
  element.style.setProperty('--admin-selected-text', selectionColors.selected_text_color)
  element.style.setProperty('--admin-selected-icon', selectionColors.selected_icon_color)
  element.style.setProperty('--tenant-focus', palette.accent)
}
