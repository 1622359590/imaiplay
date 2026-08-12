import { deriveClayColors, recommendedSelectionColors } from '@imaiplay/shared/theme/tenantTheme';
import type { TenantSelectionColors } from '@imaiplay/shared/types/theme';

const DEFAULT_ACCENT = '#6366f1';

export interface LearnerPalette {
  accent: string;
  accentHover: string;
  accentLight: string;
  accentSoft: string;
  accentStrong: string;
  accentForeground: string;
  accentContrastText: string;
  accentHoverContrastText: string;
  heading: string;
  text: string;
  muted: string;
  page: string;
  card: string;
  line: string;
  success: string;
  successLight: string;
  successStrong: string;
  warning: string;
  warningLight: string;
  danger: string;
  dangerLight: string;
  info: string;
  infoLight: string;
  violet: string;
  rose: string;
  teal: string;
  player: string;
  playerBar: string;
  white: string;
  navGlass: string;
  glassSoft: string;
  glassStrong: string;
  overlayDark: string;
  shadowSm: string;
  shadow: string;
  shadowLg: string;
  claySurface: string;
  clayShadow: string;
  clayAtmosphere: string;
  clayHighlight: string;
  clayWhiteShadow: string;
}

const SURFACE_PALETTE = {
  heading: '#0f172a',
  text: '#334155',
  muted: '#64748b',
  page: '#f8fafc',
  card: '#ffffff',
  line: '#e2e8f0',
  success: '#10b981',
  successLight: '#ecfdf5',
  successStrong: '#047857',
  warning: '#f59e0b',
  warningLight: '#fffbeb',
  danger: '#ef4444',
  dangerLight: '#fef2f2',
  info: '#3b82f6',
  infoLight: '#eff6ff',
  violet: '#8b5cf6',
  rose: '#ec4899',
  teal: '#14b8a6',
  player: '#0f172a',
  playerBar: '#1e293b',
  white: '#ffffff',
  navGlass: 'rgba(255, 255, 255, 0.85)',
  glassSoft: 'rgba(255, 255, 255, 0.12)',
  glassStrong: 'rgba(255, 255, 255, 0.22)',
  overlayDark: 'rgba(15, 23, 42, 0.42)',
  shadowSm: '0 1px 2px rgba(15, 23, 42, 0.04)',
  shadow: '0 1px 3px rgba(15, 23, 42, 0.06), 0 4px 12px rgba(15, 23, 42, 0.04)',
  shadowLg: '0 8px 30px rgba(15, 23, 42, 0.08)',
} as const;

const CLAY_WHITE_SHADOW = '#d1d9e6';

function validColor(value: string | undefined): value is string {
  return Boolean(value && /^#[0-9a-f]{6}$/i.test(value));
}

function channels(hex: string): [number, number, number] {
  return [
    Number.parseInt(hex.slice(1, 3), 16),
    Number.parseInt(hex.slice(3, 5), 16),
    Number.parseInt(hex.slice(5, 7), 16),
  ];
}

function toHex(values: [number, number, number]): string {
  return `#${values.map((value) => Math.round(value).toString(16).padStart(2, '0')).join('')}`;
}

function mix(color: string, target: '#000000' | '#ffffff', amount: number): string {
  const sourceChannels = channels(color);
  const targetChannels = channels(target);
  return toHex(sourceChannels.map((value, index) => (
    value + (targetChannels[index] - value) * amount
  )) as [number, number, number]);
}

function relativeLuminance(hex: string): number {
  const [red, green, blue] = channels(hex).map((channel) => {
    const value = channel / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}

function contrastRatio(first: string, second: string): number {
  const firstLuminance = relativeLuminance(first);
  const secondLuminance = relativeLuminance(second);
  const lighter = Math.max(firstLuminance, secondLuminance);
  const darker = Math.min(firstLuminance, secondLuminance);
  return (lighter + 0.05) / (darker + 0.05);
}

function readableForeground(accent: string, surfaces: string[]): string {
  for (let percent = 0; percent <= 100; percent += 1) {
    const candidate = mix(accent, '#000000', percent / 100);
    if (surfaces.every((surface) => contrastRatio(candidate, surface) >= 4.5)) {
      return candidate;
    }
  }
  return '#000000';
}

function readableSolidText(accent: string): string {
  return contrastRatio('#000000', accent) >= contrastRatio('#ffffff', accent)
    ? '#000000'
    : '#ffffff';
}

export function createLearnerPalette(primaryColor?: string): LearnerPalette {
  const accent = validColor(primaryColor) ? primaryColor.toLowerCase() : DEFAULT_ACCENT;
  const usesDefaultAccent = accent === DEFAULT_ACCENT;
  const accentHover = usesDefaultAccent ? '#4f46e5' : mix(accent, '#000000', 0.12);
  const accentLight = usesDefaultAccent ? '#eef2ff' : mix(accent, '#ffffff', 0.92);
  const accentSoft = usesDefaultAccent ? '#e0e7ff' : mix(accent, '#ffffff', 0.82);
  const accentStrong = usesDefaultAccent ? '#4338ca' : mix(accent, '#000000', 0.22);
  const clay = deriveClayColors(accent);
  return {
    accent,
    accentHover,
    accentLight,
    accentSoft,
    accentStrong,
    accentForeground: readableForeground(accent, [
      SURFACE_PALETTE.card,
      SURFACE_PALETTE.page,
      accentLight,
      accentSoft,
    ]),
    accentContrastText: readableSolidText(accent),
    accentHoverContrastText: readableSolidText(accentHover),
    claySurface: clay.surface,
    clayShadow: clay.shadow,
    clayAtmosphere: clay.atmosphere,
    clayHighlight: clay.highlight,
    clayWhiteShadow: CLAY_WHITE_SHADOW,
    ...SURFACE_PALETTE,
  };
}

export const LEARNER_PALETTE = createLearnerPalette();

export function createLearnerThemeTokens(primaryColor?: string) {
  const palette = createLearnerPalette(primaryColor);
  return {
    colorPrimary: palette.accent,
    colorInfo: palette.accent,
    colorPrimaryText: palette.accentForeground,
    colorPrimaryTextHover: palette.accentForeground,
    colorPrimaryTextActive: palette.accentForeground,
    colorLink: palette.accentForeground,
    colorLinkHover: palette.accentForeground,
    colorLinkActive: palette.accentForeground,
    controlOutline: palette.accentForeground,
    colorTextLightSolid: palette.accentContrastText,
    colorText: palette.text,
    colorTextHeading: palette.heading,
    colorBgLayout: palette.page,
    colorBgContainer: palette.card,
    colorBorderSecondary: palette.line,
    borderRadius: 10,
  };
}

const CSS_PROPERTIES: Array<[keyof LearnerPalette, string]> = [
  ['accent', '--learner-accent'],
  ['accentHover', '--learner-accent-hover'],
  ['accentLight', '--learner-accent-light'],
  ['accentSoft', '--learner-accent-soft'],
  ['accentStrong', '--learner-accent-strong'],
  ['accentForeground', '--learner-accent-foreground'],
  ['accentContrastText', '--learner-accent-contrast-text'],
  ['accentHoverContrastText', '--learner-accent-hover-contrast-text'],
  ['heading', '--learner-heading'],
  ['text', '--learner-text'],
  ['muted', '--learner-muted'],
  ['page', '--learner-page'],
  ['card', '--learner-card'],
  ['line', '--learner-line'],
  ['success', '--learner-success'],
  ['successLight', '--learner-success-light'],
  ['successStrong', '--learner-success-strong'],
  ['warning', '--learner-warning'],
  ['warningLight', '--learner-warning-light'],
  ['danger', '--learner-danger'],
  ['dangerLight', '--learner-danger-light'],
  ['info', '--learner-info'],
  ['infoLight', '--learner-info-light'],
  ['violet', '--learner-violet'],
  ['rose', '--learner-rose'],
  ['teal', '--learner-teal'],
  ['player', '--learner-player'],
  ['playerBar', '--learner-player-bar'],
  ['white', '--learner-white'],
  ['navGlass', '--learner-nav-glass'],
  ['glassSoft', '--learner-glass-soft'],
  ['glassStrong', '--learner-glass-strong'],
  ['overlayDark', '--learner-overlay-dark'],
  ['shadowSm', '--learner-shadow-sm'],
  ['shadow', '--learner-shadow'],
  ['shadowLg', '--learner-shadow-lg'],
  ['claySurface', '--learner-clay-surface'],
  ['clayShadow', '--learner-clay-shadow'],
  ['clayAtmosphere', '--learner-clay-atmosphere'],
  ['clayHighlight', '--learner-clay-highlight'],
  ['clayWhiteShadow', '--learner-clay-white-shadow'],
];

export function applyLearnerPalette(
  element: HTMLElement = document.documentElement,
  palette: LearnerPalette = LEARNER_PALETTE,
  selectionColors: TenantSelectionColors = recommendedSelectionColors(palette.accent),
) {
  for (const [name, property] of CSS_PROPERTIES) {
    element.style.setProperty(property, palette[name]);
  }
  element.style.setProperty('--tenant-selected-background', selectionColors.selected_background_color);
  element.style.setProperty('--tenant-selected-text', selectionColors.selected_text_color);
  element.style.setProperty('--tenant-selected-icon', selectionColors.selected_icon_color);
}
