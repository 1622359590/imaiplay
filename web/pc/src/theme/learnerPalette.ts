const DEFAULT_ACCENT = '#ff5156';

export interface LearnerPalette {
  accent: string;
  accentHover: string;
  accentSoft: string;
  accentForeground: string;
  accentContrastText: string;
  accentHoverContrastText: string;
  heading: string;
  text: string;
  muted: string;
  page: string;
  card: string;
  line: string;
}

const SURFACE_PALETTE = {
  heading: '#262626',
  text: '#595959',
  muted: '#737373',
  page: '#fafafa',
  card: '#ffffff',
  line: '#eeeeee',
} as const;

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
  const accentHover = usesDefaultAccent ? '#e84349' : mix(accent, '#000000', 0.1);
  const accentSoft = usesDefaultAccent ? '#fff1f0' : mix(accent, '#ffffff', 0.9);
  return {
    accent,
    accentHover,
    accentSoft,
    accentForeground: readableForeground(accent, [
      SURFACE_PALETTE.card,
      SURFACE_PALETTE.page,
      accentSoft,
    ]),
    accentContrastText: readableSolidText(accent),
    accentHoverContrastText: readableSolidText(accentHover),
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
  ['accentSoft', '--learner-accent-soft'],
  ['accentForeground', '--learner-accent-foreground'],
  ['accentContrastText', '--learner-accent-contrast-text'],
  ['accentHoverContrastText', '--learner-accent-hover-contrast-text'],
  ['heading', '--learner-heading'],
  ['text', '--learner-text'],
  ['muted', '--learner-muted'],
  ['page', '--learner-page'],
  ['card', '--learner-card'],
  ['line', '--learner-line'],
];

export function applyLearnerPalette(
  element: HTMLElement = document.documentElement,
  palette: LearnerPalette = LEARNER_PALETTE,
) {
  for (const [name, property] of CSS_PROPERTIES) {
    element.style.setProperty(property, palette[name]);
  }
}
