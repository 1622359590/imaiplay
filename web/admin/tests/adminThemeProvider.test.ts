import assert from 'node:assert/strict'
import { after, test } from 'node:test'
import { createServer } from 'vite'

let viteServer: Awaited<ReturnType<typeof createServer>> | undefined

async function loadThemeValueModule() {
  viteServer ??= await createServer({
    configFile: false,
    root: process.cwd(),
    appType: 'custom',
    server: { middlewareMode: true },
    optimizeDeps: { noDiscovery: true },
  })
  return viteServer.ssrLoadModule('/src/theme/adminThemeValue.ts') as Promise<
    typeof import('../src/theme/adminThemeValue.ts')
  >
}

after(async () => viteServer?.close())

test('tenant theme refresh resolves a new context value without retaining the previous primary', async () => {
  const { resolveAdminThemeValue } = await loadThemeValueModule()
  const initial = resolveAdminThemeValue({
    primary_color: '#22C55E',
    selected_background_color: '#DCFCE7',
    selected_text_color: '#14532D',
    selected_icon_color: '#166534',
    welcome_text: 'Acme 学习平台',
    brand_name: 'Acme',
    browser_title: 'Acme Admin',
  }, 'Acme Tenant')
  const refreshed = resolveAdminThemeValue({
    primary_color: '#8B5CF6',
    selected_background_color: '#EDE9FE',
    selected_text_color: '#4C1D95',
    selected_icon_color: '#5B21B6',
    welcome_text: 'Acme 学习平台',
    brand_name: 'Acme',
    browser_title: 'Acme Admin',
  }, 'Acme Tenant')

  assert.equal(initial.primaryColor, '#22C55E')
  assert.equal(refreshed.primaryColor, '#8B5CF6')
  assert.notDeepEqual(refreshed, initial)
  assert.equal(refreshed.selectedBackgroundColor, '#EDE9FE')
  assert.equal(refreshed.selectedTextColor, '#4C1D95')
  assert.equal(refreshed.selectedIconColor, '#5B21B6')
  assert.equal(refreshed.brandName, 'Acme 学习平台')
  assert.equal(refreshed.browserTitle, 'Acme Admin')
})

test('admin theme value normalizes invalid tenant colors and retains the indigo fallback', async () => {
  const { FALLBACK_ADMIN_THEME, resolveAdminThemeValue } = await loadThemeValueModule()
  const resolved = resolveAdminThemeValue({ primary_color: 'not-a-color' })

  assert.deepEqual(resolved, FALLBACK_ADMIN_THEME)
  assert.equal(FALLBACK_ADMIN_THEME.primaryColor, '#4F46E5')
})
