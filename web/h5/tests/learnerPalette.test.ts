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

test('learner text remains readable on light H5 surfaces', async () => {
  const { LEARNER_PALETTE } = await import('../src/theme/learnerPalette.ts')
  assert.ok(contrast(LEARNER_PALETTE.heading, LEARNER_PALETTE.card) >= 12)
  assert.ok(contrast(LEARNER_PALETTE.text, LEARNER_PALETTE.card) >= 7)
  assert.ok(contrast(LEARNER_PALETTE.muted, LEARNER_PALETTE.card) >= 4.5)
  assert.ok(contrast(LEARNER_PALETTE.heading, LEARNER_PALETTE.page) >= 12)
})

test('H5 learner palette derives complete tenant semantic and Clay colors', async () => {
  const paletteModule = await import('../src/theme/learnerPalette.ts') as typeof import('../src/theme/learnerPalette.ts') & {
    createLearnerPalette?: (primaryColor?: string) => Record<string, string>
  }
  const { deriveClayColors } = await import('@imaiplay/shared/theme/tenantTheme')

  assert.equal(typeof paletteModule.createLearnerPalette, 'function')
  if (!paletteModule.createLearnerPalette) return

  const palette = paletteModule.createLearnerPalette('#22C55E')
  const clay = deriveClayColors('#22C55E')

  assert.equal(palette.accent, '#22c55e')
  assert.ok(palette.accentHover)
  assert.ok(palette.accentLight)
  assert.ok(palette.accentSoft)
  assert.ok(palette.accentStrong)
  assert.ok(palette.accentForeground)
  assert.ok(palette.accentContrastText)
  assert.ok(palette.accentHoverContrastText)
  assert.equal(palette.claySurface, clay.surface)
  assert.equal(palette.clayShadow, clay.shadow)
  assert.equal(palette.clayAtmosphere, clay.atmosphere)
  assert.equal(palette.clayHighlight, clay.highlight)
  assert.ok(palette.clayWhiteShadow)
  assert.ok(palette.success)
  assert.ok(palette.warning)
  assert.ok(palette.danger)
  assert.ok(palette.info)
  assert.ok(palette.violet)
  assert.ok(palette.rose)
  assert.ok(palette.teal)
  assert.ok(palette.glassSoft)
  assert.ok(palette.glassStrong)
  assert.ok(palette.overlayDark)
  assert.equal(paletteModule.createLearnerPalette('not-a-color').accent, '#6366f1')
})
