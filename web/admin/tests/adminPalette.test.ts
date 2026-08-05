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

test('admin palette uses the approved readable coral surfaces', async () => {
  const { ADMIN_PALETTE } = await import('../src/theme/adminPalette.ts')
  assert.deepEqual(ADMIN_PALETTE, {
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
  assert.ok(contrast(ADMIN_PALETTE.heading, ADMIN_PALETTE.card) >= 12)
  assert.ok(contrast(ADMIN_PALETTE.text, ADMIN_PALETTE.card) >= 7)
  assert.ok(contrast(ADMIN_PALETTE.muted, ADMIN_PALETTE.card) >= 4.5)
})
