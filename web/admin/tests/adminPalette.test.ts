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

test('admin palette uses the shared indigo clay fallback and readable surfaces', async () => {
  const { ADMIN_PALETTE } = await import('../src/theme/adminPalette.ts')
  assert.equal(ADMIN_PALETTE.accent, '#4F46E5')
  assert.equal(ADMIN_PALETTE.accentHover, '#4338CA')
  assert.equal(ADMIN_PALETTE.accentLight, '#EEF2FF')
  assert.equal(ADMIN_PALETTE.accentSoft, '#E0E7FF')
  assert.equal(ADMIN_PALETTE.accentStrong, '#3730A3')
  assert.equal(ADMIN_PALETTE.heading, '#0F172A')
  assert.equal(ADMIN_PALETTE.page, '#F8FAFC')
  assert.equal(ADMIN_PALETTE.card, '#FFFFFF')
  assert.equal(ADMIN_PALETTE.success, '#10B981')
  assert.equal(ADMIN_PALETTE.warning, '#F59E0B')
  assert.equal(ADMIN_PALETTE.danger, '#EF4444')
  assert.equal(ADMIN_PALETTE.info, '#3B82F6')
  assert.ok(contrast(ADMIN_PALETTE.heading, ADMIN_PALETTE.card) >= 12)
  assert.ok(contrast(ADMIN_PALETTE.text, ADMIN_PALETTE.card) >= 7)
  assert.ok(contrast(ADMIN_PALETTE.muted, ADMIN_PALETTE.card) >= 4.5)
})

test('admin shell injects clay, semantic status, and independent selected colors', async () => {
  const paletteModule = await import('../src/theme/adminPalette.ts')
  const palette = paletteModule.createAdminPalette('#3582E1')
  const selectionColors = {
    selected_background_color: '#FFF1F0',
    selected_text_color: '#C5221F',
    selected_icon_color: '#8C1D18',
  }
  const tokens = paletteModule.createAdminThemeTokens(palette, selectionColors)

  assert.equal(tokens.primary, '#3582E1')
  assert.equal(tokens.info, palette.info)
  assert.equal(tokens.menuSelectedColor, '#C5221F')
  assert.equal(tokens.menuSelectedBackground, '#FFF1F0')
  assert.equal(tokens.menuHoverColor, '#3582E1')
  assert.equal(tokens.menuHoverBackground, '#DBE9FA')
  assert.equal(palette.accentHover, '#3075CB')
  assert.equal(palette.accentSoft, '#DBE9FA')
  assert.match(palette.claySurface, /^#[0-9A-F]{6}$/)
  assert.match(palette.clayShadow, /^#[0-9A-F]{6}$/)
  assert.notEqual(palette.clayShadow, palette.claySurface)
  assert.match(palette.clayAtmosphere, /^rgba\(/)
  assert.match(palette.clayHighlight, /^rgba\(/)

  const properties = new Map<string, string>()
  paletteModule.applyAdminPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement, palette, selectionColors)
  assert.equal(properties.get('--admin-accent'), '#3582E1')
  assert.equal(properties.get('--admin-accent-light'), palette.accentLight)
  assert.equal(properties.get('--admin-accent-strong'), palette.accentStrong)
  assert.equal(properties.get('--admin-clay-surface'), palette.claySurface)
  assert.equal(properties.get('--admin-clay-shadow'), palette.clayShadow)
  assert.equal(properties.get('--admin-clay-atmosphere'), palette.clayAtmosphere)
  assert.equal(properties.get('--admin-clay-highlight'), palette.clayHighlight)
  assert.equal(properties.get('--admin-clay-white-shadow'), palette.clayWhiteShadow)
  assert.equal(properties.get('--admin-success'), palette.success)
  assert.equal(properties.get('--admin-success-light'), palette.successLight)
  assert.equal(properties.get('--admin-warning'), palette.warning)
  assert.equal(properties.get('--admin-danger'), palette.danger)
  assert.equal(properties.get('--admin-info'), palette.info)
  assert.equal(properties.get('--admin-shadow'), palette.shadow)
  assert.equal(properties.get('--brand-600'), '#3582E1')
  assert.equal(properties.get('--tenant-primary'), '#3582E1')
  assert.equal(properties.get('--tenant-selected-background'), '#FFF1F0')
  assert.equal(properties.get('--tenant-selected-text'), '#C5221F')
  assert.equal(properties.get('--tenant-selected-icon'), '#8C1D18')
  assert.equal(properties.get('--tenant-focus'), '#3582E1')

  const lightPalette = paletteModule.createAdminPalette('#FFD43B')
  const lightTokens = paletteModule.createAdminThemeTokens(lightPalette, {
    selected_background_color: '#FFD43B',
    selected_text_color: '#000000',
    selected_icon_color: '#000000',
  })
  assert.equal(lightTokens.menuSelectedBackground, '#FFD43B')
  assert.equal(lightTokens.menuSelectedColor, '#000000')

  paletteModule.applyAdminPalette({
    style: {
      setProperty(name: string, value: string) {
        properties.set(name, value)
      },
    },
  } as unknown as HTMLElement, lightPalette, {
    selected_background_color: '#FFD43B',
    selected_text_color: '#000000',
    selected_icon_color: '#000000',
  })
  assert.equal(properties.get('--tenant-selected-text'), '#000000')
})
