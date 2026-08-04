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

export function applyAdminPalette(element: HTMLElement = document.documentElement) {
  for (const [name, value] of Object.entries(ADMIN_PALETTE)) {
    const property = name.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)
    element.style.setProperty(`--admin-${property}`, value)
  }
}
