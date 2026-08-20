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
    accentLight: LEARNER_PALETTE.accentLight,
    accentSoft: LEARNER_PALETTE.accentSoft,
    success: LEARNER_PALETTE.success,
    warning: LEARNER_PALETTE.warning,
    heading: LEARNER_PALETTE.heading,
    text: LEARNER_PALETTE.text,
    muted: LEARNER_PALETTE.muted,
    page: LEARNER_PALETTE.page,
    card: LEARNER_PALETTE.card,
    line: LEARNER_PALETTE.line,
  }, {
    accent: '#6366f1',
    accentHover: '#4f46e5',
    accentLight: '#eef2ff',
    accentSoft: '#e0e7ff',
    success: '#10b981',
    warning: '#f59e0b',
    heading: '#0f172a',
    text: '#334155',
    muted: '#64748b',
    page: '#f8fafc',
    card: '#ffffff',
    line: '#e2e8f0',
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
  const { deriveClayColors } = await import('@imaiplay/shared/theme/tenantTheme')
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
  const clay = deriveClayColors('#22c55e')
  paletteModule.applyLearnerPalette(element, palette)
  const tokens = paletteModule.createLearnerThemeTokens('#22c55e')

  assert.equal(properties.get('--learner-accent'), '#22c55e')
  assert.equal(properties.get('--learner-accent-light'), palette.accentLight)
  assert.equal(properties.get('--learner-accent-soft'), palette.accentSoft)
  assert.equal(properties.get('--learner-success'), palette.success)
  assert.equal(properties.get('--learner-warning'), palette.warning)
  assert.equal(properties.get('--learner-player'), palette.player)
  assert.equal(properties.get('--learner-accent-foreground'), palette.accentForeground)
  assert.equal(properties.get('--learner-accent-contrast-text'), palette.accentContrastText)
  assert.equal(properties.get('--learner-accent-hover-contrast-text'), palette.accentHoverContrastText)
  assert.equal(palette.claySurface, clay.surface)
  assert.equal(palette.clayShadow, clay.shadow)
  assert.equal(palette.clayAtmosphere, clay.atmosphere)
  assert.equal(palette.clayHighlight, clay.highlight)
  assert.equal(typeof palette.clayWhiteShadow, 'string')
  assert.equal(properties.get('--learner-clay-surface'), palette.claySurface)
  assert.equal(properties.get('--learner-clay-shadow'), palette.clayShadow)
  assert.equal(properties.get('--learner-clay-atmosphere'), palette.clayAtmosphere)
  assert.equal(properties.get('--learner-clay-highlight'), palette.clayHighlight)
  assert.equal(properties.get('--learner-clay-white-shadow'), palette.clayWhiteShadow)
  assert.equal(properties.get('--learner-login-hero-start'), palette.loginHeroStart)
  assert.equal(properties.get('--learner-login-hero-end'), palette.loginHeroEnd)
  assert.equal(properties.get('--learner-login-action'), palette.loginAction)
  assert.equal(properties.get('--learner-login-action-hover'), palette.loginActionHover)
  assert.equal(properties.get('--learner-login-glow'), palette.loginGlow)
  assert.equal(tokens.colorPrimary, '#22c55e')
  assert.equal(tokens.colorInfo, '#22c55e')
  assert.equal(tokens.colorPrimaryText, palette.accentForeground)
  assert.equal(tokens.colorLink, palette.accentForeground)
  assert.equal(tokens.controlOutline, palette.accentForeground)
  assert.equal(tokens.colorTextLightSolid, palette.accentContrastText)
})

test('login colors mute vivid tenant colors while preserving readable branded states', async () => {
  const paletteModule = await import('../src/theme/learnerPalette.ts') as typeof import('../src/theme/learnerPalette.ts') & {
    createLearnerPalette?: (primaryColor?: string) => typeof paletteModule.LEARNER_PALETTE & {
      loginHeroStart: string
      loginHeroEnd: string
      loginAction: string
      loginActionHover: string
      loginGlow: string
      loginActionContrastText: string
      loginActionHoverContrastText: string
    }
  }
  assert.equal(typeof paletteModule.createLearnerPalette, 'function')
  if (!paletteModule.createLearnerPalette) return

  const vivid = paletteModule.createLearnerPalette('#d414a0')
  assert.equal(vivid.loginHeroStart, '#8c287e')
  assert.equal(vivid.loginHeroEnd, '#6b316f')
  assert.equal(vivid.loginAction, '#a7218b')
  assert.equal(vivid.loginActionHover, '#92207c')
  assert.equal(vivid.loginGlow, '#fcf1f9')
  assert.ok(contrast('#ffffff', vivid.loginHeroStart) >= 4.5)
  assert.ok(contrast('#ffffff', vivid.loginHeroEnd) >= 4.5)
  assert.ok(contrast(vivid.loginActionContrastText, vivid.loginAction) >= 4.5)
  assert.ok(contrast(vivid.loginActionHoverContrastText, vivid.loginActionHover) >= 4.5)

  for (const brightPrimary of ['#ffd43b', '#ffffff', '#22c55e']) {
    const palette = paletteModule.createLearnerPalette(brightPrimary)
    assert.ok(contrast('#ffffff', palette.loginHeroStart) >= 4.5, `${brightPrimary} start`)
    assert.ok(contrast('#ffffff', palette.loginHeroEnd) >= 4.5, `${brightPrimary} end`)
  }

  const properties = new Map<string, string>()
  paletteModule.applyLearnerPalette({
    style: { setProperty: (name: string, value: string) => properties.set(name, value) },
  } as unknown as HTMLElement, vivid, {
    selected_background_color: '#D414A0',
    selected_text_color: '#FFFFFF',
    selected_icon_color: '#FFFFFF',
  })
  assert.equal(properties.get('--tenant-selected-focus'), '#972584')
  assert.ok(contrast(properties.get('--tenant-selected-focus')!, '#ffffff') >= 3)
})

test('learner accent foreground is readable and invalid tenant colors preserve indigo fallback', async () => {
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

  assert.equal(paletteModule.createLearnerPalette('not-a-color').accent, '#6366f1')
})
