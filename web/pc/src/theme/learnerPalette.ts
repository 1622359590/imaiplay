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
} as const;

export function applyLearnerPalette(element: HTMLElement = document.documentElement) {
  element.style.setProperty('--learner-accent', LEARNER_PALETTE.accent);
  element.style.setProperty('--learner-accent-hover', LEARNER_PALETTE.accentHover);
  element.style.setProperty('--learner-accent-soft', LEARNER_PALETTE.accentSoft);
  element.style.setProperty('--learner-heading', LEARNER_PALETTE.heading);
  element.style.setProperty('--learner-text', LEARNER_PALETTE.text);
  element.style.setProperty('--learner-muted', LEARNER_PALETTE.muted);
  element.style.setProperty('--learner-page', LEARNER_PALETTE.page);
  element.style.setProperty('--learner-card', LEARNER_PALETTE.card);
  element.style.setProperty('--learner-line', LEARNER_PALETTE.line);
}
