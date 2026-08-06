export const ADMIN_PALETTE = {
  accent: '#ff5156',
  accentHover: '#e84349',
  accentSoft: '#fff1f0',
  heading: '#262626',
  text: '#595959',
  muted: '#737373',
  page: '#fafafa',
  card: '#ffffff',
  line: '#eeeeee',
} as const

export interface AdminPalette {
  accent: string
  accentHover: string
  accentSoft: string
  heading: string
  text: string
  muted: string
  page: string
  card: string
  line: string
}

function channels(hex: string): [number, number, number] {
  return [
    Number.parseInt(hex.slice(1, 3), 16),
    Number.parseInt(hex.slice(3, 5), 16),
    Number.parseInt(hex.slice(5, 7), 16),
  ]
}

function mix(color: string, target: '#000000' | '#ffffff', amount: number): string {
  const source = channels(color)
  const destination = channels(target)
  return `#${source.map((value, index) => (
    Math.round(value + (destination[index] - value) * amount)
      .toString(16)
      .padStart(2, '0')
  )).join('')}`.toUpperCase()
}

function selectedTextColor(color: string): '#000000' | '#ffffff' {
  const [red, green, blue] = channels(color)
  const perceivedBrightness = (red * 299 + green * 587 + blue * 114) / 1000
  return perceivedBrightness >= 145 ? '#000000' : '#ffffff'
}

export function createAdminPalette(primaryColor: string = ADMIN_PALETTE.accent): AdminPalette {
  const accent = /^#[0-9a-f]{6}$/i.test(primaryColor) ? primaryColor.toUpperCase() : ADMIN_PALETTE.accent
  if (accent.toLowerCase() === ADMIN_PALETTE.accent) return ADMIN_PALETTE
  return {
    ...ADMIN_PALETTE,
    accent,
    accentHover: mix(accent, '#000000', 0.1),
    accentSoft: mix(accent, '#ffffff', 0.9),
  }
}

export function createAdminThemeTokens(palette: AdminPalette = ADMIN_PALETTE) {
  return {
    primary: palette.accent,
    info: palette.accent,
    menuSelectedColor: selectedTextColor(palette.accent),
    menuSelectedBackground: palette.accent,
    menuHoverColor: palette.accent,
    menuHoverBackground: palette.accentSoft,
  }
}

export function applyAdminPalette(
  element: HTMLElement = document.documentElement,
  palette: AdminPalette = ADMIN_PALETTE,
) {
  for (const [name, value] of Object.entries(palette)) {
    const property = name.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)
    element.style.setProperty(`--admin-${property}`, value)
  }
  element.style.setProperty('--brand-600', palette.accent)
  element.style.setProperty('--tenant-primary', palette.accent)
  element.style.setProperty('--tenant-selected', palette.accent)
  element.style.setProperty('--tenant-selected-text', createAdminThemeTokens(palette).menuSelectedColor)
  element.style.setProperty('--tenant-focus', palette.accent)
}
