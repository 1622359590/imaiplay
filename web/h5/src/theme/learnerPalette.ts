import {
  normalizePrimaryColor,
  recommendedSelectionColors,
} from '@imaiplay/shared/theme/tenantTheme'
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme'

export const LEARNER_PALETTE = {
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

export function applyLearnerPalette(
  element: HTMLElement = document.documentElement,
  primaryColor: string = LEARNER_PALETTE.accent,
  selectionColors: TenantSelectionColors = recommendedSelectionColors(primaryColor),
) {
  const accent = normalizePrimaryColor(primaryColor, LEARNER_PALETTE.accent)
  const usesDefaultAccent = accent.toLowerCase() === LEARNER_PALETTE.accent
  element.style.setProperty('--learner-accent', accent)
  element.style.setProperty('--learner-accent-hover', usesDefaultAccent
    ? LEARNER_PALETTE.accentHover
    : mix(accent, '#000000', 0.1))
  element.style.setProperty('--learner-accent-soft', usesDefaultAccent
    ? LEARNER_PALETTE.accentSoft
    : mix(accent, '#FFFFFF', 0.9))
  element.style.setProperty('--learner-heading', LEARNER_PALETTE.heading)
  element.style.setProperty('--learner-text', LEARNER_PALETTE.text)
  element.style.setProperty('--learner-muted', LEARNER_PALETTE.muted)
  element.style.setProperty('--learner-page', LEARNER_PALETTE.page)
  element.style.setProperty('--learner-card', LEARNER_PALETTE.card)
  element.style.setProperty('--learner-line', LEARNER_PALETTE.line)
  element.style.setProperty('--tenant-selected-background', selectionColors.selected_background_color)
  element.style.setProperty('--tenant-selected-text', selectionColors.selected_text_color)
  element.style.setProperty('--tenant-selected-icon', selectionColors.selected_icon_color)
  element.style.setProperty('--adm-color-primary', accent)
}

function mix(color: string, target: '#000000' | '#FFFFFF', amount: number): string {
  const channels = [1, 3, 5].map((offset) => Number.parseInt(color.slice(offset, offset + 2), 16))
  const targetChannels = [1, 3, 5].map((offset) => Number.parseInt(target.slice(offset, offset + 2), 16))
  return `#${channels.map((value, index) => Math.round(
    value + (targetChannels[index] - value) * amount,
  ).toString(16).padStart(2, '0')).join('')}`.toUpperCase()
}
