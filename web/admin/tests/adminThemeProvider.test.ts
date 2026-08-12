import assert from 'node:assert/strict'
import { after, test } from 'node:test'
import { Window } from 'happy-dom'
import { createServer } from 'vite'

const testWindow = new Window({ url: 'https://admin.example.test/' })
const browserGlobals = {
  window: testWindow,
  document: testWindow.document,
  navigator: testWindow.navigator,
  HTMLElement: testWindow.HTMLElement,
  Node: testWindow.Node,
  SVGElement: testWindow.SVGElement,
  ShadowRoot: testWindow.ShadowRoot,
  Event: testWindow.Event,
  getComputedStyle: testWindow.getComputedStyle.bind(testWindow),
  localStorage: testWindow.localStorage,
  sessionStorage: testWindow.sessionStorage,
  IS_REACT_ACT_ENVIRONMENT: true,
}
for (const [name, value] of Object.entries(browserGlobals)) {
  Object.defineProperty(globalThis, name, { configurable: true, value })
}

async function flushTheme() {
  await Promise.resolve()
  await Promise.resolve()
}

let viteServer: Awaited<ReturnType<typeof createServer>> | undefined

async function loadAdminModules() {
  viteServer ??= await createServer({
    configFile: false,
    root: process.cwd(),
    appType: 'custom',
    server: { middlewareMode: true },
    optimizeDeps: { noDiscovery: true },
  })
  const [providerModule, themeModule, storeModule, userModule] = await Promise.all([
    viteServer.ssrLoadModule('/src/components/AdminThemeProvider.tsx'),
    viteServer.ssrLoadModule('/src/api/theme.ts'),
    viteServer.ssrLoadModule('/src/store/index.ts'),
    viteServer.ssrLoadModule('/src/store/userSlice.ts'),
  ])
  return { providerModule, themeModule, storeModule, userModule }
}

after(async () => viteServer?.close())

test('AdminThemeProvider propagates tenant palette through DOM, Ant tokens, and theme refresh', { concurrency: false }, async () => {
  const React = await import('react')
  const { act } = React
  const { createRoot } = await import('react-dom/client')
  const { Provider } = await import('react-redux')
  const { theme: antdTheme } = await import('antd')
  const { providerModule, themeModule, storeModule, userModule } = await loadAdminModules()
  const { default: AdminThemeProvider } = providerModule
  const { themeApi } = themeModule
  const { store } = storeModule
  const { clearSession, setProfile } = userModule
  const { deriveClayColors } = await import('@imaiplay/shared/theme/tenantTheme')
  const { createAdminPalette } = await import('../src/theme/adminPalette.ts')

  let primary = '#22C55E'
  let getCalls = 0
  const originalGet = themeApi.get
  themeApi.get = async () => {
    getCalls += 1
    return {
      data: {
        primary_color: primary,
        selected_background_color: '#DCFCE7',
        selected_text_color: '#14532D',
        selected_icon_color: '#166534',
        welcome_text: 'Acme',
        brand_name: 'Acme',
        browser_title: 'Acme Admin',
      },
    } as Awaited<ReturnType<typeof originalGet>>
  }

  function TokenProbe() {
    const { token } = antdTheme.useToken()
    return React.createElement('output', {
      id: 'admin-token-probe',
      'data-primary': token.colorPrimary,
      'data-link': token.colorLink,
    })
  }

  const container = document.createElement('div')
  document.body.appendChild(container)
  const root = createRoot(container)
  store.dispatch(setProfile({
    id: 'admin-1', name: 'Admin', email: 'admin@example.test', role: 'admin', tenant_id: 'tenant-acme',
  }))

  try {
    await act(async () => {
      root.render(React.createElement(
        Provider,
        { store },
        React.createElement(AdminThemeProvider, null, React.createElement(TokenProbe)),
      ))
      await flushTheme()
    })

    const initialClay = deriveClayColors(primary)
    assert.equal(getCalls, 1)
    assert.equal(document.documentElement.style.getPropertyValue('--admin-accent'), primary)
    assert.equal(document.documentElement.style.getPropertyValue('--admin-clay-shadow'), initialClay.shadow)
    assert.equal(document.querySelector('#admin-token-probe')?.getAttribute('data-primary'), primary.toLowerCase())
    assert.equal(
      document.querySelector('#admin-token-probe')?.getAttribute('data-link'),
      createAdminPalette(primary).accentForeground.toLowerCase(),
    )

    primary = '#8B5CF6'
    await act(async () => {
      window.dispatchEvent(new Event('tenant-theme-changed'))
      await flushTheme()
    })

    const refreshedClay = deriveClayColors(primary)
    assert.equal(getCalls, 2)
    assert.equal(document.documentElement.style.getPropertyValue('--admin-accent'), primary)
    assert.equal(document.documentElement.style.getPropertyValue('--admin-clay-shadow'), refreshedClay.shadow)
    assert.equal(document.querySelector('#admin-token-probe')?.getAttribute('data-primary'), primary.toLowerCase())
  } finally {
    await act(async () => root.unmount())
    container.remove()
    store.dispatch(clearSession())
    themeApi.get = originalGet
  }
})

test('AdminThemeProvider uses indigo fallback without fetching a superadmin tenant theme', { concurrency: false }, async () => {
  const React = await import('react')
  const { act } = React
  const { createRoot } = await import('react-dom/client')
  const { Provider } = await import('react-redux')
  const { theme: antdTheme } = await import('antd')
  const { providerModule, themeModule, storeModule, userModule } = await loadAdminModules()
  const { default: AdminThemeProvider } = providerModule
  const { themeApi } = themeModule
  const { store } = storeModule
  const { clearSession, setProfile } = userModule

  let getCalls = 0
  const originalGet = themeApi.get
  themeApi.get = async () => {
    getCalls += 1
    throw new Error('superadmin should not load tenant theme')
  }

  function TokenProbe() {
    const { token } = antdTheme.useToken()
    return React.createElement('output', { id: 'superadmin-token-probe', 'data-primary': token.colorPrimary })
  }

  const container = document.createElement('div')
  document.body.appendChild(container)
  const root = createRoot(container)
  store.dispatch(setProfile({ id: 'super-1', name: 'Root', email: 'root@example.test', role: 'superadmin' }))

  try {
    await act(async () => {
      root.render(React.createElement(
        Provider,
        { store },
        React.createElement(AdminThemeProvider, null, React.createElement(TokenProbe)),
      ))
      await flushTheme()
    })

    assert.equal(getCalls, 0)
    assert.equal(document.documentElement.style.getPropertyValue('--admin-accent'), '#4F46E5')
    assert.equal(document.querySelector('#superadmin-token-probe')?.getAttribute('data-primary'), '#4f46e5')
  } finally {
    await act(async () => root.unmount())
    container.remove()
    store.dispatch(clearSession())
    themeApi.get = originalGet
  }
})
