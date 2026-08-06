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

test('admin shell applies the tenant brand color to selected and primary surfaces', async () => {
  const paletteModule = await import('../src/theme/adminPalette.ts')
  const palette = paletteModule.createAdminPalette('#3582E1')
  const tokens = paletteModule.createAdminThemeTokens(palette)

  assert.deepEqual(tokens, {
    primary: '#3582E1',
    info: '#3582E1',
    menuSelectedColor: '#ffffff',
    menuSelectedBackground: '#3582E1',
    menuHoverColor: '#3582E1',
    menuHoverBackground: '#EBF3FC',
  })
  assert.equal(palette.accentHover, '#3075CB')
  assert.equal(palette.accentSoft, '#EBF3FC')

  const properties = new Map<string, string>()
  paletteModule.applyAdminPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement, palette)
  assert.equal(properties.get('--admin-accent'), '#3582E1')
  assert.equal(properties.get('--brand-600'), '#3582E1')
  assert.equal(properties.get('--tenant-primary'), '#3582E1')
  assert.equal(properties.get('--tenant-selected'), '#3582E1')
  assert.equal(properties.get('--tenant-selected-text'), '#ffffff')
  assert.equal(properties.get('--tenant-focus'), '#3582E1')

  const lightPalette = paletteModule.createAdminPalette('#FFD43B')
  const lightTokens = paletteModule.createAdminThemeTokens(lightPalette)
  assert.equal(lightTokens.menuSelectedBackground, '#FFD43B')
  assert.equal(lightTokens.menuSelectedColor, '#000000')

  paletteModule.applyAdminPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement, lightPalette)
  assert.equal(properties.get('--tenant-selected-text'), '#000000')
})
