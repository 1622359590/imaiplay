import type { TenantSelectionColors, TenantThemeContract } from '../types/theme.ts'

export interface ClayColors {
  surface: string
  shadow: string
  atmosphere: string
  highlight: string
}

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

export function deriveClayColors(primaryColor: string): ClayColors {
  const normalized = normalizePrimaryColor(primaryColor, '#4F46E5')
  const { red, green, blue } = hexToRgb(normalized)
  const { hue, saturation, lightness } = rgbToHsl(red, green, blue)
  const surface = hslToRgb(hue, saturation, clamp(lightness, 44, 64))
  const shadow = hslToRgb(
    hue,
    clamp(saturation + 12, 0, 100),
    clamp(lightness - 18, 18, 46),
  )

  return {
    surface: rgbToHex(surface),
    shadow: rgbToHex(shadow),
    atmosphere: `rgba(${surface.red}, ${surface.green}, ${surface.blue}, 0.24)`,
    highlight: 'rgba(255, 255, 255, 0.36)',
  }
}

function isHexColor(value: string): boolean {
  return /^#[0-9a-f]{6}$/i.test(value)
}

function hexToRgb(color: string): RgbColor {
  return {
    red: Number.parseInt(color.slice(1, 3), 16),
    green: Number.parseInt(color.slice(3, 5), 16),
    blue: Number.parseInt(color.slice(5, 7), 16),
  }
}

function rgbToHsl(red: number, green: number, blue: number): HslColor {
  const normalizedRed = red / 255
  const normalizedGreen = green / 255
  const normalizedBlue = blue / 255
  const maximum = Math.max(normalizedRed, normalizedGreen, normalizedBlue)
  const minimum = Math.min(normalizedRed, normalizedGreen, normalizedBlue)
  const chroma = maximum - minimum
  const lightness = (maximum + minimum) / 2
  const saturation = chroma === 0 ? 0 : chroma / (1 - Math.abs(2 * lightness - 1))

  let hue = 0
  if (chroma !== 0) {
    if (maximum === normalizedRed) {
      hue = 60 * (((normalizedGreen - normalizedBlue) / chroma) % 6)
    } else if (maximum === normalizedGreen) {
      hue = 60 * ((normalizedBlue - normalizedRed) / chroma + 2)
    } else {
      hue = 60 * ((normalizedRed - normalizedGreen) / chroma + 4)
    }
  }

  return {
    hue: (hue + 360) % 360,
    saturation: saturation * 100,
    lightness: lightness * 100,
  }
}

function hslToRgb(hue: number, saturation: number, lightness: number): RgbColor {
  const normalizedSaturation = saturation / 100
  const normalizedLightness = lightness / 100
  const chroma = (1 - Math.abs(2 * normalizedLightness - 1)) * normalizedSaturation
  const hueSegment = hue / 60
  const secondary = chroma * (1 - Math.abs((hueSegment % 2) - 1))
  const match = normalizedLightness - chroma / 2

  let channels: [number, number, number]
  if (hueSegment < 1) {
    channels = [chroma, secondary, 0]
  } else if (hueSegment < 2) {
    channels = [secondary, chroma, 0]
  } else if (hueSegment < 3) {
    channels = [0, chroma, secondary]
  } else if (hueSegment < 4) {
    channels = [0, secondary, chroma]
  } else if (hueSegment < 5) {
    channels = [secondary, 0, chroma]
  } else {
    channels = [chroma, 0, secondary]
  }

  return {
    red: Math.round((channels[0] + match) * 255),
    green: Math.round((channels[1] + match) * 255),
    blue: Math.round((channels[2] + match) * 255),
  }
}

function rgbToHex({ red, green, blue }: RgbColor): string {
  return `#${[red, green, blue].map((channel) => channel.toString(16).padStart(2, '0')).join('').toUpperCase()}`
}

function clamp(value: number, minimum: number, maximum: number): number {
  return Math.min(Math.max(value, minimum), maximum)
}

interface RgbColor {
  red: number
  green: number
  blue: number
}

interface HslColor {
  hue: number
  saturation: number
  lightness: number
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
