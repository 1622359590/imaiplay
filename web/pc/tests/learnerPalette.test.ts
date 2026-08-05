import assert from 'node:assert/strict'
import test from 'node:test'

function luminance(hex: string): number {
  const channels = hex.slice(1).match(/.{2}/g)
  assert.ok(channels, `invalid color: ${hex}`)
  const [red, green, blue] = channels.map((channel) => {
    const value = Number.parseInt(channel, 16) / 255
    return value <= 0.04045
      ? value / 12.92
      : ((value + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue
}

function contrast(foreground: string, background: string): number {
  const lighter = Math.max(luminance(foreground), luminance(background))
  const darker = Math.min(luminance(foreground), luminance(background))
  return (lighter + 0.05) / (darker + 0.05)
}

test('learner text remains readable on light PC surfaces', async () => {
  const { LEARNER_PALETTE } = await import('../src/theme/learnerPalette.ts')
  assert.deepEqual({
    accent: LEARNER_PALETTE.accent,
    accentHover: LEARNER_PALETTE.accentHover,
    accentSoft: LEARNER_PALETTE.accentSoft,
    heading: LEARNER_PALETTE.heading,
    text: LEARNER_PALETTE.text,
    muted: LEARNER_PALETTE.muted,
    page: LEARNER_PALETTE.page,
    card: LEARNER_PALETTE.card,
    line: LEARNER_PALETTE.line,
  }, {
    accent: '#ff5156',
    accentHover: '#e84349',
    accentSoft: '#fff1f0',
    heading: '#262626',
    text: '#595959',
    muted: '#737373',
    page: '#fafafa',
    card: '#ffffff',
    line: '#eeeeee',
  })
  assert.ok(contrast(LEARNER_PALETTE.heading, LEARNER_PALETTE.card) >= 12)
  assert.ok(contrast(LEARNER_PALETTE.text, LEARNER_PALETTE.card) >= 7)
  assert.ok(contrast(LEARNER_PALETTE.muted, LEARNER_PALETTE.card) >= 4.5)
  assert.ok(contrast(LEARNER_PALETTE.heading, LEARNER_PALETTE.page) >= 12)
})

test('tenant primary propagates to learner CSS and Ant Design tokens', async () => {
  const paletteModule = await import('../src/theme/learnerPalette.ts') as typeof import('../src/theme/learnerPalette.ts') & {
    createLearnerPalette?: (primaryColor?: string) => typeof paletteModule.LEARNER_PALETTE & {
      accentForeground: string
      accentContrastText: string
      accentHoverContrastText: string
    }
    createLearnerThemeTokens?: (primaryColor?: string) => {
      colorPrimary: string
      colorInfo: string
      colorPrimaryText: string
      colorLink: string
      controlOutline: string
      colorTextLightSolid: string
    }
  }
  const properties = new Map<string, string>()
  const element = {
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement

  assert.equal(typeof paletteModule.createLearnerPalette, 'function')
  assert.equal(typeof paletteModule.createLearnerThemeTokens, 'function')
  if (!paletteModule.createLearnerPalette || !paletteModule.createLearnerThemeTokens) return

  const palette = paletteModule.createLearnerPalette('#22c55e')
  paletteModule.applyLearnerPalette(element, palette)
  const tokens = paletteModule.createLearnerThemeTokens('#22c55e')

  assert.equal(properties.get('--learner-accent'), '#22c55e')
  assert.equal(properties.get('--learner-accent-foreground'), palette.accentForeground)
  assert.equal(properties.get('--learner-accent-contrast-text'), palette.accentContrastText)
  assert.equal(properties.get('--learner-accent-hover-contrast-text'), palette.accentHoverContrastText)
  assert.equal(tokens.colorPrimary, '#22c55e')
  assert.equal(tokens.colorInfo, '#22c55e')
  assert.equal(tokens.colorPrimaryText, palette.accentForeground)
  assert.equal(tokens.colorLink, palette.accentForeground)
  assert.equal(tokens.controlOutline, palette.accentForeground)
  assert.equal(tokens.colorTextLightSolid, palette.accentContrastText)
})

test('learner accent foreground is readable and invalid tenant colors preserve coral fallback', async () => {
  const paletteModule = await import('../src/theme/learnerPalette.ts') as typeof import('../src/theme/learnerPalette.ts') & {
    createLearnerPalette?: (primaryColor?: string) => typeof paletteModule.LEARNER_PALETTE & { accentForeground: string }
  }

  assert.equal(typeof paletteModule.createLearnerPalette, 'function')
  if (!paletteModule.createLearnerPalette) return

  for (const primaryColor of ['#ff5156', '#777777', '#22c55e', '#ffffff']) {
    const palette = paletteModule.createLearnerPalette(primaryColor)
    assert.equal(typeof palette.accentContrastText, 'string')
    assert.ok(
      contrast(palette.accentForeground, '#ffffff') >= 4.5,
      `${primaryColor} derived ${palette.accentForeground} below 4.5:1`,
    )
    assert.ok(
      contrast(palette.accentForeground, palette.accentSoft) >= 4.5,
      `${primaryColor} derived ${palette.accentForeground} below 4.5:1 on ${palette.accentSoft}`,
    )
    assert.ok(
      contrast(palette.accentContrastText, palette.accent) >= 4.5,
      `${primaryColor} solid text ${palette.accentContrastText} below 4.5:1`,
    )
  }

  const lightTokens = paletteModule.createLearnerThemeTokens?.('#ffffff')
  assert.ok(lightTokens)
  assert.ok(contrast(lightTokens.colorPrimaryText, '#ffffff') >= 4.5)
  assert.ok(contrast(lightTokens.colorLink, '#ffffff') >= 4.5)
  assert.ok(contrast(lightTokens.controlOutline, '#ffffff') >= 4.5)

  const thresholdPalette = paletteModule.createLearnerPalette('#777777')
  assert.equal(typeof thresholdPalette.accentHoverContrastText, 'string')
  assert.ok(
    contrast(thresholdPalette.accentHoverContrastText, thresholdPalette.accentHover) >= 4.5,
    `#777777 hover text ${thresholdPalette.accentHoverContrastText} below 4.5:1 on ${thresholdPalette.accentHover}`,
  )

  assert.equal(paletteModule.createLearnerPalette('not-a-color').accent, '#ff5156')
})
