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

export const ADMIN_THEME_TOKENS = {
  primary: ADMIN_PALETTE.accent,
  info: ADMIN_PALETTE.accent,
  menuSelectedColor: ADMIN_PALETTE.accent,
  menuSelectedBackground: ADMIN_PALETTE.accentSoft,
} as const

export function applyAdminPalette(element: HTMLElement = document.documentElement) {
  for (const [name, value] of Object.entries(ADMIN_PALETTE)) {
    const property = name.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)
    element.style.setProperty(`--admin-${property}`, value)
  }
  element.style.setProperty('--brand-600', ADMIN_THEME_TOKENS.primary)
  element.style.setProperty('--tenant-primary', ADMIN_THEME_TOKENS.primary)
  element.style.setProperty('--tenant-selected', ADMIN_THEME_TOKENS.menuSelectedBackground)
  element.style.setProperty('--tenant-focus', ADMIN_THEME_TOKENS.primary)
}
