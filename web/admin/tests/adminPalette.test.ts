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

test('admin shell keeps the coral interaction palette independent of tenant portal colors', async () => {
  const paletteModule = await import('../src/theme/adminPalette.ts')
  const tokens = (paletteModule as {
    ADMIN_THEME_TOKENS?: {
      primary: string
      info: string
      menuSelectedColor: string
      menuSelectedBackground: string
    }
  }).ADMIN_THEME_TOKENS

  assert.deepEqual(tokens, {
    primary: '#ff5156',
    info: '#ff5156',
    menuSelectedColor: '#ff5156',
    menuSelectedBackground: '#fff1f0',
  })

  const properties = new Map<string, string>()
  paletteModule.applyAdminPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement)
  assert.equal(properties.get('--brand-600'), '#ff5156')
  assert.equal(properties.get('--tenant-primary'), '#ff5156')
  assert.equal(properties.get('--tenant-selected'), '#fff1f0')
  assert.equal(properties.get('--tenant-focus'), '#ff5156')
})
